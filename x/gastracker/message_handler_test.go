package gastracker

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmvm/types"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ sdk.GasMeter = &LoggingGasMeter{}

type callLog struct {
	MethodName string
	InputArg   []string
	Output     []string
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
		Output:     []string{fmt.Sprint(o)},
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
	cosmosMsg := types.CosmosMsg{}
	out, err := json.Marshal(contractOperationInfo)
	if err != nil {
		return cosmosMsg, err
	}
	cosmosMsg.Custom = out
	return cosmosMsg, nil
}

type messageHandlerTestParams struct {
	ctx                   sdk.Context
	loggingGasMeter       *LoggingGasMeter
	gasConsumptionHandler GasConsumptionMsgHandler
	loggingKeeper         *loggingGasTrackerKeeper
	testAccAddress1       sdk.AccAddress
	testAccAddress2       sdk.AccAddress
	cosmosMsg             types.CosmosMsg
}

func setupMessageHandlerTest(t *testing.T, ctx sdk.Context, keeper GasTrackingKeeper) messageHandlerTestParams {
	l := LoggingGasMeter{
		underlyingGasMeter: sdk.NewInfiniteGasMeter(),
		log:                nil,
	}
	loggingKeeper := loggingGasTrackerKeeper{underlyingKeeper: keeper}
	ctx = ctx.WithGasMeter(&l)

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	gasConsumptionMsgHandler := GasConsumptionMsgHandler{gastrackingKeeper: &loggingKeeper}

	firstContractAddress, err := sdk.AccAddressFromBech32("archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt")
	assert.NoError(t, err, "Hardcoded bech32 address should be valid account address")

	secondContractAddress, err := sdk.AccAddressFromBech32("archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk")
	assert.NoError(t, err, "Hardcoded bech32 address should be valid account address")

	cosmosMsg := types.CosmosMsg{
		Custom: []byte{1, 2, 3},
	}

	return messageHandlerTestParams{
		ctx:                   ctx,
		loggingGasMeter:       &l,
		gasConsumptionHandler: gasConsumptionMsgHandler,
		loggingKeeper:         &loggingKeeper,
		testAccAddress1:       firstContractAddress,
		testAccAddress2:       secondContractAddress,
		cosmosMsg:             cosmosMsg,
	}
}

// Test 1: Non-instantiation operation without block gas tracking in place
func TestMessageHandlerInvalidEnv1(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t)
	testParams := setupMessageHandlerTest(t, ctx, keeper)
	gasConsumptionMsgHandler := testParams.gasConsumptionHandler
	firstContractAddress := testParams.testAccAddress1
	loggingGasMeter := testParams.loggingGasMeter
	loggingKeeper := testParams.loggingKeeper
	cosmosMsg := testParams.cosmosMsg
	ctx = testParams.ctx

	contractOperationInfo := gstTypes.ContractOperationInfo{
		GasConsumed:              0,
		Operation:                gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
		RewardAddress:            "",
		GasRebateToEndUser:       false,
		CollectPremium:           false,
		PremiumPercentageCharged: 0,
	}
	cosmosMsg, err := createCosmosMsg(contractOperationInfo)
	assert.NoError(t, err, "Json unmarshalling should not fail")

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.EqualError(
		t,
		err,
		gstTypes.ErrBlockTrackingDataNotFound.Error(),
		"If no block tracking data exists "+
			"handler should throw an error",
	)

	filteredLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	require.Equal(t, 1, len(loggingKeeper.callLogs))
	filteredKeeperLogs := loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, gstTypes.ErrBlockTrackingDataNotFound, filteredKeeperLogs[0].Error)

	loggingKeeper.ResetLogs()
	loggingGasMeter.ClearLogs()
}

// Test 2: Instantiation operation without blockgastracking in place
func TestMessageHandlerInvalidEnv2(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t)
	testParams := setupMessageHandlerTest(t, ctx, keeper)
	gasConsumptionMsgHandler := testParams.gasConsumptionHandler
	firstContractAddress := testParams.testAccAddress1
	secondContractAddress := testParams.testAccAddress2
	loggingGasMeter := testParams.loggingGasMeter
	loggingKeeper := testParams.loggingKeeper
	cosmosMsg := testParams.cosmosMsg
	ctx = testParams.ctx

	contractOperationInfo := gstTypes.ContractOperationInfo{
		GasConsumed:              100,
		Operation:                gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
		RewardAddress:            secondContractAddress.String(),
		GasRebateToEndUser:       false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
	}
	cosmosMsg, err := createCosmosMsg(contractOperationInfo)
	assert.NoError(t, err, "Json unmarshalling should not fail")

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.EqualError(
		t,
		err,
		gstTypes.ErrBlockTrackingDataNotFound.Error(),
		"If no block tracking data exists "+
			"handler should throw an error",
	)
	_, err = keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.EqualError(t, err, gstTypes.ErrContractInstanceMetadataNotFound.Error(), "Contract instance metadata should not be stored by the Message handler")

	filteredLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	require.Equal(t, 1, len(loggingKeeper.callLogs))
	filteredKeeperLogs := loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, gstTypes.ErrBlockTrackingDataNotFound, filteredKeeperLogs[0].Error)

	loggingKeeper.ResetLogs()
	loggingGasMeter.ClearLogs()
}

// Test 3: Instantiation operation without tx tracking in place
func TestMessageHandlerInvalidEnv3(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t)
	testParams := setupMessageHandlerTest(t, ctx, keeper)
	gasConsumptionMsgHandler := testParams.gasConsumptionHandler
	firstContractAddress := testParams.testAccAddress1
	secondContractAddress := testParams.testAccAddress2
	loggingGasMeter := testParams.loggingGasMeter
	loggingKeeper := testParams.loggingKeeper
	cosmosMsg := testParams.cosmosMsg
	ctx = testParams.ctx

	err := keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{})
	assert.NoError(t, err)

	contractOperationInfo := gstTypes.ContractOperationInfo{
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
		"If no tx tracking data exists "+
			"handler should throw an error",
	)

	_, err = keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.EqualError(t, err, gstTypes.ErrContractInstanceMetadataNotFound.Error(), "Contract instance metadata should not be stored by the Message handler")

	filteredLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	require.Equal(t, 1, len(loggingKeeper.callLogs))
	filteredKeeperLogs := loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, gstTypes.ErrTxTrackingDataNotFound, filteredKeeperLogs[0].Error)

	loggingKeeper.ResetLogs()
	loggingGasMeter.ClearLogs()
}

// Test 4: Non-instantiation operation without contract instance metadata in place
func TestMessageHandlerInvalidEnv4(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t)
	testParams := setupMessageHandlerTest(t, ctx, keeper)
	gasConsumptionMsgHandler := testParams.gasConsumptionHandler
	firstContractAddress := testParams.testAccAddress1
	secondContractAddress := testParams.testAccAddress2
	loggingGasMeter := testParams.loggingGasMeter
	loggingKeeper := testParams.loggingKeeper
	cosmosMsg := testParams.cosmosMsg
	ctx = testParams.ctx

	testDecCoin := sdk.NewDecCoin("test", sdk.NewInt(1))

	err := keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{})
	assert.NoError(t, err)

	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{&testDecCoin}, 5)
	assert.NoError(t, err)

	contractOperationInfo := gstTypes.ContractOperationInfo{
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
		"If no contract instance metadata exists "+
			"handler should throw an error except while instantiating",
	)

	_, err = keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.EqualError(t, err, gstTypes.ErrContractInstanceMetadataNotFound.Error(), "Contract instance metadata should not be stored by the Message handler")

	filteredLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	require.Equal(t, 2, len(loggingKeeper.callLogs))
	filteredKeeperLogs := loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, gstTypes.TransactionTracking{
		MaxGasAllowed:         5,
		MaxContractRewards:    []*sdk.DecCoin{&testDecCoin},
		ContractTrackingInfos: nil,
	}, filteredKeeperLogs[0].TransactionTracking)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, gstTypes.ErrContractInstanceMetadataNotFound, filteredKeeperLogs[0].Error)

	loggingKeeper.ResetLogs()
	loggingGasMeter.ClearLogs()
}

// Test 5: Instantiation operation with premium collection on
func TestMessageHandlerSuccessfulInstantiation1(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t)
	testParams := setupMessageHandlerTest(t, ctx, keeper)
	gasConsumptionMsgHandler := testParams.gasConsumptionHandler
	firstContractAddress := testParams.testAccAddress1
	secondContractAddress := testParams.testAccAddress2
	loggingGasMeter := testParams.loggingGasMeter
	loggingKeeper := testParams.loggingKeeper
	cosmosMsg := testParams.cosmosMsg
	ctx = testParams.ctx

	testDecCoin := sdk.NewDecCoin("test", sdk.NewInt(1))

	err := keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{})
	assert.NoError(t, err)

	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{&testDecCoin}, 5)
	assert.NoError(t, err)

	contractOperationInfo := gstTypes.ContractOperationInfo{
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

	filteredLogs := loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 1)
	filteredLog := filteredLogs[0]
	assert.Equal(t, filteredLog.InputArg[0], fmt.Sprint((contractOperationInfo.PremiumPercentageCharged*contractOperationInfo.GasConsumed)/100))

	filteredLogs = loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	require.Equal(t, 3, len(loggingKeeper.callLogs))
	filteredKeeperLogs := loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, gstTypes.TransactionTracking{
		MaxGasAllowed:         5,
		MaxContractRewards:    []*sdk.DecCoin{&testDecCoin},
		ContractTrackingInfos: nil,
	}, filteredKeeperLogs[0].TransactionTracking)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("AddNewContractMetadata")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, contractInstanceMetadata, filteredKeeperLogs[0].Metadata)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 0, len(filteredKeeperLogs))

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, firstContractAddress.String(), filteredKeeperLogs[0].ContractAddress)
	require.Equal(t, contractOperationInfo.GasConsumed, filteredKeeperLogs[0].GasUsed)
	require.Equal(t, contractOperationInfo.Operation, filteredKeeperLogs[0].Operation)
	require.Equal(t, !contractOperationInfo.GasRebateToEndUser, filteredKeeperLogs[0].IsEligibleForReward)

	loggingKeeper.ResetLogs()

	loggingGasMeter.ClearLogs()

	blockGasTracking, err := keeper.GetCurrentBlockGasTracking(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(blockGasTracking.TxTrackingInfos), 1, "We should have at least one tx tracking info")

	lastTxTrackingInfo := blockGasTracking.TxTrackingInfos[len(blockGasTracking.TxTrackingInfos)-1]
	assert.Equal(t, uint64(5), lastTxTrackingInfo.MaxGasAllowed)
	assert.Equal(t, *lastTxTrackingInfo.MaxContractRewards[0], sdk.NewDecCoin("test", sdk.NewInt(1)))
	assert.Condition(t, func() bool {
		return len(lastTxTrackingInfo.ContractTrackingInfos) == 1
	})
	lastContractTrackingInfo := lastTxTrackingInfo.ContractTrackingInfos[len(lastTxTrackingInfo.ContractTrackingInfos)-1]
	assert.Equal(t, lastContractTrackingInfo.GasConsumed, contractOperationInfo.GasConsumed)
	assert.Equal(t, lastContractTrackingInfo.Address, firstContractAddress.String())
	assert.Equal(t, lastContractTrackingInfo.Operation, contractOperationInfo.Operation)
	assert.Equal(t, lastContractTrackingInfo.IsEligibleForReward, !contractOperationInfo.GasRebateToEndUser)
}

// Test 6: Instantiation operation with gas rebate on
func TestMessageHandlerSuccessfulInstantiation2(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t)
	testParams := setupMessageHandlerTest(t, ctx, keeper)
	gasConsumptionMsgHandler := testParams.gasConsumptionHandler
	firstContractAddress := testParams.testAccAddress1
	secondContractAddress := testParams.testAccAddress2
	loggingGasMeter := testParams.loggingGasMeter
	loggingKeeper := testParams.loggingKeeper
	cosmosMsg := testParams.cosmosMsg
	ctx = testParams.ctx

	testDecCoin := sdk.NewDecCoin("test", sdk.NewInt(1))

	err := keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{})
	assert.NoError(t, err)

	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{&testDecCoin}, 5)
	assert.NoError(t, err)

	contractOperationInfo := gstTypes.ContractOperationInfo{
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

	contractInstanceMetadata, err := keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.NoError(t, err, "Contract instance metadata should be stored by the Message handler")

	assert.Equal(t, contractInstanceMetadata.PremiumPercentageCharged, contractOperationInfo.PremiumPercentageCharged)
	assert.Equal(t, contractInstanceMetadata.CollectPremium, contractOperationInfo.CollectPremium)
	assert.Equal(t, contractInstanceMetadata.RewardAddress, contractOperationInfo.RewardAddress)
	assert.Equal(t, contractInstanceMetadata.GasRebateToUser, contractOperationInfo.GasRebateToEndUser)

	filteredLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 1)
	filteredLog := filteredLogs[0]
	assert.Equal(t, filteredLog.InputArg[0], fmt.Sprint(contractOperationInfo.GasConsumed))

	filteredLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	require.Equal(t, 3, len(loggingKeeper.callLogs))
	filteredKeeperLogs := loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, uint64(5), filteredKeeperLogs[0].TransactionTracking.MaxGasAllowed)
	require.Equal(t, []*sdk.DecCoin{&testDecCoin}, filteredKeeperLogs[0].TransactionTracking.MaxContractRewards)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("AddNewContractMetadata")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, contractInstanceMetadata, filteredKeeperLogs[0].Metadata)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 0, len(filteredKeeperLogs))

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, firstContractAddress.String(), filteredKeeperLogs[0].ContractAddress)
	require.Equal(t, contractOperationInfo.GasConsumed, filteredKeeperLogs[0].GasUsed)
	require.Equal(t, contractOperationInfo.Operation, filteredKeeperLogs[0].Operation)
	require.Equal(t, !contractOperationInfo.GasRebateToEndUser, filteredKeeperLogs[0].IsEligibleForReward)

	loggingKeeper.ResetLogs()

	loggingGasMeter.ClearLogs()

	blockGasTracking, err := keeper.GetCurrentBlockGasTracking(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(blockGasTracking.TxTrackingInfos), 1, "We should have at least one tx tracking info")

	lastTxTrackingInfo := blockGasTracking.TxTrackingInfos[len(blockGasTracking.TxTrackingInfos)-1]
	assert.Equal(t, uint64(5), lastTxTrackingInfo.MaxGasAllowed)
	assert.Equal(t, *lastTxTrackingInfo.MaxContractRewards[0], sdk.NewDecCoin("test", sdk.NewInt(1)))
	assert.Condition(t, func() bool {
		return len(lastTxTrackingInfo.ContractTrackingInfos) == 1
	})
	lastContractTrackingInfo := lastTxTrackingInfo.ContractTrackingInfos[len(lastTxTrackingInfo.ContractTrackingInfos)-1]
	assert.Equal(t, lastContractTrackingInfo.GasConsumed, contractOperationInfo.GasConsumed)
	assert.Equal(t, lastContractTrackingInfo.Address, firstContractAddress.String())
	assert.Equal(t, lastContractTrackingInfo.Operation, contractOperationInfo.Operation)
	assert.Equal(t, lastContractTrackingInfo.IsEligibleForReward, !contractOperationInfo.GasRebateToEndUser)
}

// Test 7: Execution operation with contract instance metadata, block and gas tracking in place
func TestMessageHandlerSuccessfulExecution(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t)
	testParams := setupMessageHandlerTest(t, ctx, keeper)
	gasConsumptionMsgHandler := testParams.gasConsumptionHandler
	firstContractAddress := testParams.testAccAddress1
	secondContractAddress := testParams.testAccAddress2
	loggingGasMeter := testParams.loggingGasMeter
	loggingKeeper := testParams.loggingKeeper
	cosmosMsg := testParams.cosmosMsg
	ctx = testParams.ctx

	testDecCoin := sdk.NewDecCoin("test", sdk.NewInt(1))

	err := keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{})
	require.NoError(t, err)

	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{&testDecCoin}, 5)
	require.NoError(t, err)

	originalContractInstanceMetadata := gstTypes.ContractInstanceMetadata{
		RewardAddress:            secondContractAddress.String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 80,
	}
	err = keeper.AddNewContractMetadata(ctx, firstContractAddress.String(), originalContractInstanceMetadata)
	require.NoError(t, err)

	executionContractOperationInfo := gstTypes.ContractOperationInfo{
		GasConsumed:              300,
		Operation:                gstTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION,
		RewardAddress:            secondContractAddress.String(),
		GasRebateToEndUser:       false,
		CollectPremium:           true,
		PremiumPercentageCharged: 60,
	}
	cosmosMsg, err = createCosmosMsg(executionContractOperationInfo)
	require.NoError(t, err, "Json unmarshalling should not fail")

	_, _, err = gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	require.NoError(
		t,
		err,
		"Handler should succeed when block tracking and tx tracking both are available",
	)

	// Since this is non instantiation call, we will not taking into account other parameters of ContractOperationInfo
	contractInstanceMetadata, err := keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	require.NoError(t, err, "Contract instance metadata should be stored by the Message handler")

	// ContractInstanceMetadata should not be modified
	require.Equal(t, originalContractInstanceMetadata, contractInstanceMetadata)

	filteredLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	require.Equal(t, 1, len(filteredLogs))
	filteredLog := filteredLogs[0]
	require.Equal(t, fmt.Sprint(executionContractOperationInfo.GasConsumed), filteredLog.InputArg[0])

	filteredLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	require.Equal(t, 0, len(filteredLogs))

	require.Equal(t, 3, len(loggingKeeper.callLogs))
	filteredKeeperLogs := loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, uint64(5), filteredKeeperLogs[0].TransactionTracking.MaxGasAllowed)
	require.Equal(t, []*sdk.DecCoin{&testDecCoin}, filteredKeeperLogs[0].TransactionTracking.MaxContractRewards)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("AddNewContractMetadata")
	require.Equal(t, 0, len(filteredKeeperLogs))

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, contractInstanceMetadata, filteredKeeperLogs[0].Metadata)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, firstContractAddress.String(), filteredKeeperLogs[0].ContractAddress)
	require.Equal(t, executionContractOperationInfo.GasConsumed, filteredKeeperLogs[0].GasUsed)
	require.Equal(t, executionContractOperationInfo.Operation, filteredKeeperLogs[0].Operation)
	require.Equal(t, !contractInstanceMetadata.GasRebateToUser, filteredKeeperLogs[0].IsEligibleForReward)

	loggingKeeper.ResetLogs()

	loggingGasMeter.ClearLogs()

	blockGasTracking, err := keeper.GetCurrentBlockGasTracking(ctx)
	require.NoError(t, err)
	require.Equal(t, len(blockGasTracking.TxTrackingInfos), 1, "We should have at least one tx tracking info")

	lastTxTrackingInfo := blockGasTracking.TxTrackingInfos[len(blockGasTracking.TxTrackingInfos)-1]
	require.Equal(t, uint64(5), lastTxTrackingInfo.MaxGasAllowed)
	require.Equal(t, *lastTxTrackingInfo.MaxContractRewards[0], sdk.NewDecCoin("test", sdk.NewInt(1)))
	require.Condition(t, func() bool {
		return len(lastTxTrackingInfo.ContractTrackingInfos) == 1
	})
	lastContractTrackingInfo := lastTxTrackingInfo.ContractTrackingInfos[len(lastTxTrackingInfo.ContractTrackingInfos)-1]
	require.Equal(t, executionContractOperationInfo.GasConsumed, lastContractTrackingInfo.GasConsumed)
	require.Equal(t, firstContractAddress.String(), lastContractTrackingInfo.Address)
	require.Equal(t, executionContractOperationInfo.Operation, lastContractTrackingInfo.Operation)
	require.Equal(t, !originalContractInstanceMetadata.GasRebateToUser, lastContractTrackingInfo.IsEligibleForReward)
}

// Test 8: Instantiation operation with premium collection off
func TestMessageHandlerSuccessfulInstantiation3(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t)
	keeper.SetParams(ctx, disableContractPremium(gstTypes.DefaultParams()))
	testParams := setupMessageHandlerTest(t, ctx, keeper)
	gasConsumptionMsgHandler := testParams.gasConsumptionHandler
	firstContractAddress := testParams.testAccAddress1
	secondContractAddress := testParams.testAccAddress2
	loggingGasMeter := testParams.loggingGasMeter
	loggingKeeper := testParams.loggingKeeper
	cosmosMsg := testParams.cosmosMsg
	ctx = testParams.ctx

	testDecCoin := sdk.NewDecCoin("test", sdk.NewInt(1))

	err := keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{})
	assert.NoError(t, err)

	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{&testDecCoin}, 5)
	assert.NoError(t, err)

	contractOperationInfo := gstTypes.ContractOperationInfo{
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

	filteredLogs := loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	require.Equal(t, 3, len(loggingKeeper.callLogs))
	filteredKeeperLogs := loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, gstTypes.TransactionTracking{
		MaxGasAllowed:         5,
		MaxContractRewards:    []*sdk.DecCoin{&testDecCoin},
		ContractTrackingInfos: nil,
	}, filteredKeeperLogs[0].TransactionTracking)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("AddNewContractMetadata")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, contractInstanceMetadata, filteredKeeperLogs[0].Metadata)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 0, len(filteredKeeperLogs))

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, firstContractAddress.String(), filteredKeeperLogs[0].ContractAddress)
	require.Equal(t, contractOperationInfo.GasConsumed, filteredKeeperLogs[0].GasUsed)
	require.Equal(t, contractOperationInfo.Operation, filteredKeeperLogs[0].Operation)
	require.Equal(t, !contractOperationInfo.GasRebateToEndUser, filteredKeeperLogs[0].IsEligibleForReward)

	loggingKeeper.ResetLogs()

	loggingGasMeter.ClearLogs()

	blockGasTracking, err := keeper.GetCurrentBlockGasTracking(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(blockGasTracking.TxTrackingInfos), 1, "We should have at least one tx tracking info")

	lastTxTrackingInfo := blockGasTracking.TxTrackingInfos[len(blockGasTracking.TxTrackingInfos)-1]
	assert.Equal(t, uint64(5), lastTxTrackingInfo.MaxGasAllowed)
	assert.Equal(t, *lastTxTrackingInfo.MaxContractRewards[0], sdk.NewDecCoin("test", sdk.NewInt(1)))
	assert.Condition(t, func() bool {
		return len(lastTxTrackingInfo.ContractTrackingInfos) == 1
	})
	lastContractTrackingInfo := lastTxTrackingInfo.ContractTrackingInfos[len(lastTxTrackingInfo.ContractTrackingInfos)-1]
	assert.Equal(t, lastContractTrackingInfo.GasConsumed, contractOperationInfo.GasConsumed)
	assert.Equal(t, lastContractTrackingInfo.Address, firstContractAddress.String())
	assert.Equal(t, lastContractTrackingInfo.Operation, contractOperationInfo.Operation)
	assert.Equal(t, lastContractTrackingInfo.IsEligibleForReward, !contractOperationInfo.GasRebateToEndUser)
}

// Test 9: Instantiation operation with gas rebates off
func TestMessageHandlerSuccessfulInstantiation4(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t)
	keeper.SetParams(ctx, disableGasRebateToUser(gstTypes.DefaultParams()))
	testParams := setupMessageHandlerTest(t, ctx, keeper)
	gasConsumptionMsgHandler := testParams.gasConsumptionHandler
	firstContractAddress := testParams.testAccAddress1
	secondContractAddress := testParams.testAccAddress2
	loggingGasMeter := testParams.loggingGasMeter
	loggingKeeper := testParams.loggingKeeper
	cosmosMsg := testParams.cosmosMsg
	ctx = testParams.ctx

	testDecCoin := sdk.NewDecCoin("test", sdk.NewInt(1))

	err := keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{})
	assert.NoError(t, err)

	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{&testDecCoin}, 5)
	assert.NoError(t, err)

	contractOperationInfo := gstTypes.ContractOperationInfo{
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

	contractInstanceMetadata, err := keeper.GetNewContractMetadata(ctx, firstContractAddress.String())
	assert.NoError(t, err, "Contract instance metadata should be stored by the Message handler")

	assert.Equal(t, contractInstanceMetadata.PremiumPercentageCharged, contractOperationInfo.PremiumPercentageCharged)
	assert.Equal(t, contractInstanceMetadata.CollectPremium, contractOperationInfo.CollectPremium)
	assert.Equal(t, contractInstanceMetadata.RewardAddress, contractOperationInfo.RewardAddress)
	assert.Equal(t, contractInstanceMetadata.GasRebateToUser, contractOperationInfo.GasRebateToEndUser)

	filteredLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	require.Equal(t, 3, len(loggingKeeper.callLogs))
	filteredKeeperLogs := loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, uint64(5), filteredKeeperLogs[0].TransactionTracking.MaxGasAllowed)
	require.Equal(t, []*sdk.DecCoin{&testDecCoin}, filteredKeeperLogs[0].TransactionTracking.MaxContractRewards)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("AddNewContractMetadata")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, contractInstanceMetadata, filteredKeeperLogs[0].Metadata)

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 0, len(filteredKeeperLogs))

	filteredKeeperLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredKeeperLogs))
	require.Equal(t, nil, filteredKeeperLogs[0].Error)
	require.Equal(t, firstContractAddress.String(), filteredKeeperLogs[0].ContractAddress)
	require.Equal(t, contractOperationInfo.GasConsumed, filteredKeeperLogs[0].GasUsed)
	require.Equal(t, contractOperationInfo.Operation, filteredKeeperLogs[0].Operation)
	require.Equal(t, !contractOperationInfo.GasRebateToEndUser, filteredKeeperLogs[0].IsEligibleForReward)

	loggingKeeper.ResetLogs()

	loggingGasMeter.ClearLogs()

	blockGasTracking, err := keeper.GetCurrentBlockGasTracking(ctx)
	assert.NoError(t, err)
	assert.Equal(t, len(blockGasTracking.TxTrackingInfos), 1, "We should have at least one tx tracking info")

	lastTxTrackingInfo := blockGasTracking.TxTrackingInfos[len(blockGasTracking.TxTrackingInfos)-1]
	assert.Equal(t, uint64(5), lastTxTrackingInfo.MaxGasAllowed)
	assert.Equal(t, *lastTxTrackingInfo.MaxContractRewards[0], sdk.NewDecCoin("test", sdk.NewInt(1)))
	assert.Condition(t, func() bool {
		return len(lastTxTrackingInfo.ContractTrackingInfos) == 1
	})
	lastContractTrackingInfo := lastTxTrackingInfo.ContractTrackingInfos[len(lastTxTrackingInfo.ContractTrackingInfos)-1]
	assert.Equal(t, lastContractTrackingInfo.GasConsumed, contractOperationInfo.GasConsumed)
	assert.Equal(t, lastContractTrackingInfo.Address, firstContractAddress.String())
	assert.Equal(t, lastContractTrackingInfo.Operation, contractOperationInfo.Operation)
	assert.Equal(t, lastContractTrackingInfo.IsEligibleForReward, !contractOperationInfo.GasRebateToEndUser)
}

// Test 0: Invalid JSON passed
func TestMessageHandlerInvalidJSON(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t)
	testParams := setupMessageHandlerTest(t, ctx, keeper)
	gasConsumptionMsgHandler := testParams.gasConsumptionHandler
	firstContractAddress := testParams.testAccAddress1
	loggingGasMeter := testParams.loggingGasMeter
	loggingKeeper := testParams.loggingKeeper
	cosmosMsg := testParams.cosmosMsg
	ctx = testParams.ctx

	_, _, err := gasConsumptionMsgHandler.DispatchMsg(ctx, firstContractAddress, "", cosmosMsg)
	assert.EqualError(
		t,
		err,
		fmt.Sprintf("invalid character '\\x0%d' looking for beginning of value", 1),
		"If custom message is invalid "+
			"handler should return unmarshalling error",
	)

	filteredLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	filteredLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	assert.Equal(t, len(filteredLogs), 0)

	require.Equal(t, 0, len(loggingKeeper.callLogs))
	loggingKeeper.ResetLogs()

	loggingGasMeter.ClearLogs()
}
