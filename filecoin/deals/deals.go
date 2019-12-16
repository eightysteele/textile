package deals

import (
	"context"

	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"
)

// Dealer manages deals for a particular Address
type Dealer struct {
	api  api.FullNode
	addr address.Address
}

func New(api api.FullNode, addr address.Address) *Dealer {
	return &Dealer{
		addr: addr,
		api:  api,
	}
}

func (d *Dealer) Create(ctx context.Context, data cid.Cid, miner address.Address,
	epochPrice types.BigInt, duration uint64) (*cid.Cid, error) {
	d.api.ClientStartDeal(ctx, data, d.addr, miner, epochPrice, duration)
}
