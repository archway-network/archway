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
func NewMsgSetContractMetadata(senderAddr, contractAddr, ownerAddr, rewardsAddr sdk.AccAddress) *MsgSetContractMetadata {
	return &MsgSetContractMetadata{
		SenderAddress:   senderAddr.String(),
		ContractAddress: contractAddr.String(),
		Metadata: ContractMetadata{
			OwnerAddress:   ownerAddr.String(),
			RewardsAddress: rewardsAddr.String(),
		},
	}
}

// Route implements the sdk.Msg interface.
func (m MsgSetContractMetadata) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (m MsgSetContractMetadata) Type() string { return TypeMsgSetContractMetadata }

// GetSigners implements the sdk.Msg interface.
func (m MsgSetContractMetadata) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(m.SenderAddress)
	if err != nil {
		panic(fmt.Errorf("parsing sender address: %w", err))
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

	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "invalid contract address")
	}

	if err := m.Metadata.Validate(); err != nil {
		return err
	}

	return nil
}
