package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/filecoin-project/lotus/api/client"
	logging "github.com/ipfs/go-log"
	"github.com/textileio/textile/filecoin/reputation"
)

var (
	daemonAddr = "127.0.0.1:1234"
	// generate with: `lotus auth create-token --perm admin`
	authToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJBbGxvdyI6WyJyZWFkIiwid3JpdGUiLCJzaWduIiwiYWRtaW4iXX0.ABFsb_wXxww9XxqD3nM5KEuXekBtNsWIKpyoDZR0Qwc"
)

func main() {
	logging.SetLogLevel("*", "fatal")
	headers := http.Header{
		"Authorization": []string{"Bearer " + authToken},
	}
	client, close, err := client.NewFullNodeRPC("ws://"+daemonAddr+"/rpc/v0", headers)
	if err != nil {
		panic(err)
	}
	defer close()

	repManager := reputation.New(client)

	go func() {
		for {
			select {
			case <-time.After(time.Second * 3):
				print("\033[H\033[2J")
				reps := repManager.Reputations()
				for _, r := range reps {
					fmt.Printf("Addr: %s, PeerID %s, Slashed: %v\n", r.Address, r.PeerID, r.GotSlashed)
				}
			}
		}
	}()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGTERM, os.Interrupt)
	<-s
	repManager.Close()
}
