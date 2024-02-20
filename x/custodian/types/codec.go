package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterCodec registers concrete types on the Amino codec
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgRegisterInterchainAccount{}, "/archway.custodian.v1.MsgRegisterInterchainAccount", nil)
	cdc.RegisterConcrete(&MsgSubmitTx{}, "/archway.custodian.v1.MsgSubmitTx", nil)
	cdc.RegisterConcrete(&MsgUpdateParams{}, "/archway.custodian.v1.MsgUpdateParams", nil)
}

// RegisterInterfaces registers the x/custodian interfaces types with the interface registry
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRegisterInterchainAccount{},
		&MsgSubmitTx{},
		&MsgUpdateParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(legacy.Cdc)
}
