package cwfees

import (
	"encoding/json"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/x/cwfees/types"
)

const (
	ModuleName       = types.ModuleName
	ConsensusVersion = 1
)

var _ module.AppModule = (*AppModule)(nil)

func NewAppModule(k Keeper) AppModule { return AppModule{k} }

type AppModule struct {
	k Keeper
}

func (a AppModule) Name() string { return ModuleName }

func (a AppModule) ConsensusVersion() uint64 { return ConsensusVersion }

func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(new(types.GenesisState))
}

func (AppModule) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	state := new(types.GenesisState)
	err := cdc.UnmarshalJSON(bz, state)
	if err != nil {
		return err
	}
	return state.Validate()
}

func (a AppModule) InitGenesis(context sdk.Context, codec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	state := new(types.GenesisState)
	codec.MustUnmarshalJSON(message, state)
	err := a.k.ImportState(context, state)
	if err != nil {
		panic(err)
	}
	return nil
}

func (a AppModule) ExportGenesis(ctx sdk.Context, codec codec.JSONCodec) json.RawMessage {
	state, err := a.k.ExportState(ctx)
	if err != nil {
		panic(err)
	}
	return codec.MustMarshalJSON(state)
}

func (a AppModule) RegisterInterfaces(ir codectypes.InterfaceRegistry) { types.RegisterInterfaces(ir) }

func (a AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterQueryServer(cfg.QueryServer(), NewQueryServer(a.k))
	types.RegisterMsgServer(cfg.MsgServer(), NewMsgServer(a.k))
}

func (a AppModule) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

func (a AppModule) GetTxCmd() *cobra.Command { return &cobra.Command{Use: ModuleName} }

func (a AppModule) GetQueryCmd() *cobra.Command { return &cobra.Command{Use: ModuleName} }

func (a AppModule) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}
