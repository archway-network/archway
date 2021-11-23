package gastracker

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule = AppModule{}
)

type AppModule struct {
	bankKeeper bankkeeper.Keeper
	keeper     GasTrackingKeeper
	mintKeeper mintkeeper.Keeper
}

func NewAppModule(keeper GasTrackingKeeper, bk bankkeeper.Keeper, mk mintkeeper.Keeper) AppModule {
	return AppModule{keeper: keeper, bankKeeper: bk, mintKeeper: mk}
}

func (a AppModule) Name() string {
	return ModuleName
}

func (a AppModule) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {

}

func (a AppModule) RegisterInterfaces(registry codectypes.InterfaceRegistry) {

}

func (a AppModule) DefaultGenesis(marshaler codec.JSONMarshaler) json.RawMessage {
	return json.RawMessage{}
}

func (a AppModule) ValidateGenesis(marshaler codec.JSONMarshaler, config client.TxEncodingConfig, message json.RawMessage) error {
	return nil
}

func (a AppModule) RegisterRESTRoutes(context client.Context, router *mux.Router) {

}

func (a AppModule) RegisterGRPCGatewayRoutes(context client.Context, mux *runtime.ServeMux) {

}

func (a AppModule) GetTxCmd() *cobra.Command {
	return nil
}

func (a AppModule) GetQueryCmd() *cobra.Command {
	return nil
}

func (a AppModule) InitGenesis(context sdk.Context, marshaler codec.JSONMarshaler, message json.RawMessage) []abci.ValidatorUpdate {
	InitParams(context, a.keeper)

	return []abci.ValidatorUpdate{}
}

func (a AppModule) ExportGenesis(context sdk.Context, marshaler codec.JSONMarshaler) json.RawMessage {
	return json.RawMessage{}
}

func (a AppModule) RegisterInvariants(registry sdk.InvariantRegistry) {

}

func (a AppModule) Route() sdk.Route {
	return sdk.NewRoute("dummy", func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		return nil, nil
	})
}

func (a AppModule) QuerierRoute() string {
	return "dummy"
}

func (a AppModule) LegacyQuerierHandler(amino *codec.LegacyAmino) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		return []byte{}, nil
	}
}

func (a AppModule) RegisterServices(configurator module.Configurator) {

}

func (a AppModule) BeginBlock(context sdk.Context, block abci.RequestBeginBlock) {
	BeginBlock(context, block, a.keeper, a.bankKeeper, a.mintKeeper)
}

func (a AppModule) EndBlock(context sdk.Context, block abci.RequestEndBlock) []abci.ValidatorUpdate {
	return EndBlock(context, a.keeper, block)
}
