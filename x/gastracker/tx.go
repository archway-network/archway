package gastracker

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSetContractMetadata{}

func (m MsgSetContractMetadata) ValidateBasic() error {
	if len(m.ContractAddress) == 0 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract address is invalid, error: %s", err.Error())
	}

	if m.Metadata == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "metadata to be set cannot be nil")
	}

	if len(m.Metadata.RewardAddress) != 0 {
		if _, err := sdk.AccAddressFromBech32(m.Metadata.RewardAddress); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "metadata's reward address is invalid, error: %s", err.Error())
		}
	}

	if len(m.Metadata.DeveloperAddress) != 0 {
		if _, err := sdk.AccAddressFromBech32(m.Metadata.DeveloperAddress); err != nil {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "metadata's developer address is invalid, error: %s", err.Error())
		}
	}

	if m.Metadata.GasRebateToUser && m.Metadata.CollectPremium {
		return ErrInvalidInitRequest1
	}

	if m.Metadata.CollectPremium {
		if m.Metadata.PremiumPercentageCharged > 200 {
			return ErrInvalidInitRequest2
		}
	}

	return nil
}

func (m MsgSetContractMetadata) Route() string {
	return RouterKey
}

func (m MsgSetContractMetadata) Type() string {
	return "set-contract-metadata"
}

func (m MsgSetContractMetadata) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

func (m MsgSetContractMetadata) GetSigners() []sdk.AccAddress {
	senderAddr, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		panic(err.Error())
	}
	return []sdk.AccAddress{senderAddr}
}
