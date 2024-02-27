package types

import (
	"strings"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
)

const Delimiter = "."

// type ICAOwner struct {
// 	contractAddress     sdk.AccAddress
// 	interchainAccountID string
// }

// func (i ICAOwner) String() string {
// 	return i.contractAddress.String() + Delimiter + i.interchainAccountID
// }

// func NewICAOwner(contractAddressBech32, interchainAccountID string) (ICAOwner, error) {
// 	sdkContractAddress, err := sdk.AccAddressFromBech32(contractAddressBech32)
// 	if err != nil {
// 		return ICAOwner{}, errors.Wrapf(ErrInvalidAccountAddress, "failed to decode address from bech32: %v", err)
// 	}

// 	return ICAOwner{contractAddress: sdkContractAddress, interchainAccountID: interchainAccountID}, nil
// }

// func NewICAOwnerFromAddress(address sdk.AccAddress, interchainAccountID string) ICAOwner {
// 	return ICAOwner{contractAddress: address, interchainAccountID: interchainAccountID}
// }

func ICAOwnerFromPort(port string) string {
	return strings.TrimPrefix(port, icatypes.ControllerPortPrefix)
}

// func (i ICAOwner) GetContract() sdk.AccAddress {
// 	return i.contractAddress
// }

// func (i ICAOwner) GetInterchainAccountID() string {
// 	return i.interchainAccountID
// }
