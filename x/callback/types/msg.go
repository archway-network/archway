package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgRequestCallback{}
	_ sdk.Msg = &MsgCancelCallback{}
	_ sdk.Msg = &MsgUpdateParams{}
)

// NewMsgRequestCallback creates a new MsgRequestCallback instance.
func NewMsgRequestCallback(
	senderAddr sdk.AccAddress,
	contractAddr sdk.AccAddress,
	jobId uint64,
	callbackHeight int64,
	fees sdk.Coin,
) *MsgRequestCallback {
	msg := &MsgRequestCallback{
		Sender:          senderAddr.String(),
		ContractAddress: contractAddr.String(),
		JobId:           jobId,
		CallbackHeight:  callbackHeight,
		Fees:            fees,
	}

	return msg
}

// GetSigners implements the sdk.Msg interface.
func (m MsgRequestCallback) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(fmt.Errorf("parsing sender address (%s): %w", m.Sender, err))
	}

	return []sdk.AccAddress{senderAddr}
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgRequestCallback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid contract address: %v", err)
	}
	if m.Fees.Denom != sdk.DefaultBondDenom {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidCoins, "invalid fees denom: %v", m.Fees.Denom)
	}

	return nil
}

// NewMsgCancelCallback creates a new MsgCancelCallback instance.
func NewMsgCancelCallback(
	senderAddr sdk.AccAddress,
	contractAddr sdk.AccAddress,
	jobId uint64,
	callbackHeight int64,
) *MsgCancelCallback {
	msg := &MsgCancelCallback{
		Sender:          senderAddr.String(),
		ContractAddress: contractAddr.String(),
		JobId:           jobId,
		CallbackHeight:  callbackHeight,
	}

	return msg
}

// GetSigners implements the sdk.Msg interface.
func (m MsgCancelCallback) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(fmt.Errorf("parsing sender address (%s): %w", m.Sender, err))
	}

	return []sdk.AccAddress{senderAddr}
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgCancelCallback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid contract address: %v", err)
	}

	return nil
}

// GetSigners implements the sdk.Msg interface.
func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(fmt.Errorf("parsing authority address (%s): %w", m.Authority, err))
	}

	return []sdk.AccAddress{senderAddr}
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid authority address: %v", err)
	}

	return m.Params.Validate()
}
