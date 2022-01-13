module github.com/archway-network/archway

go 1.16

require (
	github.com/CosmWasm/wasmd v1.0.0-beta5
	github.com/CosmWasm/wasmvm v1.0.0-beta5
	github.com/cosmos/cosmos-sdk v0.42.11
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/prometheus/client_golang v1.11.0
	github.com/rakyll/statik v0.1.7
	github.com/rs/zerolog v1.26.0
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.14
	github.com/tendermint/tm-db v0.6.4
	google.golang.org/genproto v0.0.0-20210828152312-66f60bf46e71
	google.golang.org/grpc v1.42.0
	gopkg.in/yaml.v2 v2.4.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/CosmWasm/wasmd => ./wasmd
