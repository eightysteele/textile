module github.com/textileio/textile

go 1.13

require (
	github.com/filecoin-project/filecoin-ffi v0.0.0-20191211095301-e32f5efc808b // indirect
	github.com/filecoin-project/lotus v0.1.0
	github.com/ipfs/go-log v1.0.0
	github.com/libp2p/go-libp2p-core v0.2.4
	github.com/libp2p/go-libp2p-peer v0.2.0
	github.com/whyrusleeping/cbor-gen v0.0.0-20191209162422-1c55bd7cf8aa
	go.opencensus.io v0.22.2 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543
)

replace github.com/filecoin-project/filecoin-ffi => ./filecoin/extern/filecoin-ffi
