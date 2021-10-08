module github.com/archway-network/archway

go 1.16

require (
	github.com/CosmWasm/wasmd v0.17.0
	github.com/cosmos/cosmos-sdk v0.42.5
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.11.0
	github.com/rakyll/statik v0.1.7
	github.com/rs/zerolog v1.23.0
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.8.0 // indirect
	github.com/tendermint/tendermint v0.34.10
	github.com/tendermint/tm-db v0.6.4
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/CosmWasm/wasmd => ./wasmd