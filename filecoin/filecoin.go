package filecoin

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
)

const (
	defaultWalletType = "bls"
)

type Service struct {
	api api.FullNode
}

type MinerAsk struct {
	Name              address.Address
	FILPerGiBPerBlock types.BigInt
}

type Miner struct {
	Name address.Address
}

type Deal struct {
	Cid cid.Cid
}

type Slug struct {
	Cid       cid.Cid
	Name      string
	CreatedAt time.Time
	Deals     []Deal
}

func New(api api.FullNode) *Service {
	return &Service{
		api: api,
	}
}
func (s *Service) Import(ctx context.Context, addr address.Address,
	name string, data io.Reader) (cid.Cid, error) {
	tmpF, err := ioutil.TempFile("", "import-*")
	if err != nil {
		return cid.Undef, fmt.Errorf("error when creating tmpfile: %s", err)
	}
	defer tmpF.Close()
	defer os.Remove(tmpF.Name())
	if _, err := io.Copy(tmpF, data); err != nil {
		return cid.Undef, fmt.Errorf("error when copying data to tmpfile: %s", err)
	}

	// ToDo: propose a PR or something to remove? not sure when will be necessary
	// but eventually we would need to prune whats on the node?

	return s.api.ClientImport(ctx, tmpF.Name())
}

func (s *Service) GetAsks(ctx context.Context, cid cid.Cid, pricePerEpoch types.FIL,
	durationEpoch, minProposals int) ([]MinerAsk, error) {
	panic("TODO")
}

func (s *Service) Put(cid cid.Cid, miners []address.Address) (Slug, error) {
	panic("TODO")
}

func (s *Service) CreateAddr(ctx context.Context) (address.Address, error) {
	a, err := s.api.WalletNew(ctx, defaultWalletType)
	if err != nil {
		return address.Undef, fmt.Errorf("error when generating wallet address: %s", err)
	}
	return a, nil
}
