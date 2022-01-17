package types

import (
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (b BlockGasTracking) GetGasConsumed() sdk.Dec {
	var gasConsumed uint64 = 0
	for _, txTrackingInfo := range b.TxTrackingInfos {
		for _, contractTrackingInfo := range txTrackingInfo.ContractTrackingInfos {
			gasConsumed += contractTrackingInfo.GasConsumed
		}
	}

	return sdk.NewDecFromBigInt(big.NewInt(0).SetUint64(gasConsumed))
}
