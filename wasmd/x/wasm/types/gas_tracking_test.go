package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"
	"testing"
)

func TestQueryGasTracking(t *testing.T) {
	memDB := db.NewMemDB()
	cms := store.NewCommitMultiStore(memDB)
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, nil)
	mainMeter := ctx.GasMeter()

	initialGasMeter := NewContractGasMeter(sdk.NewGasMeter(30000000), func(_ uint64, info GasConsumptionInfo) GasConsumptionInfo {
		return GasConsumptionInfo{
			SDKGas: info.SDKGas * 2,
		}
	}, "1contract")

	err := InitializeGasTracking(&ctx, &initialGasMeter)
	require.NoError(t, err, "There should not be any error")

	fmt.Println("After init", mainMeter.GasConsumed())

	err = CreateNewSession(ctx, 100)
	require.NoError(t, err, "There should not be any error")

	fmt.Println("After creating", mainMeter.GasConsumed())

	ctx.GasMeter().ConsumeGas(400, "In middle of stuff 1")

	err = DestroySession(&ctx)
	require.NoError(t, err, "There should not be any error")

	fmt.Println("After destroying", mainMeter.GasConsumed())

	err = CreateNewSession(ctx, 5000)
	require.NoError(t, err, "There should not be an error")

	ctx.GasMeter().ConsumeGas(50, "Consuming 1")

	fmt.Println(mainMeter.GasConsumed())

	err = AssociateMeterWithCurrentSession(&ctx, func(gasLimit uint64) *ContractSDKGasMeter {
		gasMeter := NewContractGasMeter(sdk.NewGasMeter(gasLimit), func(_ uint64, info GasConsumptionInfo) GasConsumptionInfo {
			return GasConsumptionInfo{
				SDKGas: info.SDKGas / 3,
			}
		}, "2contract")
		return &gasMeter
	})
	require.NoError(t, err, "There should not be an error")

	ctx.GasMeter().ConsumeGas(100, "Consuming 2")

	fmt.Println(mainMeter.GasConsumed())

	err = AddVMRecord(ctx, &VMRecord{
		OriginalVMGas: 50,
		ActualVMGas:   500,
	})
	require.NoError(t, err, "We should be able to add vm record")

	err = DestroySession(&ctx)
	require.NoError(t, err, "There should not be any error")

	fmt.Println(mainMeter.GasConsumed(), "After destroying stuff")

	err = CreateNewSession(ctx, 100)
	require.NoError(t, err, "There should not be any error")

	fmt.Println("After creating", mainMeter.GasConsumed())

	ctx.GasMeter().ConsumeGas(400, "In middle of stuff")

	err = DestroySession(&ctx)
	require.NoError(t, err, "There should not be any error")

	fmt.Println("After destroying", mainMeter.GasConsumed())

	queryGasRecords, sessionRecord, err := TerminateGasTracking(&ctx)
	require.NoError(t, err, "We should be able to terminate")

	fmt.Println("After terminating", mainMeter.GasConsumed())

	fmt.Println(queryGasRecords[0], sessionRecord)
}
