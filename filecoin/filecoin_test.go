package filecoin

import (
	"bytes"
	"context"
	"crypto/rand"
	"testing"

	"github.com/filecoin-project/lotus/chain/address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/textileio/textile/filecoin/client"
)

func TestCompleteCycle(t *testing.T) {
	daemonAddr := "127.0.0.1:1234"
	authToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.xfTwBMejr2nRslPcUn5vD3qfdZZJ7N1wZaWWmqVG3PE"
	c, cls, err := client.New(daemonAddr, authToken)
	if err != nil {
		panic("couldn't create the client")
	}
	defer cls()

	s := New(c)
	ctx := context.Background()

	// Create fresh wallet
	addr, err := s.CreateAddr(ctx)
	checkErr(t, err)

	// Make some random data
	randData := make([]byte, 512)
	_, err = rand.Read(randData)
	checkErr(t, err)

	// Import some data
	cid, err := s.Import(ctx, addr, "name", bytes.NewReader(randData))
	checkErr(t, err)

	// Get some offers for our data
	price, err := types.ParseFIL("0.0000000001")
	checkErr(t, err)
	duration := 1000 // Epochs
	minProposals := 5
	asks, err := s.GetAsks(ctx, cid, price, duration, minProposals)
	checkErr(t, err)
	if len(asks) < minProposals {
		t.Errorf("expected at least %d proposals", minProposals)
	}

	// Save the data
	miners := []address.Address{asks[0].Address, asks[1].Address}
	slug, err := s.Put(cid, miners)
	checkErr(t, err)
	_ = slug
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
