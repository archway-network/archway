package gastracker

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/CosmWasm/wasmvm/types"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

type gasTrackerKeeperCallLog struct {
	MethodName            string
	Fee                   []*sdk.DecCoin
	GasLimit              uint64
	ContractAddress       string
	RewardAddress         string
	GasUsed               uint64
	Operation             gstTypes.ContractOperation
	IsEligibleForReward   bool
	BlockGasTracking      gstTypes.BlockGasTracking
	TransactionTracking   gstTypes.TransactionTracking
	RewardToBeDistributed sdk.Coins
	Metadata              gstTypes.ContractInstanceMetadata
	ContractRewards       sdk.DecCoins
	LeftOverThreshold     uint64
	LeftOverRewardEntry   gstTypes.LeftOverRewardEntry
	Error                 error
}

type gasTrackerKeeperCallLogs []gasTrackerKeeperCallLog

func (g gasTrackerKeeperCallLogs) FilterByMethod(methodName string) gasTrackerKeeperCallLogs {
	filteredCallLogs := make(gasTrackerKeeperCallLogs, 0)

	for _, log := range g {
		if log.MethodName == methodName {
			filteredCallLogs = append(filteredCallLogs, log)
		}
	}

	return filteredCallLogs
}

func (g gasTrackerKeeperCallLogs) FilterByContractAddress(contractAddress string) gasTrackerKeeperCallLogs {
	filteredCallLogs := make(gasTrackerKeeperCallLogs, 0)

	for _, log := range g {
		if log.ContractAddress == contractAddress {
			filteredCallLogs = append(filteredCallLogs, log)
		}
	}

	return filteredCallLogs
}

func (g gasTrackerKeeperCallLogs) FilterByOperation(operation gstTypes.ContractOperation) gasTrackerKeeperCallLogs {
	filteredCallLogs := make(gasTrackerKeeperCallLogs, 0)

	for _, log := range g {
		if log.Operation == operation {
			filteredCallLogs = append(filteredCallLogs, log)
		}
	}

	return filteredCallLogs
}

type loggingGasTrackerKeeper struct {
	underlyingKeeper GasTrackingKeeper
	callLogs         gasTrackerKeeperCallLogs
}

func (l *loggingGasTrackerKeeper) ResetLogs() {
	l.callLogs = nil
}

func (l *loggingGasTrackerKeeper) TrackNewTx(ctx sdk.Context, fee []*sdk.DecCoin, gasLimit uint64) error {
	err := l.underlyingKeeper.TrackNewTx(ctx, fee, gasLimit)
	log := gasTrackerKeeperCallLog{
		MethodName: "TrackNewTx",
		Fee:        fee,
		GasLimit:   gasLimit,
		Error:      err,
	}

	l.callLogs = append(l.callLogs, log)
	return err
}

func (l *loggingGasTrackerKeeper) TrackContractGasUsage(ctx sdk.Context, contractAddress string, gasUsed uint64, operation gstTypes.ContractOperation, isEligibleForReward bool) error {
	err := l.underlyingKeeper.TrackContractGasUsage(ctx, contractAddress, gasUsed, operation, isEligibleForReward)
	log := gasTrackerKeeperCallLog{
		MethodName:          "TrackContractGasUsage",
		ContractAddress:     contractAddress,
		GasUsed:             gasUsed,
		Operation:           operation,
		IsEligibleForReward: isEligibleForReward,
		Error:               err,
	}

	l.callLogs = append(l.callLogs, log)
	return err
}

func (l *loggingGasTrackerKeeper) GetCurrentBlockTrackingInfo(ctx sdk.Context) (gstTypes.BlockGasTracking, error) {
	blockGasTrackingInfo, err := l.underlyingKeeper.GetCurrentBlockTrackingInfo(ctx)
	log := gasTrackerKeeperCallLog{
		MethodName:       "GetCurrentBlockTrackingInfo",
		BlockGasTracking: blockGasTrackingInfo,
		Error:            err,
	}

	l.callLogs = append(l.callLogs, log)
	return blockGasTrackingInfo, err
}

func (l *loggingGasTrackerKeeper) GetCurrentTxTrackingInfo(ctx sdk.Context) (gstTypes.TransactionTracking, error) {
	txTracking, err := l.underlyingKeeper.GetCurrentTxTrackingInfo(ctx)
	log := gasTrackerKeeperCallLog{
		MethodName:          "GetCurrentTxTrackingInfo",
		TransactionTracking: txTracking,
		Error:               err,
	}

	l.callLogs = append(l.callLogs, log)
	return txTracking, err
}

func (l *loggingGasTrackerKeeper) TrackNewBlock(ctx sdk.Context, blockGasTracking gstTypes.BlockGasTracking) error {
	err := l.underlyingKeeper.TrackNewBlock(ctx, blockGasTracking)
	log := gasTrackerKeeperCallLog{
		MethodName:       "TrackNewBlock",
		BlockGasTracking: blockGasTracking,
		Error:            err,
	}

	l.callLogs = append(l.callLogs, log)
	return err
}

func (l *loggingGasTrackerKeeper) AddNewContractMetadata(ctx sdk.Context, address string, metadata gstTypes.ContractInstanceMetadata) error {
	err := l.underlyingKeeper.AddNewContractMetadata(ctx, address, metadata)
	log := gasTrackerKeeperCallLog{
		MethodName:      "AddNewContractMetadata",
		ContractAddress: address,
		Metadata:        metadata,
		Error:           err,
	}

	l.callLogs = append(l.callLogs, log)
	return err
}

func (l *loggingGasTrackerKeeper) GetNewContractMetadata(ctx sdk.Context, address string) (gstTypes.ContractInstanceMetadata, error) {
	metadata, err := l.underlyingKeeper.GetNewContractMetadata(ctx, address)
	log := gasTrackerKeeperCallLog{
		MethodName:      "GetNewContractMetadata",
		ContractAddress: address,
		Metadata:        metadata,
		Error:           err,
	}
	l.callLogs = append(l.callLogs, log)
	return metadata, err
}

func (l *loggingGasTrackerKeeper) CreateOrMergeLeftOverRewardEntry(ctx sdk.Context, rewardAddress string, contractRewards sdk.DecCoins, leftOverThreshold uint64) (sdk.Coins, error) {
	rewardToBeDistributed, err := l.underlyingKeeper.CreateOrMergeLeftOverRewardEntry(ctx, rewardAddress, contractRewards, leftOverThreshold)

	log := gasTrackerKeeperCallLog{
		MethodName:            "CreateOrMergeLeftOverRewardEntry",
		RewardAddress:         rewardAddress,
		ContractRewards:       contractRewards,
		LeftOverThreshold:     leftOverThreshold,
		RewardToBeDistributed: rewardToBeDistributed,
		Error:                 err,
	}
	l.callLogs = append(l.callLogs, log)
	return rewardToBeDistributed, err
}

func (l *loggingGasTrackerKeeper) GetLeftOverRewardEntry(ctx sdk.Context, rewardAddress string) (gstTypes.LeftOverRewardEntry, error) {
	leftOverEntry, err := l.underlyingKeeper.GetLeftOverRewardEntry(ctx, rewardAddress)

	log := gasTrackerKeeperCallLog{
		MethodName:          "GetLeftOverRewardEntry",
		RewardAddress:       rewardAddress,
		LeftOverRewardEntry: leftOverEntry,
		Error:               err,
	}

	l.callLogs = append(l.callLogs, log)
	return leftOverEntry, err
}

func (l *loggingGasTrackerKeeper) GetPreviousBlockTrackingInfo(ctx sdk.Context) (gstTypes.BlockGasTracking, error) {
	blockGasTracking, err := l.underlyingKeeper.GetPreviousBlockTrackingInfo(ctx)

	log := gasTrackerKeeperCallLog{
		MethodName:       "GetPreviousBlockTrackingInfo",
		BlockGasTracking: blockGasTracking,
		Error:            err,
	}

	l.callLogs = append(l.callLogs, log)
	return blockGasTracking, err
}

func (l *loggingGasTrackerKeeper) MarkEndOfTheBlock(ctx sdk.Context) error {
	err := l.underlyingKeeper.MarkEndOfTheBlock(ctx)

	log := gasTrackerKeeperCallLog{
		MethodName: "MarkEndOfTheBlock",
		Error:      err,
	}

	l.callLogs = append(l.callLogs, log)
	return err
}

type loggingWASMQuerier struct {
	LastCallWithSmart bool
	RawRequest        []byte
	Key               []byte
	TimesInvoked      uint64
	ContractAddress   string
	WrapperRequest    *gstTypes.GasTrackingQueryRequestWrapper
}

func (l *loggingWASMQuerier) Reset() {
	l.LastCallWithSmart = false
	l.RawRequest = nil
	l.Key = nil
	l.ContractAddress = ""
	l.WrapperRequest = nil
	l.TimesInvoked = 0
}

func (l *loggingWASMQuerier) QuerySmart(ctx sdk.Context, contractAddr sdk.AccAddress, req []byte) ([]byte, error) {
	l.LastCallWithSmart = true
	l.RawRequest = req
	l.Key = nil
	l.TimesInvoked += 1
	l.ContractAddress = contractAddr.String()

	var requestWrapper gstTypes.GasTrackingQueryRequestWrapper
	err := json.Unmarshal(req, &requestWrapper)
	if err != nil || requestWrapper.MagicString != gstTypes.MagicString {
		return []byte{4}, nil
	}

	l.WrapperRequest = &requestWrapper

	resultWrapper := gstTypes.GasTrackingQueryResultWrapper{
		GasConsumed:   234,
		QueryResponse: []byte{5},
	}
	resp, err := json.Marshal(resultWrapper)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (l *loggingWASMQuerier) QueryRaw(ctx sdk.Context, contractAddr sdk.AccAddress, key []byte) []byte {
	l.LastCallWithSmart = false
	l.RawRequest = nil
	l.Key = key
	l.TimesInvoked += 1
	l.ContractAddress = contractAddr.String()
	l.WrapperRequest = nil
	return []byte{7}
}

func TestWASMQueryPlugin(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)
	loggingKeeper := loggingGasTrackerKeeper{underlyingKeeper: keeper}
	loggingQuerier := loggingWASMQuerier{}

	loggingGasMeter := LoggingGasMeter{underlyingGasMeter: sdk.NewInfiniteGasMeter()}
	ctx = ctx.WithGasMeter(&loggingGasMeter)

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	plugin := NewGasTrackingWASMQueryPlugin(&loggingKeeper, &loggingQuerier)

	wasmQuery := types.WasmQuery{
		Smart: &types.SmartQuery{
			ContractAddr: "",
			Msg:          []byte{1},
		},
		Raw: nil,
	}
	_, err := plugin(ctx, &wasmQuery)
	require.EqualError(
		t,
		err,
		sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, wasmQuery.Smart.ContractAddr).Error(),
		"Query should error due to invalid address",
	)
	require.Equal(t, loggingQuerier.TimesInvoked, uint64(0))
	loggingQuerier.Reset()

	// We are not in a tx, so there should not be any tracking call happening.
	query := types.SmartQuery{
		ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		Msg:          []byte{1},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}
	_, err = plugin(ctx, &wasmQuery)
	require.NoError(
		t,
		err,
		"Query should succeed",
	)

	require.Equal(t, len(loggingKeeper.callLogs), 1)
	filteredLogs := loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, len(filteredLogs), 1)
	require.Equal(t, gstTypes.ErrBlockTrackingDataNotFound, filteredLogs[0].Error)

	require.Equal(t, loggingQuerier.TimesInvoked, uint64(1))
	require.Equal(t, loggingQuerier.LastCallWithSmart, true)
	require.Equal(t, loggingQuerier.RawRequest, query.Msg)
	require.Equal(t, loggingQuerier.ContractAddress, query.ContractAddr)

	gasMeterLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas")
	require.Equal(t, 0, len(gasMeterLogs))

	gasMeterLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	require.Equal(t, 0, len(gasMeterLogs))

	loggingQuerier.Reset()
	loggingKeeper.ResetLogs()
	loggingGasMeter.log = nil

	err = keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{})
	require.NoError(t, err, "Tracking new block should succeed")

	testDecCoin := sdk.NewDecCoin("test", sdk.NewInt(1))
	currentFee := []*sdk.DecCoin{&testDecCoin}
	currentGasLimit := uint64(5)
	err = keeper.TrackNewTx(ctx, currentFee, currentGasLimit)
	require.NoError(t, err, "Tracking new tx should succeed")

	contractMetadata := gstTypes.ContractInstanceMetadata{
		RewardAddress:            "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk",
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 50,
	}
	err = keeper.AddNewContractMetadata(ctx, "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt", contractMetadata)
	require.NoError(t, err, "Adding new contract metadata should succeed")

	query = types.SmartQuery{
		ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		Msg:          []byte{1},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}
	_, err = plugin(ctx, &wasmQuery)
	require.NoError(
		t,
		err,
		"Query should succeed",
	)

	require.Equal(t, loggingQuerier.TimesInvoked, uint64(1))
	require.Equal(t, loggingQuerier.LastCallWithSmart, true)
	require.Equal(t, loggingQuerier.ContractAddress, query.ContractAddr)
	require.Nil(t, loggingQuerier.Key)
	require.Equal(t, loggingQuerier.WrapperRequest, &gstTypes.GasTrackingQueryRequestWrapper{
		MagicString:  gstTypes.MagicString,
		QueryRequest: []byte{1},
	})

	require.Equal(t, 3, len(loggingKeeper.callLogs))

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, contractMetadata, filteredLogs[0].Metadata)

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, currentGasLimit, filteredLogs[0].TransactionTracking.MaxGasAllowed)
	require.Equal(t, currentFee, filteredLogs[0].TransactionTracking.MaxContractRewards)

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, uint64(234), filteredLogs[0].GasUsed)
	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_QUERY, filteredLogs[0].Operation)
	require.Equal(t, false, filteredLogs[0].IsEligibleForReward)

	gasMeterLogs = loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	require.Equal(t, 1, len(gasMeterLogs))
	require.Equal(t, fmt.Sprint(234), gasMeterLogs[0].InputArg[0])

	gasMeterLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	require.Equal(t, 0, len(gasMeterLogs))

	loggingKeeper.ResetLogs()
	loggingQuerier.Reset()
	loggingGasMeter.log = nil

	contractMetadata = gstTypes.ContractInstanceMetadata{
		RewardAddress:            "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
	}
	err = keeper.AddNewContractMetadata(ctx, "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk", contractMetadata)
	require.NoError(t, err, "Adding new contract metadata should succeed")

	query = types.SmartQuery{
		ContractAddr: "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk",
		Msg:          []byte{5},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}
	_, err = plugin(ctx, &wasmQuery)
	require.NoError(
		t,
		err,
		"Query should succeed",
	)

	require.Equal(t, loggingQuerier.TimesInvoked, uint64(1))
	require.Equal(t, loggingQuerier.LastCallWithSmart, true)
	require.Equal(t, loggingQuerier.ContractAddress, query.ContractAddr)
	require.Nil(t, loggingQuerier.Key)
	require.Equal(t, loggingQuerier.WrapperRequest, &gstTypes.GasTrackingQueryRequestWrapper{
		MagicString:  gstTypes.MagicString,
		QueryRequest: []byte{5},
	})

	require.Equal(t, 3, len(loggingKeeper.callLogs))

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, contractMetadata, filteredLogs[0].Metadata)

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("GetCurrentTxTrackingInfo")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, currentGasLimit, filteredLogs[0].TransactionTracking.MaxGasAllowed)
	require.Equal(t, currentFee, filteredLogs[0].TransactionTracking.MaxContractRewards)

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, uint64(234), filteredLogs[0].GasUsed)
	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_QUERY, filteredLogs[0].Operation)
	require.Equal(t, true, filteredLogs[0].IsEligibleForReward)

	gasMeterLogs = loggingGasMeter.log.FilterLogWithMethodName("RefundGas")
	require.Equal(t, 0, len(gasMeterLogs))

	gasMeterLogs = loggingGasMeter.log.FilterLogWithMethodName("ConsumeGas").FilterLogWithDescriptor(gstTypes.PremiumGasDescriptor)
	require.Equal(t, 1, len(gasMeterLogs))
	require.Equal(t, fmt.Sprint((234*contractMetadata.PremiumPercentageCharged)/100), gasMeterLogs[0].InputArg[0])
}
