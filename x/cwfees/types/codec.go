package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(ir codectypes.InterfaceRegistry) {
	msgservice.RegisterMsgServiceDesc(ir, &_Msg_serviceDesc)
}
