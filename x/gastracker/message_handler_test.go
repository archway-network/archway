package gastracker

import (
	"encoding/json"
	"fmt"
	"github.com/CosmWasm/wasmvm/types"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ sdk.GasMeter = &LoggingGasMeter{}

type callLog struct {
	MethodName string
	InputArg []string
	Output []string
}

type callLogs []callLog

func (c callLogs) FilterLogWithMethodName(methodName string) callLogs {
	filteredCallLogs := make(callLogs, 0)

	for _, log := range c {
		if log.MethodName == methodName {
			filteredCallLogs = append(filteredCallLogs, log)
		}
	}

	return filteredCallLogs
}

func (c callLogs) FilterLogWithDescriptor(descriptor string) callLogs {
	filteredCallLogs := make(callLogs, 0)

	for _, log := range c {
		if log.MethodName == "ConsumeGas" || log.MethodName == "RefundGas" {
			if log.InputArg[1] == descriptor {
				filteredCallLogs = append(filteredCallLogs, log)
			}
		}
	}

	return filteredCallLogs
}

type LoggingGasMeter struct {
	underlyingGasMeter sdk.GasMeter
	log                callLogs
}

func (l *LoggingGasMeter) Logs() callLogs {
	return l.log
}

func (l *LoggingGasMeter) ClearLogs() {
	l.log = []callLog{}
}

func (l *LoggingGasMeter) GasConsumed() sdk.Gas {
	o := l.underlyingGasMeter.GasConsumed()
	l.log = append(l.log, callLog{
		MethodName: "GasConsumed",
		InputArg:   []string{},
		Output:     []string{fmt.Sprint(o)},
	})
	return o
}

func (l *LoggingGasMeter) GasConsumedToLimit() sdk.Gas {
	o := l.underlyingGasMeter.GasConsumedToLimit()
	l.log = append(l.log, callLog{
		MethodName: "GasConsumedToLimit",
		InputArg:   []string{},
		Output:    []string{fmt.Sprint(o)},
	})
	return o
}

func (l *LoggingGasMeter) Limit() sdk.Gas {
	o := l.underlyingGasMeter.Limit()
	l.log = append(l.log, callLog{
		MethodName: "Limit",
		InputArg:   []string{},
		Output:     []string{fmt.Sprint(o)},
	})
	return o
}

func (l *LoggingGasMeter) ConsumeGas(amount sdk.Gas, descriptor string) {
	l.underlyingGasMeter.ConsumeGas(amount, descriptor)
	l.log = append(l.log, callLog{
		MethodName: "ConsumeGas",
		InputArg:   []string{fmt.Sprint(amount), fmt.Sprint(descriptor)},
		Output:     []string{},
	})
}

func (l *LoggingGasMeter) RefundGas(amount sdk.Gas, descriptor string) {
	l.underlyingGasMeter.RefundGas(amount, descriptor)
	l.log = append(l.log, callLog{
		MethodName: "RefundGas",
		InputArg:   []string{fmt.Sprint(amount), fmt.Sprint(descriptor)},
		Output:     []string{},
	})
}

func (l *LoggingGasMeter) IsPastLimit() bool {
	o := l.underlyingGasMeter.IsPastLimit()
	l.log = append(l.log, callLog{
		MethodName: "IsPastLimit",
		InputArg:   []string{},
		Output:     []string{fmt.Sprint(o)},
	})
	return o
}

func (l *LoggingGasMeter) IsOutOfGas() bool {
	o := l.underlyingGasMeter.IsOutOfGas()
	l.log = append(l.log, callLog{
		MethodName: "IsOutOfGas",
		InputArg:   []string{},
		Output:     []string{fmt.Sprint(o)},
	})
	return o
}

func (l *LoggingGasMeter) String() string {
	o := l.underlyingGasMeter.String()
	l.log = append(l.log, callLog{
		MethodName: "String",
		InputArg:   []string{},
		Output:     []string{fmt.Sprint(o)},
	})
	return o
}

func createCosmosMsg(contractOperationInfo gstTypes.ContractOperationInfo) (types.CosmosMsg, error) {
	cosmosMsg := types.CosmosMsg{
	}
	out, err := json.Marshal(contractOperationInfo)
	if err != nil {
		return cosmosMsg, err
	}
	cosmosMsg.Custom = out
	return cosmosMsg, nil
}

func TestMessageHandler(t *testing.T) {
	l := LoggingGasMeter{
		underlyingGasMeter: sdk.NewInfiniteGasMeter(),
		log:                nil,
	}

	ctx, keeper := CreateTestKeeperAndContext(t)
	ctx = ctx.WithGasMeter(&l)

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	gasConsumptionMsgHandler := GasConsumptionMsgHandler{gastrackingKeeper: keeper}

	firstContractAddress, err := sdk.AccAddressFromBech32("archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt")
	assert.NoError(t, err, "Hardcoded bech32 address should be valid account address")

	secondContractAddress, err := sdk.AccAddressFromBech32("archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk")
	assert.NoError(t, err, "Hardcoded bech32 address should be valid account address")

	_, err = sdk.AccAddressFromBech32("archway14hj2tavq8fpesdwxxcu44rty3hh90vhudldltd")
	assert.NoError(t, err, "Hardcoded bech32 address should be valid account address")

	_, err = sdk.AccAddressFromBech32("archway1aakfpghcanxtc45gpqlx8j3rq0zcpyf4jtf0e4")
	assert.NoError(t, err, "Hardcoded bech32 address should be valid account address")


	// Test 0: Invalid JSON passed
	cosmosMsg := types.CosmosMsg{
		Custom:       []byte{1,2,3},
	}

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.EqualError(
		t,
		err,
		fmt.Sprintf("invalid character '\\x0%d' looking for beginning of value", 1),
		"If custom message is invalid " +
			"handler should return unmarshalling error",
	)

	filteredLogs := l.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = l.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	l.ClearLogs()

	// Test 1: Non-instantiation operation without block gas tracking in place
	contractOperationInfo := gstTypes.ContractOperationInfo{
		GasConsumed:              0,
		Operation:                gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
		RewardAddress:            "",
		GasRebateToEndUser:       false,
		CollectPremium:           false,
		PremiumPercentageCharged: 0,
	}
	cosmosMsg, err = createCosmosMsg(contractOperationInfo)
	assert.NoError(t, err, "Json unmarshalling should not fail")

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.EqualError(
		t,
		err,
		gstTypes.ErrBlockTrackingDataNotFound.Error(),
		"If no block tracking data exists " +
			"handler should throw an error",
	)

	filteredLogs = l.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = l.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	l.ClearLogs()


	// Test 2: Instantiation operation without blockgastracking in place
	contractOperationInfo = gstTypes.ContractOperationInfo{
		GasConsumed:              100,
		Operation:                gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
		RewardAddress:            secondContractAddress.String(),
		GasRebateToEndUser:       false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
	}
	cosmosMsg, err = createCosmosMsg(contractOperationInfo)
	assert.NoError(t, err, "Json unmarshalling should not fail")

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.EqualError(
		t,
		err,
		gstTypes.ErrBlockTrackingDataNotFound.Error(),
		"If no block tracking data exists " +
			"handler should throw an error",
	)
	_, err = keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.EqualError(t, err, gstTypes.ErrContractInstanceMetadataNotFound.Error(),"Contract instance metadata should not be stored by the Message handler")

	filteredLogs = l.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = l.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	l.ClearLogs()

	// Test 3: Instantiation operation without tx tracking in place
	err = keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{})
	assert.NoError(t, err)

	contractOperationInfo = gstTypes.ContractOperationInfo{
		GasConsumed:              100,
		Operation:                gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
		RewardAddress:            secondContractAddress.String(),
		GasRebateToEndUser:       false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
	}
	cosmosMsg, err = createCosmosMsg(contractOperationInfo)
	assert.NoError(t, err, "Json unmarshalling should not fail")

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.EqualError(
		t,
		err,
		gstTypes.ErrTxTrackingDataNotFound.Error(),
		"If no tx tracking data exists " +
			"handler should throw an error",
	)

	_, err = keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.EqualError(t, err, gstTypes.ErrContractInstanceMetadataNotFound.Error(),"Contract instance metadata should not be stored by the Message handler")

	filteredLogs = l.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = l.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	l.ClearLogs()

	// Test 4: Non-instantiation operation without contract instance metadata in place
	testDecCoin := sdk.NewDecCoin("test", sdk.NewInt(1))
	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{&testDecCoin}, 5)
	assert.NoError(t, err)

	contractOperationInfo = gstTypes.ContractOperationInfo{
		GasConsumed:              100,
		Operation:                gstTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION,
		RewardAddress:            secondContractAddress.String(),
		GasRebateToEndUser:       true,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
	}
	cosmosMsg, err = createCosmosMsg(contractOperationInfo)
	assert.NoError(t, err, "Json unmarshalling should not fail")

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.EqualError(
		t,
		err,
		gstTypes.ErrContractInstanceMetadataNotFound.Error(),
		"If no contract instance metadata exists " +
			"handler should throw an error except while instantiating",
	)

	_, err = keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.EqualError(t, err, gstTypes.ErrContractInstanceMetadataNotFound.Error(),"Contract instance metadata should not be stored by the Message handler")

	filteredLogs = l.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = l.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	l.ClearLogs()


	// Test 5: Instantiation operation with premium collection on
	contractOperationInfo = gstTypes.ContractOperationInfo{
		GasConsumed:              100,
		Operation:                gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
		RewardAddress:            secondContractAddress.String(),
		GasRebateToEndUser:       false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
	}
	cosmosMsg, err = createCosmosMsg(contractOperationInfo)
	assert.NoError(t, err, "Json unmarshalling should not fail")

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.NoError(
		t,
		err,
		"Handler should succeed when block tracking and tx tracking both are available",
	)

	contractInstanceMetadata, err := keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.NoError(t, err, "Contract instance metadata should be stored by the Message handler")

	assert.Equal(t, contractInstanceMetadata.PremiumPercentageCharged, contractOperationInfo.PremiumPercentageCharged)
	assert.Equal(t, contractInstanceMetadata.CollectPremium, contractOperationInfo.CollectPremium)
	assert.Equal(t, contractInstanceMetadata.RewardAddress, contractOperationInfo.RewardAddress)
	assert.Equal(t, contractInstanceMetadata.GasRebateToUser, contractOperationInfo.GasRebateToEndUser)

	filteredLogs = l.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 1)
	filteredLog := filteredLogs[0]
	assert.Equal(t, filteredLog.InputArg[0], fmt.Sprint((contractOperationInfo.PremiumPercentageCharged * contractOperationInfo.GasConsumed) / 100))

	filteredLogs = l.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	l.ClearLogs()

	blockGasTracking, err := keeper.GetCurrentBlockTrackingInfo(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(blockGasTracking.TxTrackingInfos), 1, "We should have at least one tx tracking info")

	lastTxTrackingInfo := blockGasTracking.TxTrackingInfos[len(blockGasTracking.TxTrackingInfos) - 1]
	assert.Equal(t, uint64(5), lastTxTrackingInfo.MaxGasAllowed)
	assert.Equal(t, *lastTxTrackingInfo.MaxContractRewards[0], sdk.NewDecCoin("test", sdk.NewInt(1)))
	assert.Condition(t, func() bool {
		return len(lastTxTrackingInfo.ContractTrackingInfos) == 1
	})
	lastContractTrackingInfo := lastTxTrackingInfo.ContractTrackingInfos[len(lastTxTrackingInfo.ContractTrackingInfos) - 1]
	assert.Equal(t, lastContractTrackingInfo.GasConsumed, contractOperationInfo.GasConsumed)
	assert.Equal(t, lastContractTrackingInfo.Address, firstContractAddress.String())
	assert.Equal(t, lastContractTrackingInfo.Operation, contractOperationInfo.Operation)
	assert.Equal(t, lastContractTrackingInfo.IsEligibleForReward, !contractOperationInfo.GasRebateToEndUser)


	// Test 6: Instantiation operation with gas rebate on
	contractOperationInfo = gstTypes.ContractOperationInfo{
		GasConsumed:              200,
		Operation:                gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
		RewardAddress:            secondContractAddress.String(),
		GasRebateToEndUser:       true,
		CollectPremium:           false,
		PremiumPercentageCharged: 50,
	}
	cosmosMsg, err = createCosmosMsg(contractOperationInfo)
	assert.NoError(t, err, "Json unmarshalling should not fail")

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.NoError(
		t,
		err,
		"Handler should succeed when block tracking and tx tracking both are available",
	)

	contractInstanceMetadata, err = keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.NoError(t, err, "Contract instance metadata should be stored by the Message handler")

	assert.Equal(t, contractInstanceMetadata.PremiumPercentageCharged, contractOperationInfo.PremiumPercentageCharged)
	assert.Equal(t, contractInstanceMetadata.CollectPremium, contractOperationInfo.CollectPremium)
	assert.Equal(t, contractInstanceMetadata.RewardAddress, contractOperationInfo.RewardAddress)
	assert.Equal(t, contractInstanceMetadata.GasRebateToUser, contractOperationInfo.GasRebateToEndUser)

	filteredLogs = l.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 1)
	filteredLog = filteredLogs[0]
	assert.Equal(t, filteredLog.InputArg[0], fmt.Sprint(contractOperationInfo.GasConsumed))

	filteredLogs = l.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	l.ClearLogs()

	blockGasTracking, err = keeper.GetCurrentBlockTrackingInfo(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(blockGasTracking.TxTrackingInfos), 1, "We should have at least one tx tracking info")

	lastTxTrackingInfo = blockGasTracking.TxTrackingInfos[len(blockGasTracking.TxTrackingInfos) - 1]
	assert.Equal(t, uint64(5), lastTxTrackingInfo.MaxGasAllowed)
	assert.Equal(t, *lastTxTrackingInfo.MaxContractRewards[0], sdk.NewDecCoin("test", sdk.NewInt(1)))
	assert.Condition(t, func() bool {
		return len(lastTxTrackingInfo.ContractTrackingInfos) == 2
	})
	lastContractTrackingInfo = lastTxTrackingInfo.ContractTrackingInfos[len(lastTxTrackingInfo.ContractTrackingInfos) - 1]
	assert.Equal(t, lastContractTrackingInfo.GasConsumed, contractOperationInfo.GasConsumed)
	assert.Equal(t, lastContractTrackingInfo.Address, firstContractAddress.String())
	assert.Equal(t, lastContractTrackingInfo.Operation, contractOperationInfo.Operation)
	assert.Equal(t, lastContractTrackingInfo.IsEligibleForReward, !contractOperationInfo.GasRebateToEndUser)

	// Test 7: Execution operation with contract instance metadata, block and gas tracking in place
	executionContractOperationInfo := gstTypes.ContractOperationInfo{
		GasConsumed:              300,
		Operation:                gstTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION,
		RewardAddress:            secondContractAddress.String(),
		GasRebateToEndUser:       false,
		CollectPremium:           true,
		PremiumPercentageCharged: 60,
	}
	cosmosMsg, err = createCosmosMsg(executionContractOperationInfo)
	assert.NoError(t, err, "Json unmarshalling should not fail")

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.NoError(
		t,
		err,
		"Handler should succeed when block tracking and tx tracking both are available",
	)

	// Since this is non instantiation call, we will not taking into account other parameters of ContractOperationInfo
	contractInstanceMetadata, err = keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.NoError(t, err, "Contract instance metadata should be stored by the Message handler")

	// ContractInstanceMetadata should not be modified
	assert.Equal(t, contractInstanceMetadata.PremiumPercentageCharged, contractOperationInfo.PremiumPercentageCharged)
	assert.Equal(t, contractInstanceMetadata.CollectPremium, contractOperationInfo.CollectPremium)
	assert.Equal(t, contractInstanceMetadata.RewardAddress, contractOperationInfo.RewardAddress)
	assert.Equal(t, contractInstanceMetadata.GasRebateToUser, contractOperationInfo.GasRebateToEndUser)


	filteredLogs = l.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 1)
	filteredLog = filteredLogs[0]
	assert.Equal(t, filteredLog.InputArg[0], fmt.Sprint(executionContractOperationInfo.GasConsumed))

	filteredLogs = l.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	l.ClearLogs()

	blockGasTracking, err = keeper.GetCurrentBlockTrackingInfo(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(blockGasTracking.TxTrackingInfos), 1, "We should have at least one tx tracking info")

	lastTxTrackingInfo = blockGasTracking.TxTrackingInfos[len(blockGasTracking.TxTrackingInfos) - 1]
	assert.Equal(t, uint64(5), lastTxTrackingInfo.MaxGasAllowed)
	assert.Equal(t, *lastTxTrackingInfo.MaxContractRewards[0], sdk.NewDecCoin("test", sdk.NewInt(1)))
	assert.Condition(t, func() bool {
		return len(lastTxTrackingInfo.ContractTrackingInfos) == 3
	})
	lastContractTrackingInfo = lastTxTrackingInfo.ContractTrackingInfos[len(lastTxTrackingInfo.ContractTrackingInfos) - 1]
	assert.Equal(t, lastContractTrackingInfo.GasConsumed, executionContractOperationInfo.GasConsumed)
	assert.Equal(t, lastContractTrackingInfo.Address, firstContractAddress.String())
	assert.Equal(t, lastContractTrackingInfo.Operation, executionContractOperationInfo.Operation)
	assert.Equal(t, lastContractTrackingInfo.IsEligibleForReward, !contractOperationInfo.GasRebateToEndUser)
}



