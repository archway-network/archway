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

// TODO: satisfy GasTrackingKeeper interface
func (l *loggingGasTrackerKeeper) SetParams(ctx sdk.Context, params gstTypes.Params) {
	l.underlyingKeeper.SetParams(ctx, params)
}

func (l *loggingGasTrackerKeeper) IsGasTrackingEnabled(ctx sdk.Context) (res bool) {
	return l.underlyingKeeper.IsGasTrackingEnabled(ctx)
}

func (l *loggingGasTrackerKeeper) IsDappInflationRewardsEnabled(ctx sdk.Context) (res bool) {
	return l.underlyingKeeper.IsDappInflationRewardsEnabled(ctx)
}
func (l *loggingGasTrackerKeeper) IsGasRebateEnabled(ctx sdk.Context) (res bool) {
	return l.underlyingKeeper.IsGasRebateEnabled(ctx)
}
func (l *loggingGasTrackerKeeper) IsGasRebateToUserEnabled(ctx sdk.Context) bool {
	return l.underlyingKeeper.IsGasRebateToUserEnabled(ctx)
}
func (l *loggingGasTrackerKeeper) IsContractPremiumEnabled(ctx sdk.Context) (res bool) {
	return l.underlyingKeeper.IsContractPremiumEnabled(ctx)
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

type loggingWASMQuerier struct {
	LastCallWithSmart    bool
	RawRequest           []byte
	GiveInvalidSmartResp bool
	Key                  []byte
	TimesInvoked         uint64
	ContractAddress      string
	WrapperRequest       *gstTypes.GasTrackingQueryRequestWrapper
}

func (l *loggingWASMQuerier) Reset() {
	l.LastCallWithSmart = false
	l.RawRequest = nil
	l.Key = nil
	l.ContractAddress = ""
	l.WrapperRequest = nil
	l.TimesInvoked = 0
	l.GiveInvalidSmartResp = false
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

	if l.GiveInvalidSmartResp {
		return []byte{1}, nil
	}

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

type queryHandlerTestParams struct {
	ctx             sdk.Context
	loggingGasMeter *LoggingGasMeter
	loggingQuerier  *loggingWASMQuerier
	loggingKeeper   *loggingGasTrackerKeeper
}

func setupQueryHandlerTest(t *testing.T, ctx sdk.Context, keeper GasTrackingKeeper) queryHandlerTestParams {
	l := LoggingGasMeter{
		underlyingGasMeter: sdk.NewInfiniteGasMeter(),
		log:                nil,
	}
	loggingKeeper := loggingGasTrackerKeeper{underlyingKeeper: keeper}
	ctx = ctx.WithGasMeter(&l)

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	return queryHandlerTestParams{
		ctx:             ctx,
		loggingGasMeter: &l,
		loggingKeeper:   &loggingKeeper,
		loggingQuerier:  &loggingWASMQuerier{},
	}
}

// Test raw query handling
func TestWASMQueryPluginRaw(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)
	params := setupQueryHandlerTest(t, ctx, keeper)
	ctx = params.ctx
	loggingKeeper := params.loggingKeeper
	loggingQuerier := params.loggingQuerier

	plugin := NewGasTrackingWASMQueryPlugin(loggingKeeper, loggingQuerier)

	wasmQuery := types.WasmQuery{
		Raw: &types.RawQuery{
			ContractAddr: "",
			Key:          []byte{1},
		},
	}
	_, err := plugin(ctx, &wasmQuery)
	require.EqualError(t, err, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, wasmQuery.Raw.ContractAddr).Error(), "Raw query should return an error")

	wasmQuery = types.WasmQuery{
		Raw: &types.RawQuery{
			ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
			Key:          []byte{1},
		},
	}
	_, err = plugin(ctx, &wasmQuery)
	require.NoError(t, err, "Raw query should succeed")

	require.Equal(t, loggingQuerier.TimesInvoked, uint64(1))
	require.Equal(t, loggingQuerier.LastCallWithSmart, false)
	require.Equal(t, loggingQuerier.Key, wasmQuery.Raw.Key)
	require.Equal(t, loggingQuerier.ContractAddress, wasmQuery.Raw.ContractAddr)
}

// Test smart query handling
func TestWASMQueryPluginSmart(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)
	params := setupQueryHandlerTest(t, ctx, keeper)
	ctx = params.ctx
	loggingKeeper := params.loggingKeeper
	loggingQuerier := params.loggingQuerier
	loggingGasMeter := params.loggingGasMeter

	plugin := NewGasTrackingWASMQueryPlugin(loggingKeeper, loggingQuerier)

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
	require.Equal(t, uint64(0), loggingQuerier.TimesInvoked)
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

	// Without contract metadata there should be an error
	query := types.SmartQuery{
		ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		Msg:          []byte{1},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}

	_, err = plugin(ctx, &wasmQuery)
	require.EqualError(
		t,
		err,
		gstTypes.ErrContractInstanceMetadataNotFound.Error(),
		"Query should error due to invalid address",
	)

	loggingKeeper.ResetLogs()
	loggingQuerier.Reset()
	loggingGasMeter.log = nil

	contractMetadata := gstTypes.ContractInstanceMetadata{
		RewardAddress:            "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk",
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 50,
	}
	err = keeper.AddNewContractMetadata(ctx, "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt", contractMetadata)
	require.NoError(t, err, "Adding new contract metadata should succeed")

	loggingQuerier.GiveInvalidSmartResp = true
	query = types.SmartQuery{
		ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		Msg:          []byte{1},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}
	_, err = plugin(ctx, &wasmQuery)
	require.EqualError(
		t,
		err,
		"invalid character '\\x01' looking for beginning of value",
		"Query should fail",
	)

	loggingKeeper.ResetLogs()
	loggingQuerier.Reset()
	loggingGasMeter.log = nil

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

	require.Equal(t, 2, len(loggingKeeper.callLogs))

	filteredLogs := loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, contractMetadata, filteredLogs[0].Metadata)

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, uint64(234), filteredLogs[0].GasUsed)
	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_QUERY, filteredLogs[0].Operation)
	require.Equal(t, false, filteredLogs[0].IsEligibleForReward)

	gasMeterLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
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

	require.Equal(t, 2, len(loggingKeeper.callLogs))

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, contractMetadata, filteredLogs[0].Metadata)

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

// Test of outermost conditions of wasm query plugin
func TestWASMQueryPlugin(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)
	params := setupQueryHandlerTest(t, ctx, keeper)
	ctx = params.ctx
	loggingKeeper := params.loggingKeeper
	loggingQuerier := params.loggingQuerier

	plugin := NewGasTrackingWASMQueryPlugin(loggingKeeper, loggingQuerier)

	wasmQuery := types.WasmQuery{}
	_, err := plugin(ctx, &wasmQuery)
	require.EqualError(t, err, types.UnsupportedRequest{Kind: "unknown WasmQuery variant"}.Error(), "Plugin should return an error")

	wasmQuery = types.WasmQuery{
		Smart: &types.SmartQuery{
			ContractAddr: "",
			Msg:          []byte{1},
		},
		Raw: &types.RawQuery{
			ContractAddr: "",
			Key:          []byte{1},
		},
	}
	_, err = plugin(ctx, &wasmQuery)
	require.EqualError(t, err, types.UnsupportedRequest{Kind: "only one WasmQuery variant can be replied to"}.Error(), "Plugin should return an error")

}

// Test of outermost conditions of wasm query plugin
func TestWASMQueryPluginSmartWithoutGasRebateToUser(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)
	keeperParams := disableGasRebateToUser(gstTypes.DefaultParams())
	keeper.SetParams(ctx, keeperParams)
	params := setupQueryHandlerTest(t, ctx, keeper)
	ctx = params.ctx
	loggingKeeper := params.loggingKeeper
	loggingQuerier := params.loggingQuerier
	loggingGasMeter := params.loggingGasMeter

	plugin := NewGasTrackingWASMQueryPlugin(loggingKeeper, loggingQuerier)

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
	require.Equal(t, uint64(0), loggingQuerier.TimesInvoked)
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

	// Without contract metadata there should be an error
	query := types.SmartQuery{
		ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		Msg:          []byte{1},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}
	_, err = plugin(ctx, &wasmQuery)
	require.EqualError(
		t,
		err,
		gstTypes.ErrContractInstanceMetadataNotFound.Error(),
		"Query should error due to invalid address",
	)

	loggingKeeper.ResetLogs()
	loggingQuerier.Reset()
	loggingGasMeter.log = nil

	contractMetadata := gstTypes.ContractInstanceMetadata{
		RewardAddress:            "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk",
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 50,
	}
	err = keeper.AddNewContractMetadata(ctx, "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt", contractMetadata)
	require.NoError(t, err, "Adding new contract metadata should succeed")

	loggingQuerier.GiveInvalidSmartResp = true
	query = types.SmartQuery{
		ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		Msg:          []byte{1},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}
	_, err = plugin(ctx, &wasmQuery)
	require.EqualError(
		t,
		err,
		"invalid character '\\x01' looking for beginning of value",
		"Query should fail",
	)

	loggingKeeper.ResetLogs()
	loggingQuerier.Reset()
	loggingGasMeter.log = nil

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

	require.Equal(t, 2, len(loggingKeeper.callLogs))

	filteredLogs := loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, contractMetadata, filteredLogs[0].Metadata)

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, uint64(234), filteredLogs[0].GasUsed)
	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_QUERY, filteredLogs[0].Operation)
	require.Equal(t, false, filteredLogs[0].IsEligibleForReward)

	gasMeterLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	require.Equal(t, 0, len(gasMeterLogs))

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

	require.Equal(t, 2, len(loggingKeeper.callLogs))

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, contractMetadata, filteredLogs[0].Metadata)

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

func TestWASMQueryPluginSmartWithoutContractPremium(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)
	keeperParams := disableContractPremium(gstTypes.DefaultParams())
	keeper.SetParams(ctx, keeperParams)

	params := setupQueryHandlerTest(t, ctx, keeper)
	ctx = params.ctx
	loggingKeeper := params.loggingKeeper
	loggingQuerier := params.loggingQuerier
	loggingGasMeter := params.loggingGasMeter

	plugin := NewGasTrackingWASMQueryPlugin(loggingKeeper, loggingQuerier)

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
	require.Equal(t, uint64(0), loggingQuerier.TimesInvoked)
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

	// Without contract metadata there should be an error
	query := types.SmartQuery{
		ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		Msg:          []byte{1},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}
	_, err = plugin(ctx, &wasmQuery)
	require.EqualError(
		t,
		err,
		gstTypes.ErrContractInstanceMetadataNotFound.Error(),
		"Query should error due to invalid address",
	)

	loggingKeeper.ResetLogs()
	loggingQuerier.Reset()
	loggingGasMeter.log = nil

	contractMetadata := gstTypes.ContractInstanceMetadata{
		RewardAddress:            "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk",
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 50,
	}
	err = keeper.AddNewContractMetadata(ctx, "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt", contractMetadata)
	require.NoError(t, err, "Adding new contract metadata should succeed")

	loggingQuerier.GiveInvalidSmartResp = true
	query = types.SmartQuery{
		ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		Msg:          []byte{1},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}
	_, err = plugin(ctx, &wasmQuery)
	require.EqualError(
		t,
		err,
		"invalid character '\\x01' looking for beginning of value",
		"Query should fail",
	)

	loggingKeeper.ResetLogs()
	loggingQuerier.Reset()
	loggingGasMeter.log = nil

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

	require.Equal(t, 2, len(loggingKeeper.callLogs))

	filteredLogs := loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, contractMetadata, filteredLogs[0].Metadata)

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, uint64(234), filteredLogs[0].GasUsed)
	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_QUERY, filteredLogs[0].Operation)
	require.Equal(t, false, filteredLogs[0].IsEligibleForReward)

	gasMeterLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	require.Equal(t, 1, len(gasMeterLogs))

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

	require.Equal(t, 2, len(loggingKeeper.callLogs))

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, contractMetadata, filteredLogs[0].Metadata)

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
	require.Equal(t, 0, len(gasMeterLogs))
}

func TestWASMQueryPluginSmartWithoutContractPremiumOrGasRebateToUser(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)
	keeperParams := disableGasRebateToUser(disableContractPremium(gstTypes.DefaultParams()))
	keeper.SetParams(ctx, keeperParams)

	params := setupQueryHandlerTest(t, ctx, keeper)
	ctx = params.ctx
	loggingKeeper := params.loggingKeeper
	loggingQuerier := params.loggingQuerier
	loggingGasMeter := params.loggingGasMeter

	plugin := NewGasTrackingWASMQueryPlugin(loggingKeeper, loggingQuerier)

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
	require.Equal(t, uint64(0), loggingQuerier.TimesInvoked)
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

	// Without contract metadata there should be an error
	query := types.SmartQuery{
		ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		Msg:          []byte{1},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}
	_, err = plugin(ctx, &wasmQuery)
	require.EqualError(
		t,
		err,
		gstTypes.ErrContractInstanceMetadataNotFound.Error(),
		"Query should error due to invalid address",
	)

	loggingKeeper.ResetLogs()
	loggingQuerier.Reset()
	loggingGasMeter.log = nil

	contractMetadata := gstTypes.ContractInstanceMetadata{
		RewardAddress:            "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk",
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 50,
	}
	err = keeper.AddNewContractMetadata(ctx, "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt", contractMetadata)
	require.NoError(t, err, "Adding new contract metadata should succeed")

	loggingQuerier.GiveInvalidSmartResp = true
	query = types.SmartQuery{
		ContractAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		Msg:          []byte{1},
	}
	wasmQuery = types.WasmQuery{
		Smart: &query,
		Raw:   nil,
	}
	_, err = plugin(ctx, &wasmQuery)
	require.EqualError(
		t,
		err,
		"invalid character '\\x01' looking for beginning of value",
		"Query should fail",
	)

	loggingKeeper.ResetLogs()
	loggingQuerier.Reset()
	loggingGasMeter.log = nil

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

	require.Equal(t, 2, len(loggingKeeper.callLogs))

	filteredLogs := loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, contractMetadata, filteredLogs[0].Metadata)

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("TrackContractGasUsage")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, uint64(234), filteredLogs[0].GasUsed)
	require.Equal(t, gstTypes.ContractOperation_CONTRACT_OPERATION_QUERY, filteredLogs[0].Operation)
	require.Equal(t, false, filteredLogs[0].IsEligibleForReward)

	gasMeterLogs := loggingGasMeter.log.FilterLogWithMethodName("RefundGas").FilterLogWithDescriptor(gstTypes.GasRebateToUserDescriptor)
	require.Equal(t, 0, len(gasMeterLogs))

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

	require.Equal(t, 2, len(loggingKeeper.callLogs))

	filteredLogs = loggingKeeper.callLogs.FilterByMethod("GetNewContractMetadata")
	require.Equal(t, 1, len(filteredLogs))
	require.Equal(t, nil, filteredLogs[0].Error)
	require.Equal(t, query.ContractAddr, filteredLogs[0].ContractAddress)
	require.Equal(t, contractMetadata, filteredLogs[0].Metadata)

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
	require.Equal(t, 0, len(gasMeterLogs))
}
