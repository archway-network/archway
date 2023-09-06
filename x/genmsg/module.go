package genmsg

import (
	"encoding/json"
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/archway-network/archway/x/genmsg/types"
)

const (
	ModuleName = "genmsg"
)

var _ module.AppModule = (*AppModule)(nil)

// MessageRouter ADR 031 request type routing
type MessageRouter interface {
	Handler(msg sdk.Msg) baseapp.MsgServiceHandler
}

func NewAppModule(h MessageRouter) AppModule {
	return AppModule{h}
}

type AppModule struct {
	router MessageRouter
}

func (a AppModule) Name() string { return ModuleName }

func (AppModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(new(types.GenesisState))
}

func (AppModule) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	state := new(types.GenesisState)
	if err := cdc.UnmarshalJSON(bz, state); err != nil {
		return fmt.Errorf("failed to unmarshal x/%s genesis state: %w", ModuleName, err)
	}
	return validateGenesis(cdc, state)
}

func (a AppModule) InitGenesis(context sdk.Context, codec codec.JSONCodec, message json.RawMessage) []abci.ValidatorUpdate {
	state := new(types.GenesisState)
	codec.MustUnmarshalJSON(message, state)
	err := initGenesis(context, codec, a.router, state)
	if err != nil {
		panic(err)
	}
	return nil
}

func (a AppModule) ExportGenesis(_ sdk.Context, codec codec.JSONCodec) json.RawMessage {
	return codec.MustMarshalJSON(new(types.GenesisState))
}

func (a AppModule) RegisterRESTRoutes(_ client.Context, _ *mux.Router) {}

func (a AppModule) RegisterGRPCGatewayRoutes(_ client.Context, _ *runtime.ServeMux) {}

func (a AppModule) GetTxCmd() *cobra.Command { return &cobra.Command{Use: ModuleName} }

func (a AppModule) GetQueryCmd() *cobra.Command { return &cobra.Command{Use: ModuleName} }

func (a AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (a AppModule) RegisterServices(_ module.Configurator) {}

func (a AppModule) ConsensusVersion() uint64 { return 0 }

func (a AppModule) RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}

func (a AppModule) RegisterInterfaces(_ codectypes.InterfaceRegistry) {}
