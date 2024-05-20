package callback

import (
	"context"
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codecTypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/x/callback/client/cli"
	"github.com/archway-network/archway/x/callback/keeper"
	"github.com/archway-network/archway/x/callback/types"
)

var (
	_ module.AppModuleBasic  = AppModuleBasic{}
	_ module.AppModule       = AppModule{}
	_ module.HasABCIEndBlock = AppModule{}
)

// AppModuleBasic defines the basic application module for this module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the module's name.
func (a AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the module's types on the given LegacyAmino codec.
func (a AppModuleBasic) RegisterLegacyAminoCodec(amino *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(amino)
}

// RegisterInterfaces registers the module's interface types.
func (a AppModuleBasic) RegisterInterfaces(registry codecTypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the module.
func (a AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the module.
func (a AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var state types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &state); err != nil {
		return fmt.Errorf("failed to unmarshal x/%s genesis state: %w", types.ModuleName, err)
	}

	return state.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the module.
func (a AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, serveMux *runtime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), serveMux, types.NewQueryClient(clientCtx)); err != nil {
		panic(fmt.Errorf("registering query handler for x/%s: %w", types.ModuleName, err))
	}
}

// GetTxCmd returns the root tx command for the module.
func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}

// GetQueryCmd returns no root query command for the module.
func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return cli.GetQueryCmd()
}

// AppModule implements an application module for this module.
type AppModule struct {
	AppModuleBasic

	keeper       keeper.Keeper
	wasmKeeper   types.WasmKeeperExpected
	errorsKeeper types.ErrorsKeeperExpected
}

// NewAppModule creates a new AppModule object.
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, wk types.WasmKeeperExpected, ek types.ErrorsKeeperExpected) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		wasmKeeper:     wk,
		errorsKeeper:   ek,
	}
}

// RegisterInvariants registers the module invariants.
func (a AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
}

// RegisterServices registers the module services.
func (a AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(a.keeper))
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(a.keeper))
}

// InitGenesis performs genesis initialization for the module. It returns no validator updates.
func (a AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(bz, &genesisState)

	InitGenesis(ctx, a.keeper, genesisState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the module.
func (a AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state := ExportGenesis(ctx, a.keeper)
	return cdc.MustMarshalJSON(state)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (a AppModule) ConsensusVersion() uint64 {
	return 1
}

// EndBlock returns the end blocker for the module. It returns no validator updates.
func (a AppModule) EndBlock(ctx context.Context) ([]abci.ValidatorUpdate, error) {
	return EndBlocker(sdk.UnwrapSDKContext(ctx), a.keeper, a.wasmKeeper, a.errorsKeeper)
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}
