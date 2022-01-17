package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/gastracker"
)

func (b BlockGasTracking) GetGasConsumed() sdk.Dec {
	var gasConsumed uint64 = 0
	for _, txTrackingInfo := range b.TxTrackingInfos {
		for _, contractTrackingInfo := range txTrackingInfo.ContractTrackingInfos {
			gasConsumed += contractTrackingInfo.GasConsumed
		}
	}

	return sdk.NewDecFromBigInt(gastracker.ConvertUint64ToBigInt(gasConsumed))
}
