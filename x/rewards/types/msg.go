package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	TypeMsgSetContractMetadata = "set-contract-metadata"
	TypeMsgWithdrawRewards     = "withdraw-rewards"
	TypeMsgFlatFee             = "flat-fee"
	TypeMsgUpdateParams        = "update-params"
)

var (
	_ sdk.Msg = &MsgSetContractMetadata{}
	_ sdk.Msg = &MsgWithdrawRewards{}
	_ sdk.Msg = &MsgSetFlatFee{}
	_ sdk.Msg = &MsgUpdateParams{}
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
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}

	if err := m.Metadata.Validate(false); err != nil {
		return err
	}

	return nil
}

// NewMsgWithdrawRewardsByLimit creates a new MsgWithdrawRewards instance using the records limit oneof option.
func NewMsgWithdrawRewardsByLimit(senderAddr sdk.AccAddress, recordsLimit uint64) *MsgWithdrawRewards {
	return &MsgWithdrawRewards{
		RewardsAddress: senderAddr.String(),
		Mode: &MsgWithdrawRewards_RecordsLimit_{
			RecordsLimit: &MsgWithdrawRewards_RecordsLimit{
				Limit: recordsLimit,
			},
		},
	}
}

// NewMsgWithdrawRewardsByIDs creates a new MsgWithdrawRewards instance using the record IDs oneof option.
func NewMsgWithdrawRewardsByIDs(senderAddr sdk.AccAddress, recordIDs []uint64) *MsgWithdrawRewards {
	return &MsgWithdrawRewards{
		RewardsAddress: senderAddr.String(),
		Mode: &MsgWithdrawRewards_RecordIds{
			RecordIds: &MsgWithdrawRewards_RecordIDs{
				Ids: recordIDs,
			},
		},
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
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid rewards address: %v", err)
	}

	if m.Mode == nil {
		return errorsmod.Wrap(sdkErrors.ErrInvalidRequest, "invalid mode: nil")
	}

	switch modeReq := m.Mode.(type) {
	case *MsgWithdrawRewards_RecordsLimit_:
		if modeReq == nil {
			return errorsmod.Wrap(sdkErrors.ErrInvalidRequest, "invalid records limit: nil mode object")
		}
		if modeReq.RecordsLimit == nil {
			return errorsmod.Wrap(sdkErrors.ErrInvalidRequest, "invalid records limit: nil request")
		}
	case *MsgWithdrawRewards_RecordIds:
		if modeReq == nil {
			return errorsmod.Wrap(sdkErrors.ErrInvalidRequest, "invalid record IDs: nil mode object")
		}
		if modeReq.RecordIds == nil {
			return errorsmod.Wrap(sdkErrors.ErrInvalidRequest, "invalid record IDs: nil request")
		}

		if len(modeReq.RecordIds.Ids) == 0 {
			return errorsmod.Wrap(sdkErrors.ErrInvalidRequest, "invalid record IDs: empty")
		}

		idsSet := make(map[uint64]struct{})
		for _, id := range m.GetRecordIds().Ids {
			if id == 0 {
				return errorsmod.Wrap(sdkErrors.ErrInvalidRequest, "invalid record IDs: must be GT 0")
			}

			if _, ok := idsSet[id]; ok {
				return errorsmod.Wrapf(sdkErrors.ErrInvalidRequest, "invalid record IDs: duplicate ID (%d)", id)
			}
			idsSet[id] = struct{}{}
		}
	default:
		return errorsmod.Wrapf(sdkErrors.ErrUnknownRequest, "unknown withdraw rewards mode: %T", m.Mode)
	}

	return nil
}

// NewMsgFlatFee creates a new MsgSetFlatFee instance.
func NewMsgFlatFee(senderAddr, contractAddr sdk.AccAddress, flatFee sdk.Coin) *MsgSetFlatFee {
	msg := &MsgSetFlatFee{
		SenderAddress:   senderAddr.String(),
		ContractAddress: contractAddr.String(),
		FlatFeeAmount:   flatFee,
	}

	return msg
}

// GetSigners implements the sdk.Msg interface.
func (m MsgSetFlatFee) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(m.SenderAddress)
	if err != nil {
		panic(fmt.Errorf("parsing sender address (%s): %w", m.SenderAddress, err))
	}

	return []sdk.AccAddress{senderAddr}
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgSetFlatFee) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.SenderAddress); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}
	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid contract address: %v", err)
	}

	return nil
}

// NewMsgUpdateParams creates a new MsgUpdateParams instance.
func NewMsgUpdateParams(senderAddr sdk.AccAddress, params Params) *MsgUpdateParams {
	msg := &MsgUpdateParams{
		Authority: senderAddr.String(),
		Params:    params,
	}

	return msg
}

// Route implements the sdk.Msg interface.
func (m MsgUpdateParams) Route() string { return RouterKey }

// Type implements the sdk.Msg interface.
func (m MsgUpdateParams) Type() string { return TypeMsgFlatFee }

// GetSigners implements the sdk.Msg interface.
func (m MsgUpdateParams) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		panic(fmt.Errorf("parsing sender address (%s): %w", m.Authority, err))
	}

	return []sdk.AccAddress{senderAddr}
}

// GetSignBytes implements the sdk.Msg interface.
func (m MsgUpdateParams) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(&m)
	return sdk.MustSortJSON(bz)
}

// ValidateBasic implements the sdk.Msg interface.
func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errorsmod.Wrapf(sdkErrors.ErrInvalidAddress, "invalid sender address: %v", err)
	}

	return nil
}
