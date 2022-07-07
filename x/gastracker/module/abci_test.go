package module

import (
	"fmt"
	"github.com/gogo/protobuf/proto"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
	tmLog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	db "github.com/tendermint/tm-db"

	gastracker "github.com/archway-network/archway/x/gastracker"
	gstTypes "github.com/archway-network/archway/x/gastracker"
	keeper "github.com/archway-network/archway/x/gastracker/keeper"
)

type Behaviour int

const (
	Log Behaviour = iota
	Error
	Panic
)

// NOTE: this is needed to allow the keeper to set BlockGasTracking
var (
	storeKey = sdk.NewKVStoreKey(gastracker.StoreKey)
)

type RewardTransferKeeperCallLogs struct {
	Method          string
	senderModule    string
	recipientModule string
	recipientAddr   string
	amt             sdk.Coins
}

type TestRewardTransferKeeper struct {
	Logs []*RewardTransferKeeperCallLogs
	B    Behaviour
}

func (t *TestRewardTransferKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	switch t.B {
	case Log:
		t.Logs = append(t.Logs, &RewardTransferKeeperCallLogs{
			Method:        "SendCoinsFromModuleToAccount",
			senderModule:  senderModule,
			recipientAddr: recipientAddr.String(),
			amt:           amt,
		})
	case Error:
		return fmt.Errorf("TestError")
	case Panic:
		panic("TestPanic")
	}
	return nil
}

func (t *TestRewardTransferKeeper) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	switch t.B {
	case Log:
		t.Logs = append(t.Logs, &RewardTransferKeeperCallLogs{
			Method:          "SendCoinsFromModuleToModule",
			senderModule:    senderModule,
			recipientModule: recipientModule,
			amt:             amt,
		})
	case Error:
		return fmt.Errorf("TestError")
	case Panic:
		panic("TestPanic")
	}
	return nil
}

type TestMintParamsKeeper struct {
	B                Behaviour
	BlocksPerYear    uint64
	AnnualProvisions sdk.Dec
}

func (t *TestMintParamsKeeper) GetParams(_ sdk.Context) (params mintTypes.Params) {
	if t.B == Panic {
		panic("TestPanic")
	}
	p := mintTypes.Params{
		MintDenom:     "test",
		BlocksPerYear: 100,
	}

	if t.BlocksPerYear != 0 {
		p.BlocksPerYear = t.BlocksPerYear
	}

	return p
}

func (t *TestMintParamsKeeper) GetMinter(_ sdk.Context) (minter mintTypes.Minter) {
	if t.B == Panic {
		panic("TestPanic")
	}

	m := mintTypes.Minter{
		AnnualProvisions: sdk.NewDec(76500),
	}

	if !t.AnnualProvisions.IsNil() {
		m.AnnualProvisions = t.AnnualProvisions
	}

	return m
}
func createLogModule(module string, mod string, coins sdk.Coins) *RewardTransferKeeperCallLogs {
	return &RewardTransferKeeperCallLogs{
		Method:          "SendCoinsFromModuleToModule",
		senderModule:    module,
		recipientModule: mod,
		amt:             coins,
	}
}
func createLogAddr(module string, addr string, coins sdk.Coins) *RewardTransferKeeperCallLogs {
	return &RewardTransferKeeperCallLogs{
		Method:        "SendCoinsFromModuleToAccount",
		senderModule:  module,
		recipientAddr: addr,
		amt:           coins,
	}
}
func disableGasTracking(params gstTypes.Params) gstTypes.Params {
	params.GasTrackingSwitch = false
	return params
}
func disableContractPremium(params gstTypes.Params) gstTypes.Params {
	params.ContractPremiumSwitch = false
	return params
}

func setInflationRewardCap(params gstTypes.Params, percentage uint64) gstTypes.Params {
	params.InflationRewardCapPercentage = percentage
	return params
}

func enableInflationRewardCap(params gstTypes.Params) gstTypes.Params {
	params.InflationRewardCapSwitch = true
	return params
}

func disableDappInflation(params gstTypes.Params) gstTypes.Params {
	params.DappInflationRewardsSwitch = false
	return params
}
func disableGasRebate(params gstTypes.Params) gstTypes.Params {
	params.GasRebateSwitch = false
	return params
}
func disableGasRebateToUser(params gstTypes.Params) gstTypes.Params {
	params.GasRebateToUserSwitch = false
	return params
}

// Test the conditions under which BeginBlocker and EndBlocker should panic or not panic
func TestABCIPanicBehaviour(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t, sdk.AccAddress{})

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	ctx = ctx.WithBlockHeight(0)

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)
	require.PanicsWithError(t, gstTypes.ErrBlockTrackingDataNotFound.Error(), func() {
		BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	}, "BeginBlock should panic")

	ctx = ctx.WithBlockHeight(1)

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// We should not have made any call to reward keeper
	require.Zero(t, testRewardKeeper.Logs, "No logs should be there as no need to make new calls")
	// We would have overwritten the TrackNewBlock obj
	blockGasTracking, err := keeper.GetCurrentBlockTracking(ctx)
	require.NoError(t, err, "We should be able to get new block gas tracking")
	require.Equal(t, gstTypes.BlockGasTracking{}, blockGasTracking, "We should have overwritten block gas tracking obj")
}

func TestABCIContractMetadataCommit(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])

	contractMetadatas := []gstTypes.ContractInstanceMetadata{
		{
			RewardAddress:    spareAddress[5].String(),
			GasRebateToUser:  false,
			DeveloperAddress: spareAddress[0].String(),
		},
		{
			RewardAddress:    spareAddress[6].String(),
			GasRebateToUser:  false,
			DeveloperAddress: spareAddress[0].String(),
		},
		{
			RewardAddress:            spareAddress[6].String(),
			GasRebateToUser:          false,
			CollectPremium:           true,
			PremiumPercentageCharged: 50,
			DeveloperAddress:         spareAddress[0].String(),
		},
		{
			RewardAddress:            spareAddress[5].String(),
			GasRebateToUser:          false,
			CollectPremium:           true,
			PremiumPercentageCharged: 150,
			DeveloperAddress:         spareAddress[0].String(),
		},
	}

	for i := 1; i <= 4; i++ {
		err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[i], contractMetadatas[i-1])
		require.NoError(t, err, "We should be able to add new contract metadata")
	}

	ctx = ctx.WithBlockHeight(1)

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)

	// Now all pending metadata should have been committed
	for i := 1; i <= 4; i++ {
		_, err := keeper.GetPendingContractMetadataChange(ctx, spareAddress[i])
		require.EqualError(t, err, gstTypes.ErrContractInstanceMetadataNotFound.Error(), "We should not be able to get pending metadata")

		retrievedMetadata, err := keeper.GetContractMetadata(ctx, spareAddress[i])
		require.NoError(t, err, "We should be able to get committed metadata")
		require.Equal(t, contractMetadatas[i-1], retrievedMetadata, "Retrieved metadata should be same")
	}
}

// Test reward calculation

// Inflation reward cap for a block is hardcoded at 20% of total inflation reward
// So, for per block inflation of 765 (76500/100), we need to take 20% of it which is
// 153.
// Total gas across block is 20 (4+8+2+4+2)
// So, inflation reward for
// Contract "1" is: 0.00306 (153 * 4 / 200000)
// Contract "2" is: 0.00612 (153 * 8 / 200000)
// Contract "3" is: 0.00153 (153 * 2 / 200000)
// Contract "2" is: 0.00306 (153 * 4 / 200000)
// Contract "4" is: 0.00153 (153 * 2 / 200000)
// All above is in "test" denomination, since that is the denomination minter is minting
// Now, coming to gas reward calculations:
// For First tx entry:
//    Gas Used = 20
//    "1" Contract's reward is: 1 * (4 / 20) = 0.2test and 0.33333 * (2 / 20) = 0.0666666test1
//    "2" Contract's reward is: 1 * (8 / 20) = 0.4test and 0.33333 * (4 / 20) = 0.1333333test1
//    "3" Contract's reward is: 1 * (2 / 20) = 0.15test (0.1test + 0.05test (Premium)) and 0.33333 * (2 / 20) = 0.0499995test1 (0.033333test1 + 0.0166665test1 (premium))
// For Second tx entry:
//   Gas Used = 4
//   "2" Contract's reward is: 2 * (4 / 4) = 2test and 0.5 * (4 / 4) = 0.5test1
// Total rewards:
// For Contract "1": 0.20306test (0.00306 + 0.2) and 0.0666666test1
// For Contract "2": 2.40918test (0.00612 + 0.00306 + 0.4 + 2) and 0.6333333test1 (0.1333333 + 0.5)
// For Contract "3": 0.15153test (0.00153 + 0.15) and 0.04999995test1
// For Contract "4": 0.00153test (0.00153)
// Reward distribution per address:
// (for contract "1" and "4") "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt": 0.20306test + 0.00153test (0.20459test) and 0.0666666test1
// (for contract "2" and "3") "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk": 2.56071test  and 0.6666666test1
// So, we should be fetching 3test (0.202295 + 2.555296 = 2.757591 rounded to 3) and 1test1 (0.6333333 + 0.04999995 =  0.68333325 rounded to 1)
// from the fee collector
// Since, left over threshold is hard coded to 1, we should be transferring 0test to "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt" and
// 2test to "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk" and left over rewards should be 0.202295test,0.0666666test1 and 0.555355test and 0.68333325test1
// respectively.
func TestRewardCalculation(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := gstTypes.DefaultParams()
	expected := expect{
		rewardsA: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.20459")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.56071")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.683333333333333333")),
		},
		logs: []*RewardTransferKeeperCallLogs{
			createLogModule(authTypes.FeeCollectorName, gstTypes.ContractRewardCollector, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(3)), sdk.NewCoin("test1", sdk.NewInt(1)))),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[5].String(), sdk.NewCoins()),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[6].String(), sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(2)))),
		},
	}

	expectedRewardCalculationEventFromFirstTx := []gstTypes.ContractRewardCalculationEvent{
		{
			ContractAddress:  spareAddress[1].String(),
			GasConsumed:      4,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00306"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.20306")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")))),
		},
		{
			ContractAddress:  spareAddress[2].String(),
			GasConsumed:      8,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00612"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.40612")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.133333333333333333")))),
		},
		{
			ContractAddress:  spareAddress[3].String(),
			GasConsumed:      2,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00153"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.15153")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.05")))),
		},
		{
			ContractAddress:  spareAddress[4].String(),
			GasConsumed:      2,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00153"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00153")))),
		},
	}

	firstTxTotalRewards := CalculateTotalRewardFromEvents(expectedRewardCalculationEventFromFirstTx)

	expectedRewardCalculationEventsFromSecondTx := []gstTypes.ContractRewardCalculationEvent{
		{
			ContractAddress:  spareAddress[2].String(),
			GasConsumed:      2,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00306"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("2.00306")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.5")))),
		},
	}

	secondTxTotalRewards := CalculateTotalRewardFromEvents(expectedRewardCalculationEventsFromSecondTx)

	expectedRewardCalculationEvents := append(expectedRewardCalculationEventFromFirstTx, expectedRewardCalculationEventsFromSecondTx...)

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200000))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	remainingFeeFirstTx := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(2).QuoInt64(3)))
	remainingFeeSecondTx := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(2).QuoInt64(3)))

	totalFeeFirstTx := firstTxMaxContractReward.Add(remainingFeeFirstTx...)
	totalFeeSecondTx := secondTxMaxContractReward.Add(remainingFeeSecondTx...)

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[6].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[6].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[4], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 150,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 4, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      20,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstTx[0], &remainingFeeFirstTx[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 6,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      4,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeSecondTx[0], &remainingFeeSecondTx[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsA), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsA); i++ {
		require.Equal(t, expected.rewardsA[i], *leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], *leftOverEntry.ContractRewards[i])
	}

	beginBlockEvents := ctx.EventManager().Events()
	eventType := proto.MessageName(&gstTypes.ContractRewardCalculationEvent{})
	currentIndexToValidate := 0
	for _, event := range beginBlockEvents {
		if event.Type == eventType {
			generatedEvent, err := sdk.TypedEventToEvent(&expectedRewardCalculationEvents[currentIndexToValidate])
			require.NoError(t, err, "should not be an error in event generation")
			currentIndexToValidate += 1

			generatedAttributeValue, err := FindAttributeFromEvent(generatedEvent, "contract_rewards")
			require.NoError(t, err, "should not be an error in finding attribute")

			actualAttributeValue, err := FindAttributeFromEvent(event, "contract_rewards")
			require.NoError(t, err, "should not be an error in finding attribute")

			require.Equal(t, generatedAttributeValue, actualAttributeValue, "generated attribute and actual attribute of event should be equal")
		}
	}
	require.Equal(t, currentIndexToValidate, len(expectedRewardCalculationEvents), "BeginBlock should have generated all events")

	// Now that we have verified that events are as expected, we can check the total reward per tx is less than the fee paid
	rewardDiff, isNegative := totalFeeFirstTx.SafeSub(firstTxTotalRewards)
	require.True(t, rewardDiff.IsAllPositive() && !isNegative, "reward difference must be positive")

	rewardDiff, isNegative = totalFeeSecondTx.SafeSub(secondTxTotalRewards)
	require.True(t, rewardDiff.IsAllPositive() && !isNegative, "reward difference must be positive")
}

// In a real world scenario where without safety net we would be giving more rewards
// Reference:
// Cosmos total supply: 260,906,513 * 10^6 uatom = 260906513000000 uatom
// Blocks per year: 4,360,000
// Inflation min: 7%
// Inflation per year would be: 18263455910000 uatom
// Inflation per block is: 4188866 uatom
// 20% of that inflation is: 837773 uatom
// Avg Transaction fee is: 2000 uatom
// Block gas limit is: 2000000 gas
// If we have a tx with 400000 gas, that would mean it would get 20% of 837773 atom which equals to: 167554.64 uatom
// For simplicity of calculation let's remove last two digits from everything and round them
// Inflation per block is: 4200
// 20% of that inflation is: 840
// Avg transaction fee is: 20
// Block gas limit is: 2000000 gas
// if we have a tx with 400000 gas, that would mean it would get 20% of 840 which equals to: 168 which is greater than fee of 20
func TestContractRewardsWithoutCapWithRealWorldParams(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := gstTypes.DefaultParams()
	expected := expect{
		rewardsA: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.333333333333333333")),
		},
		logs: []*RewardTransferKeeperCallLogs{
			createLogModule(authTypes.FeeCollectorName, gstTypes.ContractRewardCollector, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(169)), sdk.NewCoin("test1", sdk.NewInt(1)))),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[5].String(), sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(169)))),
		},
	}

	expectedRewardCalculationEventFromFirstTx := []gstTypes.ContractRewardCalculationEvent{
		{
			ContractAddress:  spareAddress[1].String(),
			GasConsumed:      4,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("168"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("169")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.333333333333333333")))),
		},
	}

	firstTxTotalRewards := CalculateTotalRewardFromEvents(expectedRewardCalculationEventFromFirstTx)

	expectedRewardCalculationEvents := expectedRewardCalculationEventFromFirstTx

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))

	remainingFeeFirstTx := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(4).QuoInt64(3)))

	totalFeeFirstTx := firstTxMaxContractReward.Add(remainingFeeFirstTx...)

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}
	testMintParamsKeeper.AnnualProvisions = sdk.NewDec(4360000 * 4200)
	testMintParamsKeeper.BlocksPerYear = 4360000

	ctx = ctx.WithBlockHeight(2)
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(2000000))

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 1, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      400000,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstTx[0], &remainingFeeFirstTx[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  200000,
					OriginalSdkGas: 200000,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsA), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsA); i++ {
		require.Equal(t, expected.rewardsA[i], *leftOverEntry.ContractRewards[i])
	}

	beginBlockEvents := ctx.EventManager().Events()
	eventType := proto.MessageName(&gstTypes.ContractRewardCalculationEvent{})
	currentIndexToValidate := 0
	for _, event := range beginBlockEvents {
		if event.Type == eventType {
			generatedEvent, err := sdk.TypedEventToEvent(&expectedRewardCalculationEvents[currentIndexToValidate])
			require.NoError(t, err, "should not be an error in event generation")
			currentIndexToValidate += 1

			generatedAttributeValue, err := FindAttributeFromEvent(generatedEvent, "contract_rewards")
			require.NoError(t, err, "should not be an error in finding attribute")

			actualAttributeValue, err := FindAttributeFromEvent(event, "contract_rewards")
			require.NoError(t, err, "should not be an error in finding attribute")

			require.Equal(t, generatedAttributeValue, actualAttributeValue, "generated attribute and actual attribute of event should be equal")
		}
	}
	require.Equal(t, currentIndexToValidate, len(expectedRewardCalculationEvents), "BeginBlock should have generated all events")

	// Now that we have verified that events are as expected, we can see that we are giving more rewards than fee paid
	rewardDiff, isNegative := totalFeeFirstTx.SafeSub(firstTxTotalRewards)
	require.True(t, !rewardDiff.IsAllPositive() && isNegative, "reward difference must be positive")
}

// In a real world scenario where without safety net we would be giving more rewards
// Reference:
// Cosmos total supply: 260,906,513 * 10^6 uatom = 260906513000000 uatom
// Blocks per year: 4,360,000
// Inflation min: 7%
// Inflation per year would be: 18263455910000 uatom
// Inflation per block is: 4188866 uatom
// 20% of that inflation is: 837773 uatom
// Avg Transaction fee is: 2000 uatom
// Block gas limit is: 2000000 gas
// If we have a tx with 400000 gas, that would mean it would get 20% of 837773 atom which equals to: 167554.64 uatom
// For simplicity of calculation let's remove last two digits from everything and round them
// Inflation per block is: 4200
// 20% of that inflation is: 840
// Avg transaction fee is: 20
// Block gas limit is: 2000000 gas
// if we have a tx with 400000 gas, that would mean it would get 20% of 840 which equals to: 168 which is greater than fee of 20
func TestContractRewardsWithCapAndRealWorldParams(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := setInflationRewardCap(enableInflationRewardCap(gstTypes.DefaultParams()), 90)
	expected := expect{
		rewardsA: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.8")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.333333333333333333")),
		},
		logs: []*RewardTransferKeeperCallLogs{
			createLogModule(authTypes.FeeCollectorName, gstTypes.ContractRewardCollector, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(3)), sdk.NewCoin("test1", sdk.NewInt(1)))),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[5].String(), sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(2)))),
		},
	}

	expectedRewardCalculationEventFromFirstTx := []gstTypes.ContractRewardCalculationEvent{
		{
			ContractAddress:  spareAddress[1].String(),
			GasConsumed:      4,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("1.8"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("2.8")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.333333333333333333")))),
		},
	}

	firstTxTotalRewards := CalculateTotalRewardFromEvents(expectedRewardCalculationEventFromFirstTx)

	expectedRewardCalculationEvents := expectedRewardCalculationEventFromFirstTx

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))

	remainingFeeFirstTx := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(4).QuoInt64(3)))

	totalFeeFirstTx := firstTxMaxContractReward.Add(remainingFeeFirstTx...)

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}
	testMintParamsKeeper.AnnualProvisions = sdk.NewDec(4360000 * 4200)
	testMintParamsKeeper.BlocksPerYear = 4360000

	ctx = ctx.WithBlockHeight(2)
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(2000000))

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 1, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      400000,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstTx[0], &remainingFeeFirstTx[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  200000,
					OriginalSdkGas: 200000,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsA), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsA); i++ {
		require.Equal(t, expected.rewardsA[i], *leftOverEntry.ContractRewards[i])
	}

	beginBlockEvents := ctx.EventManager().Events()
	eventType := proto.MessageName(&gstTypes.ContractRewardCalculationEvent{})
	currentIndexToValidate := 0
	for _, event := range beginBlockEvents {
		if event.Type == eventType {
			generatedEvent, err := sdk.TypedEventToEvent(&expectedRewardCalculationEvents[currentIndexToValidate])
			require.NoError(t, err, "should not be an error in event generation")
			currentIndexToValidate += 1

			generatedAttributeValue, err := FindAttributeFromEvent(generatedEvent, "contract_rewards")
			require.NoError(t, err, "should not be an error in finding attribute")

			actualAttributeValue, err := FindAttributeFromEvent(event, "contract_rewards")
			require.NoError(t, err, "should not be an error in finding attribute")

			require.Equal(t, generatedAttributeValue, actualAttributeValue, "generated attribute and actual attribute of event should be equal")
		}
	}
	require.Equal(t, currentIndexToValidate, len(expectedRewardCalculationEvents), "BeginBlock should have generated all events")

	// Now that we have verified that events are as expected, we can see that we are giving more rewards than fee paid
	rewardDiff, isNegative := totalFeeFirstTx.SafeSub(firstTxTotalRewards)
	require.True(t, rewardDiff.IsAllPositive() && !isNegative, "reward difference must be positive")
}

// In a scenario where with safety net we are always giving less reward than the fee paid
func TestContractRewardsWithDappInflationCapMoreThanUncapped(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := setInflationRewardCap(enableInflationRewardCap(gstTypes.DefaultParams()), 90)
	expected := expect{
		rewardsA: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.20459")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.56071")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.683333333333333333")),
		},
		logs: []*RewardTransferKeeperCallLogs{
			createLogModule(authTypes.FeeCollectorName, gstTypes.ContractRewardCollector, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(3)), sdk.NewCoin("test1", sdk.NewInt(1)))),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[5].String(), sdk.NewCoins()),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[6].String(), sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(2)))),
		},
	}

	expectedRewardCalculationEventFromFirstTx := []gstTypes.ContractRewardCalculationEvent{
		{
			ContractAddress:  spareAddress[1].String(),
			GasConsumed:      4,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00306"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.20306")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")))),
		},
		{
			ContractAddress:  spareAddress[2].String(),
			GasConsumed:      8,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00612"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.40612")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.133333333333333333")))),
		},
		{
			ContractAddress:  spareAddress[3].String(),
			GasConsumed:      2,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00153"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.15153")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.05")))),
		},
		{
			ContractAddress:  spareAddress[4].String(),
			GasConsumed:      2,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00153"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00153")))),
		},
	}

	firstTxTotalRewards := CalculateTotalRewardFromEvents(expectedRewardCalculationEventFromFirstTx)

	expectedRewardCalculationEventsFromSecondTx := []gstTypes.ContractRewardCalculationEvent{
		{
			ContractAddress:  spareAddress[2].String(),
			GasConsumed:      2,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00306"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("2.00306")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.5")))),
		},
	}

	secondTxTotalRewards := CalculateTotalRewardFromEvents(expectedRewardCalculationEventsFromSecondTx)

	expectedRewardCalculationEvents := append(expectedRewardCalculationEventFromFirstTx, expectedRewardCalculationEventsFromSecondTx...)

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200000))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	remainingFeeFirstTx := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))
	remainingFeeSecondTx := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))

	totalFeeFirstTx := firstTxMaxContractReward.Add(remainingFeeFirstTx...)
	totalFeeSecondTx := secondTxMaxContractReward.Add(remainingFeeSecondTx...)

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[6].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[6].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[4], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 150,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 4, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      20,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstTx[0], &remainingFeeFirstTx[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 6,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      4,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeSecondTx[0], &remainingFeeSecondTx[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsA), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsA); i++ {
		require.Equal(t, expected.rewardsA[i], *leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], *leftOverEntry.ContractRewards[i])
	}

	beginBlockEvents := ctx.EventManager().Events()
	eventType := proto.MessageName(&gstTypes.ContractRewardCalculationEvent{})
	currentIndexToValidate := 0
	for _, event := range beginBlockEvents {
		if event.Type == eventType {
			generatedEvent, err := sdk.TypedEventToEvent(&expectedRewardCalculationEvents[currentIndexToValidate])
			require.NoError(t, err, "should not be an error in event generation")
			currentIndexToValidate += 1

			generatedAttributeValue, err := FindAttributeFromEvent(generatedEvent, "contract_rewards")
			require.NoError(t, err, "should not be an error in finding attribute")

			actualAttributeValue, err := FindAttributeFromEvent(event, "contract_rewards")
			require.NoError(t, err, "should not be an error in finding attribute")

			require.Equal(t, generatedAttributeValue, actualAttributeValue, "generated attribute and actual attribute of event should be equal")
		}
	}
	require.Equal(t, currentIndexToValidate, len(expectedRewardCalculationEvents), "BeginBlock should have generated all events")

	// Now that we have verified that events are as expected, we can check the total reward per tx is less than the fee paid
	rewardDiff, isNegative := totalFeeFirstTx.SafeSub(firstTxTotalRewards)
	require.True(t, rewardDiff.IsAllPositive() && !isNegative, "reward difference must be positive")

	rewardDiff, isNegative = totalFeeSecondTx.SafeSub(secondTxTotalRewards)
	require.True(t, rewardDiff.IsAllPositive() && !isNegative, "reward difference must be positive")
}

// In a scenario where with safety net we are always giving less reward than the fee paid
// Inflation reward cap for a block is at 20% of total inflation reward
// So, for per block inflation of 765 (76500/100), we need to take 20% of it which is
// 153.
// Total gas across block is 20 (4+8+2+4+2)
// So, inflation reward for
// tx 1:
// remaining fee is: 2test, 90% of that is: 1.8test
// Contract "1" is: MIN ((153 * 4 / 200), (1.8*4/16)) = 0.45test
// Contract "2" is: MIN ((0.00612 (153 * 8 / 200), (1.8*8/16)) = 0.9test
// Contract "3" is: MIN ((153 * 2 / 200), (1.8*2/16)) = 0.225test
// Contract "4" is: MIN ((153 * 2 / 200), (1.8*2/16)) = 0.225test
// tx 2:
// remaining fee is: 4test, 90% of that is: 3.6test
// Contract "2" is:  MIN((153 * 4 / 200), (3.6*4/4)) = 3.06test
// All above is in "test" denomination, since that is the denomination minter is minting
// Now, coming to gas reward calculations:
// For First tx entry:
//    Gas Used = 20
//    "1" Contract's reward is: 1 * (4 / 20) = 0.2test and 0.33333 * (2 / 20) = 0.0666666test1
//    "2" Contract's reward is: 1 * (8 / 20) = 0.4test and 0.33333 * (4 / 20) = 0.1333333test1
//    "3" Contract's reward is: 1 * (2 / 20) = 0.15test (0.1test + 0.05test (Premium)) and 0.33333 * (2 / 20) = 0.0499995test1 (0.033333test1 + 0.0166665test1 (premium))
// For Second tx entry:
//   Gas Used = 4
//   "2" Contract's reward is: 2 * (4 / 4) = 2test and 0.5 * (4 / 4) = 0.5test1
// Total rewards:
// For Contract "1": 0.65test (0.45 + 0.2) and 0.0666666test1
// For Contract "2": 6.36test (0.9 + 3.06 + 0.4 + 2) and 0.6333333test1 (0.1333333 + 0.5)
// For Contract "3": 0.375test (0.225 + 0.15) and 0.04999995test1
// For Contract "4": 0.225test (0.225)
// Reward distribution per address:
// (for contract "1" and "4") "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt": 0.65test + 0.225test (0.875test) and 0.0666666test1
// (for contract "2" and "3") "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk": 6.735test  and 0.6666666test1
// So, we should be fetching 8test (0.875 + 6.735 = 7.61 rounded to 8) and 1test1 (0.6333333 + 0.04999995 =  0.68333325 rounded to 1)
// from the fee collector
// Since, left over threshold is hard coded to 1, we should be transferring 0test to "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt" and
// 6test to "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk" and left over rewards should be 0.875test,0.0666666test1 and 0.735test and 0.68333325test1
// respectively.
func TestContractRewardsWithDappInflationCapLessThanUncapped(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := setInflationRewardCap(enableInflationRewardCap(gstTypes.DefaultParams()), 90)
	expected := expect{
		rewardsA: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.875")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.735")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.683333333333333333")),
		},
		logs: []*RewardTransferKeeperCallLogs{
			createLogModule(authTypes.FeeCollectorName, gstTypes.ContractRewardCollector, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(8)), sdk.NewCoin("test1", sdk.NewInt(1)))),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[5].String(), sdk.NewCoins()),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[6].String(), sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(6)))),
		},
	}

	expectedRewardCalculationEventFromFirstTx := []gstTypes.ContractRewardCalculationEvent{
		{
			ContractAddress:  spareAddress[1].String(),
			GasConsumed:      4,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.45"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.65")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")))),
		},
		{
			ContractAddress:  spareAddress[2].String(),
			GasConsumed:      8,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.9"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("1.3")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.133333333333333333")))),
		},
		{
			ContractAddress:  spareAddress[3].String(),
			GasConsumed:      2,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.225"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.375")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.05")))),
		},
		{
			ContractAddress:  spareAddress[4].String(),
			GasConsumed:      2,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.225"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.225")))),
		},
	}

	firstTxTotalRewards := CalculateTotalRewardFromEvents(expectedRewardCalculationEventFromFirstTx)

	expectedRewardCalculationEventsFromSecondTx := []gstTypes.ContractRewardCalculationEvent{
		{
			ContractAddress:  spareAddress[2].String(),
			GasConsumed:      2,
			InflationRewards: ConvertToProtoDecCoin(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("3.06"))),
			ContractRewards:  ConvertToProtoDecCoins(sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("5.06")), sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.5")))),
		},
	}

	secondTxTotalRewards := CalculateTotalRewardFromEvents(expectedRewardCalculationEventsFromSecondTx)

	expectedRewardCalculationEvents := append(expectedRewardCalculationEventFromFirstTx, expectedRewardCalculationEventsFromSecondTx...)

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	remainingFeeFirstTx := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(4).QuoInt64(3)))
	remainingFeeSecondTx := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(4)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(4).QuoInt64(3)))

	totalFeeFirstTx := firstTxMaxContractReward.Add(remainingFeeFirstTx...)
	totalFeeSecondTx := secondTxMaxContractReward.Add(remainingFeeSecondTx...)

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[6].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[6].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[4], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 150,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 4, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      20,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstTx[0], &remainingFeeFirstTx[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 6,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      4,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeSecondTx[0], &remainingFeeSecondTx[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsA), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsA); i++ {
		require.Equal(t, expected.rewardsA[i], *leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], *leftOverEntry.ContractRewards[i])
	}

	beginBlockEvents := ctx.EventManager().Events()
	eventType := proto.MessageName(&gstTypes.ContractRewardCalculationEvent{})
	currentIndexToValidate := 0
	for _, event := range beginBlockEvents {
		if event.Type == eventType {
			generatedEvent, err := sdk.TypedEventToEvent(&expectedRewardCalculationEvents[currentIndexToValidate])
			require.NoError(t, err, "should not be an error in event generation")
			currentIndexToValidate += 1

			generatedAttributeValue, err := FindAttributeFromEvent(generatedEvent, "contract_rewards")
			require.NoError(t, err, "should not be an error in finding attribute")

			actualAttributeValue, err := FindAttributeFromEvent(event, "contract_rewards")
			require.NoError(t, err, "should not be an error in finding attribute")

			require.Equal(t, generatedAttributeValue, actualAttributeValue, "generated attribute and actual attribute of event should be equal")
		}
	}
	require.Equal(t, currentIndexToValidate, len(expectedRewardCalculationEvents), "BeginBlock should have generated all events")

	// Now that we have verified that events are as expected, we can check the total reward per tx is less than the fee paid
	rewardDiff, isNegative := totalFeeFirstTx.SafeSub(firstTxTotalRewards)
	require.True(t, rewardDiff.IsAllPositive() && !isNegative, "reward difference must be positive")

	rewardDiff, isNegative = totalFeeSecondTx.SafeSub(secondTxTotalRewards)
	require.True(t, rewardDiff.IsAllPositive() && !isNegative, "reward difference must be positive")
}

func TestContractRewardsWithoutContractPremium(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := disableContractPremium(gstTypes.DefaultParams())
	expected := expect{
		rewardsA: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.20459")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.51071")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.666666666666666666")),
		},
		logs: []*RewardTransferKeeperCallLogs{
			createLogModule(authTypes.FeeCollectorName, gstTypes.ContractRewardCollector, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(3)), sdk.NewCoin("test1", sdk.NewInt(1)))),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[5].String(), sdk.NewCoins()),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[6].String(), sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(2)))),
		},
	}

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200000))
	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	remainingFeeFirstContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))
	remainingFeeSecondContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[6].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[6].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[4], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 150,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 4, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      20,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstContract[0], &remainingFeeFirstContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 6,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      4,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeSecondContract[0], &remainingFeeSecondContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsA), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsA); i++ {
		require.Equal(t, expected.rewardsA[i], *leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], *leftOverEntry.ContractRewards[i])
	}
}

func TestContractRewardsWithDappInflationCapAndZeroPercentage(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := setInflationRewardCap(enableInflationRewardCap(gstTypes.DefaultParams()), 0)
	expected := expect{
		rewardsA: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.200000000000000000")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.55000")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.683333333333333333")),
		},
		logs: []*RewardTransferKeeperCallLogs{
			createLogModule(authTypes.FeeCollectorName, gstTypes.ContractRewardCollector, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(3)), sdk.NewCoin("test1", sdk.NewInt(1)))),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[5].String(), sdk.NewCoins()),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[6].String(), sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(2)))),
		},
	}

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200000))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	remainingFeeFirstContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))
	remainingFeeSecondContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[6].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[6].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[4], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 150,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 4, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      20,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstContract[0], &remainingFeeFirstContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 6,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      4,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeSecondContract[0], &remainingFeeSecondContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsA), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsA); i++ {
		require.Equal(t, expected.rewardsA[i], *leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], *leftOverEntry.ContractRewards[i])
	}
}

func ConvertToProtoDecCoin(decCoin sdk.DecCoin) *sdk.DecCoin {
	return &decCoin
}

func ConvertToProtoDecCoins(coins sdk.DecCoins) []*sdk.DecCoin {
	pt := make([]*sdk.DecCoin, len(coins))
	for i, coin := range coins {
		coinCopy := sdk.NewDecCoinFromDec(coin.Denom, coin.Amount)
		pt[i] = &coinCopy
	}
	return pt
}

func FindAttributeFromEvent(event sdk.Event, attributeName string) ([]byte, error) {
	for _, attr := range event.Attributes {
		if string(attr.Key) == attributeName {
			return attr.Value, nil
		}
	}

	return nil, fmt.Errorf("attribute not found")
}

func CalculateTotalRewardFromEvents(events []gstTypes.ContractRewardCalculationEvent) sdk.DecCoins {
	var totalRewards sdk.DecCoins
	for _, event := range events {
		eventReward := make(sdk.DecCoins, len(event.ContractRewards))
		for i, reward := range event.ContractRewards {
			eventReward[i] = *reward
		}
		totalRewards = totalRewards.Add(eventReward...)
	}
	return totalRewards
}

func TestContractRewardsWithoutDappInflation(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := disableDappInflation(gstTypes.DefaultParams())
	expected := expect{
		rewardsA: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.200000000000000000")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.55000")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.683333333333333333")),
		},
		logs: []*RewardTransferKeeperCallLogs{
			createLogModule(authTypes.FeeCollectorName, gstTypes.ContractRewardCollector, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(3)), sdk.NewCoin("test1", sdk.NewInt(1)))),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[5].String(), sdk.NewCoins()),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[6].String(), sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(2)))),
		},
	}

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200000))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	remainingFeeFirstContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))
	remainingFeeSecondContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[6].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[6].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[4], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 150,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 4, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      20,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstContract[0], &remainingFeeFirstContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 6,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      4,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeSecondContract[0], &remainingFeeSecondContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsA), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsA); i++ {
		require.Equal(t, expected.rewardsA[i], *leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], *leftOverEntry.ContractRewards[i])
	}
}
func TestContractRewardsWithoutGasRebate(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := disableGasRebate(gstTypes.DefaultParams())
	expected := expect{
		rewardsA: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.00459")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.01071")),
		},
		logs: []*RewardTransferKeeperCallLogs{
			createLogModule(authTypes.FeeCollectorName, gstTypes.ContractRewardCollector, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(1)))),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[5].String(), sdk.NewCoins()),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[6].String(), sdk.NewCoins()),
		},
	}

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200000))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	remainingFeeFirstContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))
	remainingFeeSecondContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[6].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[6].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[4], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 150,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 4, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      20,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstContract[0], &remainingFeeFirstContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 6,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      4,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeSecondContract[0], &remainingFeeSecondContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsA), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsA); i++ {
		require.Equal(t, expected.rewardsA[i], *leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], *leftOverEntry.ContractRewards[i])
	}
}

func TestContractRewardWithoutGasRebateAndDappInflation(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := disableDappInflation(disableGasRebate(gstTypes.DefaultParams()))
	expected := expect{
		rewardsA: []sdk.DecCoin{},
		rewardsB: []sdk.DecCoin{},
		logs:     []*RewardTransferKeeperCallLogs{},
	}

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200000))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	remainingFeeFirstContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))
	remainingFeeSecondContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[6].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[6].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[4], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 150,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 4, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      20,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstContract[0], &remainingFeeFirstContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 6,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      4,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeSecondContract[0], &remainingFeeSecondContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	_, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.EqualError(t, err, gstTypes.ErrRewardEntryNotFound.Error(), "We should get left over entry not found")

	_, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.EqualError(t, err, gstTypes.ErrRewardEntryNotFound.Error(), "We should get left over entry not found")
}

func TestContractRewardsWithoutGasTracking(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := disableGasTracking(gstTypes.DefaultParams())
	expected := expect{
		rewardsA: []sdk.DecCoin{},
		rewardsB: []sdk.DecCoin{},
		logs:     []*RewardTransferKeeperCallLogs{},
	}

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200000))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	remainingFeeFirstContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))
	remainingFeeSecondContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[6].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[6].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[4], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 150,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 4, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      20,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstContract[0], &remainingFeeFirstContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 6,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      4,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeSecondContract[0], &remainingFeeSecondContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	_, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.EqualError(t, err, gstTypes.ErrRewardEntryNotFound.Error(), "We should get left over entry not found")

	_, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.EqualError(t, err, gstTypes.ErrRewardEntryNotFound.Error(), "We should get left over entry not found")
}

// "4" Contract's reward is: 1 * (1 / 10) = 0.1test and 0.333333333333 * (1 / 10) = 0.0333333333333test1
func TestContractRewardsWithoutGasRebateToUser(t *testing.T) {
	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	type expect struct {
		rewardsA []sdk.DecCoin
		rewardsB []sdk.DecCoin
		logs     []*RewardTransferKeeperCallLogs
	}
	params := disableGasRebateToUser(gstTypes.DefaultParams())
	expected := expect{
		rewardsA: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.30459")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.1")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.56071")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.683333333333333333")),
		},
		logs: []*RewardTransferKeeperCallLogs{
			createLogModule(authTypes.FeeCollectorName, gstTypes.ContractRewardCollector, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(3)), sdk.NewCoin("test1", sdk.NewInt(1)))),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[5].String(), sdk.NewCoins()),
			createLogAddr(gstTypes.ContractRewardCollector, spareAddress[6].String(), sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(2)))),
		},
	}

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])
	ctx = ctx.WithBlockGasMeter(sdk.NewGasMeter(200000))

	keeper.SetParams(ctx, params)

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	remainingFeeFirstContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))
	remainingFeeSecondContract := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(10)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(26).QuoInt64(3)))

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}

	ctx = ctx.WithBlockHeight(2)

	err := keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[5].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		RewardAddress:    spareAddress[6].String(),
		GasRebateToUser:  false,
		DeveloperAddress: spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[6].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddPendingChangeForContractMetadata(ctx, spareAddress[0], spareAddress[4], gstTypes.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 150,
		DeveloperAddress:         spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	// Commit the pending changes
	numberOfMetadataCommitted, err := keeper.CommitPendingContractMetadata(ctx)
	require.NoError(t, err, "We should be able to commit pending contract metadata")
	require.Equal(t, 4, numberOfMetadataCommitted, "Number of metadata commits should match")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx)

	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      20,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeFirstContract[0], &remainingFeeFirstContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 6,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  2,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      4,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			RemainingFee:       []*sdk.DecCoin{&remainingFeeSecondContract[0], &remainingFeeSecondContract[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 2,
					OriginalVmGas:  2,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// Let's check reward keeper call logs first
	require.Equal(t, len(expected.logs), len(testRewardKeeper.Logs))
	for i := 0; i < len(expected.logs); i++ {
		require.Equal(t, expected.logs[i], testRewardKeeper.Logs[i])
	}

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[5])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsA), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsA); i++ {
		require.Equal(t, expected.rewardsA[i], *leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], *leftOverEntry.ContractRewards[i])
	}
}

// TODO: this is shared test util, that is copied
// from /keeper/keeper_test, refactor
func createTestBaseKeeperAndContext(t *testing.T, contractAdmin sdk.AccAddress) (sdk.Context, *keeper.Keeper) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	appCodec := encodingConfig.Marshaler

	memDB := db.NewMemDB()
	ms := store.NewCommitMultiStore(memDB)

	mkey := sdk.NewKVStoreKey("test")
	tkey := sdk.NewTransientStoreKey("transient_test")
	tstoreKey := sdk.NewTransientStoreKey(gastracker.TStoreKey)

	ms.MountStoreWithDB(mkey, sdk.StoreTypeIAVL, memDB)
	ms.MountStoreWithDB(tkey, sdk.StoreTypeIAVL, memDB)
	ms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, memDB)
	ms.MountStoreWithDB(tstoreKey, sdk.StoreTypeTransient, memDB)

	err := ms.LoadLatestVersion()
	require.NoError(t, err, "Loading latest version should not fail")

	pkeeper := paramskeeper.NewKeeper(appCodec, encodingConfig.Amino, mkey, tkey)
	subspace := pkeeper.Subspace(gstTypes.ModuleName)

	keeper := keeper.NewGasTrackingKeeper(
		storeKey,
		appCodec,
		subspace,
		NewTestContractInfoView(contractAdmin.String()),
		wasmkeeper.NewDefaultWasmGasRegister(),
	)

	ctx := sdk.NewContext(ms, tmproto.Header{
		Time: time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, tmLog.NewTMLogger(os.Stdout))

	params := gstTypes.DefaultParams()
	subspace.SetParamSet(ctx, &params)
	return ctx, keeper
}

func CreateTestKeeperAndContext(t *testing.T, contractAdmin sdk.AccAddress) (sdk.Context, keeper.GasTrackingKeeper) {
	return createTestBaseKeeperAndContext(t, contractAdmin)
}

func CreateTestBlockEntry(ctx sdk.Context, blockTracking gstTypes.BlockGasTracking) {
	kvStore := ctx.KVStore(storeKey)
	bz, err := simapp.MakeTestEncodingConfig().Marshaler.Marshal(&blockTracking)
	if err != nil {
		panic(err)
	}
	kvStore.Set([]byte(gstTypes.CurrentBlockTrackingKey), bz)
}

type TestContractInfoView struct {
	keeper.ContractInfoView
	adminMap     map[string]string
	defaultAdmin string
}

func NewTestContractInfoView(defaultAdmin string) *TestContractInfoView {
	return &TestContractInfoView{
		adminMap:     make(map[string]string),
		defaultAdmin: defaultAdmin,
	}
}

func (t *TestContractInfoView) GetContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo {
	if admin, ok := t.adminMap[contractAddress.String()]; ok {
		return &wasmTypes.ContractInfo{Admin: admin}
	} else {
		return &wasmTypes.ContractInfo{Admin: t.defaultAdmin}
	}
}

func (t *TestContractInfoView) AddContractToAdminMapping(contractAddress string, admin string) {
	t.adminMap[contractAddress] = admin
}

var _ keeper.ContractInfoView = &TestContractInfoView{}

func GenerateRandomAccAddress() sdk.AccAddress {
	var address sdk.AccAddress = make([]byte, 20)
	_, err := rand.Read(address)
	if err != nil {
		panic(err)
	}
	return address
}
