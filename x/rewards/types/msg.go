package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgSetContractMetadata = "set-contract-metadata"
)

var (
	_ sdk.Msg = &MsgSetContractMetadata{}
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
