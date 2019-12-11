package reputation

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/actors"
	"github.com/filecoin-project/lotus/chain/address"
	"github.com/filecoin-project/lotus/chain/types"
	logging "github.com/ipfs/go-log"
	"github.com/libp2p/go-libp2p-core/peer"
	cbg "github.com/whyrusleeping/cbor-gen"
)

var log = logging.Logger("reputation")

const (
	apiTimeout             = time.Second * 5
	maxConcurrCalculations = 1
)

type Manager struct {
	api api.FullNode

	lock        sync.Mutex
	reputations []Reputation

	stopped bool
	close   chan struct{}
	closed  chan struct{}
}

func New(api api.FullNode) *Manager {
	m := &Manager{
		api:    api,
		close:  make(chan struct{}),
		closed: make(chan struct{}),
	}
	go m.refreshState()
	return m
}

type Reputation struct {
	Address    address.Address
	PeerID     peer.ID
	GotSlashed bool
	AvgLatency time.Duration
	Country    string
}

func (m *Manager) Reputations() []Reputation {
	m.lock.Lock()
	defer m.lock.Unlock()
	ret := make([]Reputation, len(m.reputations))
	copy(ret, m.reputations)
	return ret
}

func (m *Manager) Close() {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.stopped {
		return
	}
	close(m.close)
	<-m.closed
}

func (m *Manager) refreshState() {
	defer close(m.closed)
	for {
		select {
		case <-m.close:
			return
		case <-time.After(time.Second * 10):
			ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
			addrs, err := m.api.StateListMiners(ctx, nil)
			if err != nil {
				//log.Errorf("getting miner list failed: %s", err)
				cancel()
				continue
			}

			rateLim := make(chan struct{}, maxConcurrCalculations)
			var wg sync.WaitGroup
			out := make(chan Reputation)
			for _, a := range addrs {
				wg.Add(1)
				go func(ctx context.Context, a address.Address) {
					defer wg.Done()
					rateLim <- struct{}{}
					r, err := m.calculate(ctx, a)
					<-rateLim
					if err != nil {
						log.Errorf("reputation calculation of %s failed: %w", a, err)
						return
					}
					out <- r
				}(ctx, a)
			}
			go func() {
				wg.Wait()
				close(out)
			}()
			var newReputations []Reputation
			for r := range out {
				newReputations = append(newReputations, r)
			}

			m.lock.Lock()
			m.reputations = newReputations
			m.lock.Unlock()
			cancel()
		}
	}
}

func (m *Manager) calculate(ctx context.Context, a address.Address) (Reputation, error) {
	peerID, err := m.api.StateMinerPeerID(ctx, a, nil)
	if err != nil {
		return Reputation{}, fmt.Errorf("getting miner peerID failed: %s %v", peerID.Pretty(), err)
	}

	res, err := m.api.StateCall(ctx, &types.Message{
		To:     a,
		From:   a,
		Method: actors.MAMethods.IsSlashed,
	}, nil)
	if err != nil {
		return Reputation{}, fmt.Errorf("statecall to IsSlashed failed with: %v", err)
	}
	if res.ExitCode != 0 {
		return Reputation{}, fmt.Errorf("call to IsSlashed failed with exitcode: %d", res.ExitCode)
	}

	return Reputation{
		Address:    a,
		PeerID:     peerID,
		GotSlashed: bytes.Equal(res.Return, cbg.CborBoolTrue),
	}, nil
}
