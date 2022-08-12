package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgSetContractMetadata = "set-contract-metadata"
	TypeMsgWithdrawRewards     = "withdraw-rewards"
)

var (
	_ sdk.Msg = &MsgSetContractMetadata{}
	_ sdk.Msg = &MsgWithdrawRewards{}
)

// NewMsgSetContractMetadata creates a new MsgSetContractMetadata instance.
func NewMsgSetContractMetadata(senderAddr, contractAddr sdk.AccAddress, ownerAddr, rewardsAddr *sdk.AccAddress) *MsgSetContractMetadata {
	msg := &MsgSetContractMetadata{
		SenderAddress: senderAddr.String(),
		Metadata: ContractMetadata{
			ContractAddress: contractAddr.String(),
		},
	}

	if ownerAddr != nil {
		msg.Metadata.OwnerAddress = ownerAddr.String()
	}
	if rewardsAddr != nil {
		msg.Metadata.RewardsAddress = rewardsAddr.String()
	}

	return msg
}

// Route implements the sdk.Msg interface.
func (m MsgSetContractMetadata) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (m MsgSetContractMetadata) Type() string { return TypeMsgSetContractMetadata }

// GetSigners implements the sdk.Msg interface.
func (m MsgSetContractMetadata) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(m.SenderAddress)
	if err != nil {
		panic(fmt.Errorf("parsing sender address (%s): %w", m.SenderAddress, err))
	}

	return []sdk.AccAddress{senderAddr}
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgSetContractMetadata) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgSetContractMetadata) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.SenderAddress); err != nil {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "invalid sender address")
	}

	if err := m.Metadata.Validate(false); err != nil {
		return err
	}

	return nil
}

// NewMsgWithdrawRewards creates a new MsgWithdrawRewards instance.
func NewMsgWithdrawRewards(senderAddr sdk.AccAddress) *MsgWithdrawRewards {
	return &MsgWithdrawRewards{
		RewardsAddress: senderAddr.String(),
	}
}

// Route implements the sdk.Msg interface.
func (m MsgWithdrawRewards) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (m MsgWithdrawRewards) Type() string { return TypeMsgWithdrawRewards }

// GetSigners implements the sdk.Msg interface.
func (m MsgWithdrawRewards) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(m.RewardsAddress)
	if err != nil {
		panic(fmt.Errorf("parsing rewards address (%s): %w", m.RewardsAddress, err))
	}

	return []sdk.AccAddress{senderAddr}
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgWithdrawRewards) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgWithdrawRewards) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.RewardsAddress); err != nil {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "invalid rewards address")
	}

	return nil
}
