package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var _ sdk.Msg = (*MsgRegisterAsGranter)(nil)

func (m *MsgRegisterAsGranter) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.GrantingContract)}
}

func (m *MsgRegisterAsGranter) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(m.GrantingContract)
	if err != nil {
		return err
	}
	return nil
}
