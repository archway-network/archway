module github.com/archway-network/archway

go 1.16

require (
	github.com/CosmWasm/wasmd v0.25.0
	github.com/CosmWasm/wasmvm v1.0.0-beta10
	github.com/cosmos/cosmos-sdk v0.45.1
	github.com/cosmos/ibc-go/v2 v2.2.0
	github.com/gin-gonic/gin v1.7.0 // indirect
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/prometheus/client_golang v1.12.1
	github.com/rakyll/statik v0.1.7
	github.com/snikch/goodman v0.0.0-20171125024755-10e37e294daa
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.3.0
	github.com/stretchr/testify v1.7.1
	github.com/tendermint/tendermint v0.34.16
	github.com/tendermint/tm-db v0.6.7
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	google.golang.org/genproto v0.0.0-20220317150908-0efb43f6373e
	google.golang.org/grpc v1.45.0
	gopkg.in/yaml.v2 v2.4.0
)

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/CosmWasm/wasmd => github.com/archway-network/wasmd v0.25.0-archway
