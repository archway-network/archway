package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ sdk.Msg = &MsgSetContractMetadata{}

func (m MsgSetContractMetadata) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Admin); err != nil {
		return err
	}

	if _, err := sdk.AccAddressFromBech32(m.ContractAddress); err != nil {
		return err
	}

	if m.Metadata == nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract metadata cannot be set to nil")
	}

	if m.Metadata.CollectPremium && m.Metadata.GasRebateToUser {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract metadata cannot have both collect premium and "+
			"gas rebate turned on")
	}

	if m.Metadata.CollectPremium {
		if m.Metadata.PremiumPercentageCharged > 200 {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "contract metadata cannot have premium percentage greater "+
				"than 200")
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
	adminAddr, err := sdk.AccAddressFromBech32(m.Admin)
	if err != nil {
		panic(err.Error())
	}
	return []sdk.AccAddress{adminAddr}
}
