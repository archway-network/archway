package types

import (
	"fmt"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/wasmbinding/pkg"
)

// ContracFlatFeeRequest is the Query.FlatFee request.
type ContractFlatFeeRequest struct {
	// ContractAddress is the bech32 encoded contract address.
	ContractAddress string `json:"contract_address"`
}

// ContractFlatFeeResponse is the Query.Metadata response.
type ContractFlatFeeResponse struct {
	// The amount which has been set as the contract flat fee
	FlatFeeAmount wasmVmTypes.Coin `json:"flat_fee_amount"`
}

// Validate performs request fields validation.
func (r ContractFlatFeeRequest) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.ContractAddress); err != nil {
		return fmt.Errorf("contractAddress: parsing: %w", err)
	}

	return nil
}

// MustGetContractAddress returns the contract address as sdk.AccAddress.
// CONTRACT: panics in case of an error (should not happen since we validate the request).
func (r ContractFlatFeeRequest) MustGetContractAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(r.ContractAddress)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: meta request: parsing contractAddress: %w", err))
	}

	return addr
}

// NewContractFlatFeeResponse converts rewardsTypes.FlatFee to ContractFlatFeeResponse.
func NewContractFlatFeeResponse(flatFee sdk.Coin) ContractFlatFeeResponse {
	return ContractFlatFeeResponse{
		FlatFeeAmount: pkg.SDKCoinToWasm(flatFee),
	}
}
