package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgSubscribeToError{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// GetSigners implements the sdk.Msg interface.
func (m MsgSubscribeToError) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Sender)}
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgSubscribeToError) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid contract address: %v", err)
	}
	if err := m.Fee.Validate(); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidCoins, "invalid fee: %v", err)
	}
	return nil
}

// GetSigners implements the sdk.Msg interface.
func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{sdk.MustAccAddressFromBech32(m.Authority)}
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid authority address: %v", err)
	}

	return m.Params.Validate()
}
