package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgRegisterContract{}
	_ sdk.Msg = &MsgRegisterCode{}
)
