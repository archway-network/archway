package types

import (
	"fmt"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/wasmbinding/pkg"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// SetFlatFeeRequest is the Msg.SetFlatFee request.
type SetFlatFeeRequest struct {
	// ContractAddress is the contract for which flatfee needs to be set.
	ContractAddress string `json:"contract_address"`
	// RewardsAddress if not empty, changes the rewards distribution destination address.
	FlatFeeAmount wasmVmTypes.Coin `json:"flat_fee_amount"`
}

// Validate performs request fields validation.
func (r SetFlatFeeRequest) Validate() error {
	if _, err := sdk.AccAddressFromBech32(r.ContractAddress); err != nil {
		return fmt.Errorf("contractAddress: parsing: %w", err)
	}

	coin, err := pkg.WasmCoinToSDK(r.FlatFeeAmount)
	if err != nil {
		return fmt.Errorf("flatFeeAmount: parsing: %w", err)
	}

	err = coin.Validate()
	if err != nil {
		return fmt.Errorf("flatFeeAmount: validation error : %w", err)
	}

	return nil
}

// MustGetSdkCoinOk returns the contract address as sdk.AccAddress if set to be updated.
// CONTRACT: panics in case of an error.
func (r SetFlatFeeRequest) MustGetSdkCoinOk() sdk.Coin {
	coin, err := pkg.WasmCoinToSDK(r.FlatFeeAmount)
	if err != nil {
		// Should not happen since we validate the request before this call
		panic(fmt.Errorf("wasm bindings: flat fee update: parsing flat fee amount: %w", err))
	}
	return coin
}

// ToSDK convert the SetFlatFeeRequest to a rewardsTypes.SetFlatFee.
func (r SetFlatFeeRequest) ToSDK() rewardsTypes.FlatFee {
	fee := r.MustGetSdkCoinOk()
	return rewardsTypes.FlatFee{
		ContractAddress: r.ContractAddress,
		FlatFee:         fee,
	}
}
