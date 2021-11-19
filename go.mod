module github.com/archway-network/archway

go 1.16

require (
	github.com/CosmWasm/wasmd v0.18.1-0.20210901073821-4e242e082c59
	github.com/CosmWasm/wasmvm v0.16.1
	github.com/cosmos/cosmos-sdk v0.42.9
	github.com/gogo/protobuf v1.3.3
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/prometheus/client_golang v1.11.0
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1 // indirect
	github.com/rs/zerolog v1.25.0
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.11
	github.com/tendermint/tm-db v0.6.4
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1
