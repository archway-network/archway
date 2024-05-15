// Package rewards defines a module that tracks dApp rewards (fee rebate and inflation rewards).
// Rewards distribution to contract's rewardsAddress (if set) is done via the Withdrawal operation triggered by a Tx or by a contract (WASM bindings).
// Collected rewards tracking data is pruned by the x/rewards module's EndBlocker.
// CONTRACT: module's Ante handler must be called after the x/tracking Ante since it relies on transaction ID tracking generates.
package rewards

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
	simTypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/x/rewards/client/cli"
	"github.com/archway-network/archway/x/rewards/keeper"
	"github.com/archway-network/archway/x/rewards/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.AppModule      = AppModule{}
)

// ConsensusVersion defines the current x/rewards module consensus version.
const ConsensusVersion = 2

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
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
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

func (a AppModuleBasic) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

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

	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
	}
}

// RegisterInvariants registers the module invariants.
func (a AppModule) RegisterInvariants(ir sdk.InvariantRegistry) {
	keeper.RegisterInvariants(ir, a.keeper)
}

// RegisterServices registers the module services.
func (a AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), keeper.NewQueryServer(a.keeper))
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServer(a.keeper))

	m := keeper.NewMigrator(a.keeper)
	if err := cfg.RegisterMigration(types.ModuleName, 1, m.Migrate1to2); err != nil {
		panic(fmt.Sprintf("failed to migrate x/%s from version 1 to 2: %v", types.ModuleName, err))
	}
}

// InitGenesis performs genesis initialization for the module. It returns no validator updates.
func (a AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, bz json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(bz, &genesisState)

	a.keeper.InitGenesis(ctx, &genesisState)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis returns the exported genesis state as raw bytes for the module.
func (a AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state := a.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(state)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (a AppModule) ConsensusVersion() uint64 {
	return ConsensusVersion
}

// BeginBlock returns the begin blocker for the module.
func (a AppModule) BeginBlock(ctx sdk.Context) {}

// EndBlock returns the end blocker for the module. It returns no validator updates.
func (a AppModule) EndBlock(ctx sdk.Context) []abci.ValidatorUpdate {
	return EndBlocker(ctx, a.keeper)
}

// AppModuleSimulation functions

// GenerateGenesisState creates a randomized GenState of the module.
func (a AppModule) GenerateGenesisState(input *module.SimulationState) {}

// // RegisterStoreDecoder registers a decoder for the module's types.
// func (a AppModule) RegisterStoreDecoder(_ storetypes.StoreDecoderRegistry) {
// }

// WeightedOperations returns all the module operations with their respective weights.
func (a AppModule) WeightedOperations(_ module.SimulationState) []simTypes.WeightedOperation {
	return []simTypes.WeightedOperation{}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}
