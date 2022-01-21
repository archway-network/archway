package gastracker

import (
	"context"
	"encoding/json"
	"github.com/archway-network/archway/x/gastracker/client/cli"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"math/rand"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.AppModule      = AppModule{}
)

type AppModuleBasic struct {
}

func (a AppModuleBasic) Name() string {
	return ModuleName
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	gstTypes.RegisterLegacyAminoCodec(amino)
}

func (a AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	gstTypes.RegisterInterfaces(registry)
}

func (a AppModuleBasic) DefaultGenesis(marshaler codec.JSONCodec) json.RawMessage {
	return marshaler.MustMarshalJSON(&gstTypes.GenesisState{})
}

func (a AppModuleBasic) ValidateGenesis(marshaler codec.JSONCodec, config client.TxEncodingConfig, message json.RawMessage) error {
	return nil
}

func (a AppModuleBasic) RegisterRESTRoutes(context client.Context, router *mux.Router) {

}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, serveMux *runtime.ServeMux) {
	gstTypes.RegisterQueryHandlerClient(context.Background(), serveMux, gstTypes.NewQueryClient(clientCtx))
}

func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

type AppModule struct {
	AppModuleBasic

	bankKeeper bankkeeper.Keeper
	keeper     GasTrackingKeeper
	mintKeeper mintkeeper.Keeper
}

func (a AppModule) GenerateGenesisState(input *module.SimulationState) {
	//TODO implement me
	panic("implement me")
}

func (a AppModule) ProposalContents(simState module.SimulationState) []simtypes.WeightedProposalContent {
	//TODO implement me
	panic("implement me")
}

func (a AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	//TODO implement me
	panic("implement me")
}

func (a AppModule) RegisterStoreDecoder(registry sdk.StoreDecoderRegistry) {
	//TODO implement me
	panic("implement me")
}

func (a AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	//TODO implement me
	panic("implement me")
}

func (a AppModule) ConsensusVersion() uint64 {
	return 1
}

func NewAppModule(keeper GasTrackingKeeper, bk bankkeeper.Keeper, mk mintkeeper.Keeper) AppModule {
	return AppModule{keeper: keeper, bankKeeper: bk, mintKeeper: mk}
}

func (a AppModule) InitGenesis(context sdk.Context, marshaler codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	InitParams(context, a.keeper)

	return []abci.ValidatorUpdate{}
}

func (a AppModule) ExportGenesis(context sdk.Context, marshaler codec.JSONCodec) json.RawMessage {
	return marshaler.MustMarshalJSON(&gstTypes.GenesisState{})
}

func (a AppModule) RegisterInvariants(registry sdk.InvariantRegistry) {

}

func (a AppModule) Route() sdk.Route {
	return sdk.NewRoute(RouterKey, NewHandler(a.keeper))
}

func (a AppModule) QuerierRoute() string {
	return QuerierRoute
}

func (a AppModule) LegacyQuerierHandler(amino *codec.LegacyAmino) sdk.Querier {
	return NewLegacyQuerier(a.keeper)
}

func (a AppModule) RegisterServices(cfg module.Configurator) {
	gstTypes.RegisterMsgServer(cfg.MsgServer(), NewMsgServer(a.keeper))
	gstTypes.RegisterQueryServer(cfg.QueryServer(), NewGRPCQuerier(a.keeper))
}

func (a AppModule) BeginBlock(context sdk.Context, block abci.RequestBeginBlock) {
	BeginBlock(context, block, a.keeper, a.bankKeeper, a.mintKeeper)
}

func (a AppModule) EndBlock(context sdk.Context, block abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
