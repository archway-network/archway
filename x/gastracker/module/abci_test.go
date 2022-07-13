package module

import (
	"fmt"
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

func (t *TestRewardTransferKeeper) SendCoinsFromModuleToAccount(_ sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
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

func (t *TestRewardTransferKeeper) SendCoinsFromModuleToModule(_ sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
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
	B Behaviour
}

func (t *TestMintParamsKeeper) GetParams(_ sdk.Context) (params mintTypes.Params) {
	if t.B == Panic {
		panic("TestPanic")
	}
	return mintTypes.Params{
		MintDenom:     "test",
		BlocksPerYear: 100,
	}
}

func (t *TestMintParamsKeeper) GetMinter(_ sdk.Context) (minter mintTypes.Minter) {
	if t.B == Panic {
		panic("TestPanic")
	}
	return mintTypes.Minter{
		AnnualProvisions: sdk.NewDec(76500),
	}
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

func disableDappInflation(params gstTypes.Params) gstTypes.Params {
	params.DappInflationRewardsRatio = sdk.ZeroDec()
	return params
}
func disableGasRebate(params gstTypes.Params) gstTypes.Params {
	params.DappTxFeeRebateRatio = sdk.ZeroDec()
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

	ctx = ctx.WithBlockHeight(2)
	require.PanicsWithError(t, gstTypes.ErrBlockTrackingDataNotFound.Error(), func() {
		BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper)
	}, "BeginBlock should panic")

	ctx = ctx.WithBlockHeight(1)

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper)
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

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper)

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
// Total gas across block is 10 (2+4+1+2+1)
// So, inflation reward for
// Contract "1" is: 0.00153 (153 * 2 / 200000)
// Contract "2" is: 0.00306 (153 * 4 / 200000)
// Contract "3" is: 0.000765 (153 * 1 / 200000)
// Contract "2" is: 0.00153 (153 * 2 / 200000)
// Contract "4" is: 0.000765 (153 * 1 / 200000)
// All above is in "test" denomination, since that is the denomination minter is minting
// Now, coming to gas reward calculations:
// For First tx entry:
//    Gas Used = 10
//    "1" Contract's reward is: 1 * (2 / 10) = 0.2test and 0.33333 * (2 / 10) = 0.0666666test1
//    "2" Contract's reward is: 1 * (4 / 10) = 0.4test and 0.33333 * (4 / 10) = 0.1333333test1
//    "3" Contract's reward is: 1 * (1 / 10) = 0.15test (0.1test + 0.05test (Premium)) and 0.33333 * (1 / 10) = 0.04999995test1 (0.0333333test1 + 0.01666665test1 (premium))
// For Second tx entry:
//   Gas Used = 2
//   "2" Contract's reward is: 2 * (2 / 2) = 2test and 0.5 * (2 / 2) = 0.5test1
// Total rewards:
// For Contract "1": 0.20153test (0.00153 + 0.2) and 0.0666666test1
// For Contract "2": 2.40459test (0.00306 + 0.00153 + 0.4 + 2) and 0.6333333test1 (0.1333333 + 0.5)
// For Contract "3": 0.150765test (0.000765 + 0.15) and 0.04999995test1
// For Contract "4": 0.000765test (0.000706)
// Reward distribution per address:
// (for contract "1" and "4") "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt": 0.20153test + 0.000765test (0.202295test) and 0.0666666test1
// (for contract "2" and "3") "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk": 2.555355test  and 0.6666666test1
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
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.202295")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.555355")),
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

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}

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

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      10,
			MaxContractRewards: []sdk.DecCoin{firstTxMaxContractReward[0], firstTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 3,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      2,
			MaxContractRewards: []sdk.DecCoin{secondTxMaxContractReward[0], secondTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper)
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
		require.Equal(t, expected.rewardsA[i], leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], leftOverEntry.ContractRewards[i])
	}
}

//func TestRewardCalculation(t *testing.T) {
//}

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
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.202295")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.505355")),
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

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}

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

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      10,
			MaxContractRewards: []sdk.DecCoin{firstTxMaxContractReward[0], firstTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 3,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      2,
			MaxContractRewards: []sdk.DecCoin{secondTxMaxContractReward[0], secondTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper)
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
		require.Equal(t, expected.rewardsA[i], leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], leftOverEntry.ContractRewards[i])
	}
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
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.2")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.55")),
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

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}

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

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      10,
			MaxContractRewards: []sdk.DecCoin{firstTxMaxContractReward[0], firstTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 3,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      2,
			MaxContractRewards: []sdk.DecCoin{secondTxMaxContractReward[0], secondTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper)
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
		require.Equal(t, expected.rewardsA[i], leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], leftOverEntry.ContractRewards[i])
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
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.002295")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.005355")),
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

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}

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

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      10,
			MaxContractRewards: []sdk.DecCoin{firstTxMaxContractReward[0], firstTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 3,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      2,
			MaxContractRewards: []sdk.DecCoin{secondTxMaxContractReward[0], secondTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper)
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
		require.Equal(t, expected.rewardsA[i], leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], leftOverEntry.ContractRewards[i])
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

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}

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

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      10,
			MaxContractRewards: []sdk.DecCoin{firstTxMaxContractReward[0], firstTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 3,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      2,
			MaxContractRewards: []sdk.DecCoin{secondTxMaxContractReward[0], secondTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper)
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

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}

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

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      10,
			MaxContractRewards: []sdk.DecCoin{firstTxMaxContractReward[0], firstTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 3,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      2,
			MaxContractRewards: []sdk.DecCoin{secondTxMaxContractReward[0], secondTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper)
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
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.302295")),
			sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.100000000000000000")),
		},
		rewardsB: []sdk.DecCoin{
			sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.555355")),
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

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}

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

	CreateTestBlockEntry(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []gstTypes.TransactionTracking{
		{
			MaxGasAllowed:      10,
			MaxContractRewards: []sdk.DecCoin{firstTxMaxContractReward[0], firstTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[1].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[2].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 3,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[3].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address:        spareAddress[4].String(),
					OriginalVmGas:  1,
					OriginalSdkGas: 0,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed:      2,
			MaxContractRewards: []sdk.DecCoin{secondTxMaxContractReward[0], secondTxMaxContractReward[1]},
			ContractTrackingInfos: []gstTypes.ContractGasTracking{
				{
					Address:        spareAddress[2].String(),
					OriginalSdkGas: 1,
					OriginalVmGas:  1,
					Operation:      gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper)
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
		require.Equal(t, expected.rewardsA[i], leftOverEntry.ContractRewards[i])
	}

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[6])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expected.rewardsB), len(leftOverEntry.ContractRewards))
	for i := 0; i < len(expected.rewardsB); i++ {
		require.Equal(t, expected.rewardsB[i], leftOverEntry.ContractRewards[i])
	}
}

// TODO: this is shared test util, that is copied
// from /keeper/keeper_test, refactor
func createTestBaseKeeperAndContext(t *testing.T, contractAdmin sdk.AccAddress) (sdk.Context, keeper.Keeper) {
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
		&TestMintParamsKeeper{B: Log},
	)

	ctx := sdk.NewContext(ms, tmproto.Header{
		Time: time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, tmLog.NewTMLogger(os.Stdout))

	params := gstTypes.DefaultParams()
	subspace.SetParamSet(ctx, &params)
	return ctx, keeper
}

func CreateTestKeeperAndContext(t *testing.T, contractAdmin sdk.AccAddress) (sdk.Context, keeper.Keeper) {
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

func (t *TestContractInfoView) GetContractInfo(_ sdk.Context, contractAddress sdk.AccAddress) *wasmTypes.ContractInfo {
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
