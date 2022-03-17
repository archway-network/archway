package types

import (
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"
	"testing"
)

func newContractGasMeterByRef(underlyingMeter sdk.GasMeter, gasCalculationFn func(_ uint64, info GasConsumptionInfo) GasConsumptionInfo, contractAddress string) *ContractSDKGasMeter {
	gasMeter := NewContractGasMeter(underlyingMeter, gasCalculationFn, contractAddress)
	return &gasMeter
}

func TestGasTracking(t *testing.T) {
	memDB := db.NewMemDB()
	cms := store.NewCommitMultiStore(memDB)
	ctx := sdk.NewContext(cms, tmproto.Header{}, false, nil)
	mainMeter := ctx.GasMeter()

	contracts := []string{"1contract", "2contract", "3contract"}

	initialGasMeter := NewContractGasMeter(sdk.NewGasMeter(30000000), func(_ uint64, info GasConsumptionInfo) GasConsumptionInfo {
		return GasConsumptionInfo{
			SDKGas: info.SDKGas * 2,
		}
	}, contracts[0])

	// { Initialization
	err := InitializeGasTracking(&ctx, &initialGasMeter)
	require.NoError(t, err, "There should not be any error")

	require.Equal(t, uint64(0), mainMeter.GasConsumed(), "there should not be any consumption on main meter")

	// {{ Session 1 is created
	err = CreateNewSession(ctx, 100)
	require.NoError(t, err, "There should not be any error")

	require.Equal(t, uint64(0), mainMeter.GasConsumed(), "there should not be any consumption on main meter")

	ctx.GasMeter().ConsumeGas(400, "")

	// {{} Session 1 is destroyed
	err = DestroySession(&ctx)
	require.NoError(t, err, "There should not be any error")

	// {{}{ Session 2 session is created
	err = CreateNewSession(ctx, 5000)
	require.NoError(t, err, "There should not be an error")

	ctx.GasMeter().ConsumeGas(50, "")

	// {{}{ Session 2: Meter associated
	err = AssociateMeterWithCurrentSession(&ctx, newContractGasMeterByRef(sdk.NewGasMeter(5000), func(_ uint64, info GasConsumptionInfo) GasConsumptionInfo {
		return GasConsumptionInfo{
			SDKGas: info.SDKGas / 3,
		}
	}, contracts[1]))
	require.NoError(t, err, "There should not be an error")

	ctx.GasMeter().ConsumeGas(100, "")

	// {{}{{ Session 3 created
	err = CreateNewSession(ctx, 5000)
	require.NoError(t, err, "There should not be an error")

	// {{}{{ Session 3: Meter associated
	err = AssociateMeterWithCurrentSession(&ctx, newContractGasMeterByRef(sdk.NewGasMeter(5000), func(_ uint64, info GasConsumptionInfo) GasConsumptionInfo {
		return GasConsumptionInfo{
			SDKGas: info.SDKGas * 7,
		}
	}, contracts[2]))
	require.NoError(t, err, "There should not be an error")

	ctx.GasMeter().ConsumeGas(140, "")

	ctx.GasMeter().ConsumeGas(3, "")

	// {{}{{ Session 3: Add vm record
	err = AddVMRecord(ctx, &VMRecord{
		OriginalVMGas: 1,
		ActualVMGas:   2,
	})
	require.NoError(t, err, "We should be able to add vm record")

	// {{}{{} Session 3: Destroyed
	err = DestroySession(&ctx)
	require.NoError(t, err, "There should not be any error")

	require.Equal(t, uint64(143*7), mainMeter.GasConsumed(), "main meter must have consumed 143*7 gas")

	// {{}{{} Session 2: VM Record added
	err = AddVMRecord(ctx, &VMRecord{
		OriginalVMGas: 3,
		ActualVMGas:   4,
	})
	require.NoError(t, err, "We should be able to add vm record")

	require.Equal(t, uint64(143*7), mainMeter.GasConsumed(), "main meter must have consumed 143*7 gas")

	// {{}{{}} Session 2 Destroyed
	err = DestroySession(&ctx)
	require.NoError(t, err, "There should not be any error")

	require.Equal(t, uint64(143*7)+uint64(100/3), mainMeter.GasConsumed(), "main meter must have consumed 143*7 + 100/3 gas")

	// {{}{{}}{ Session 4 Created
	err = CreateNewSession(ctx, 100)
	require.NoError(t, err, "There should not be any error")

	ctx.GasMeter().ConsumeGas(400, "")

	// {{}{{}}{} Session 4 Destroyed
	err = DestroySession(&ctx)
	require.NoError(t, err, "There should not be any error")

	require.Equal(t, uint64(143*7)+uint64(100/3), mainMeter.GasConsumed(), "main meter must have consumed 143*7 + 100/3 gas")

	// {{}{{}}{} VM Record added for initial gas meter
	err = AddVMRecord(ctx, &VMRecord{
		OriginalVMGas: 5,
		ActualVMGas:   6,
	})
	require.NoError(t, err, "We should be able to add vm record")

	// {{}{{}}{}} Terminated session
	queryGasRecords, sessionRecord, err := TerminateGasTracking(&ctx)
	require.NoError(t, err, "We should be able to terminate")

	require.Equal(t, uint64(143*7)+uint64(100/3)+uint64(850*2), mainMeter.GasConsumed(), "main meter must have consumed 143*7 + 100/3 + 850*2 gas")

	require.Equal(t, 2, len(queryGasRecords), "2 gas meter sessions were created")
	require.Equal(t, contracts[2], queryGasRecords[0].ContractAddress)
	require.Equal(t, uint64(143), queryGasRecords[0].OriginalSDKGas)
	require.Equal(t, uint64(143*7), queryGasRecords[0].ActualSDKGas)
	require.Equal(t, uint64(1), queryGasRecords[0].OriginalVMGas)
	require.Equal(t, uint64(2), queryGasRecords[0].ActualVMGas)

	require.Equal(t, contracts[1], queryGasRecords[1].ContractAddress)
	require.Equal(t, uint64(100), queryGasRecords[1].OriginalSDKGas)
	require.Equal(t, uint64(33), queryGasRecords[1].ActualSDKGas)
	require.Equal(t, uint64(3), queryGasRecords[1].OriginalVMGas)
	require.Equal(t, uint64(4), queryGasRecords[1].ActualVMGas)

	require.Equal(t, contracts[0], sessionRecord.ContractAddress)
	require.Equal(t, uint64(850), sessionRecord.OriginalSDKGas)
	require.Equal(t, uint64(1700), sessionRecord.ActualSDKGas)
	require.Equal(t, uint64(5), sessionRecord.OriginalVMGas)
	require.Equal(t, uint64(6), sessionRecord.ActualVMGas)
}
