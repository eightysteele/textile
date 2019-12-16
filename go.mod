module github.com/textileio/textile

go 1.13

require (
	github.com/filecoin-project/filecoin-ffi v0.0.0-20191211095301-e32f5efc808b // indirect
	github.com/filecoin-project/lotus v0.1.2-0.20191216120345-c25f61656205
	github.com/ipfs/go-cid v0.0.4
	github.com/ipfs/go-log v1.0.0
	github.com/libp2p/go-libp2p-core v0.2.4
	github.com/textileio/go-textile-core v0.0.0-20191205233641-31fc120682c9
	github.com/whyrusleeping/cbor-gen v0.0.0-20191212224538-d370462a7e8a
	go.opencensus.io v0.22.2 // indirect
	golang.org/x/xerrors v0.0.0-20191204190536-9bdfabe68543 // indirect
)

replace github.com/filecoin-project/filecoin-ffi => ../lotus/extern/filecoin-ffi
