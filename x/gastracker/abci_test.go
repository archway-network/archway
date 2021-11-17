package gastracker

import (
	"fmt"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/abci/types"
	"testing"
)

type Behaviour int

const (
	Log Behaviour = iota
	Error
	Panic
)

type RewardTransferKeeperCallLogs struct {
	Method string
	senderModule string
	recipientModule string
	recipientAddr string
	amt sdk.Coins
}

type TestRewardTransferKeeper struct {
	Logs []*RewardTransferKeeperCallLogs
	B Behaviour
}

func (t *TestRewardTransferKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	switch t.B {
	case Log:
		t.Logs = append(t.Logs, &RewardTransferKeeperCallLogs{
			Method: "SendCoinsFromModuleToAccount",
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


func (t * TestRewardTransferKeeper) SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
	switch t.B {
	case Log:
		t.Logs = append(t.Logs, &RewardTransferKeeperCallLogs{
			Method: "SendCoinsFromModuleToModule",
			senderModule:  senderModule,
			recipientModule: recipientModule,
			amt:           amt,
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
		MintDenom: "test",
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

func TestBlockTracking(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)

	config := sdk.GetConfig()
	config.SetBech32PrefixForAccount("archway", "archway")

	zeroDecCoin := sdk.NewDecCoinFromDec("test", sdk.NewDec(0))

	// Empty new block tracking object
	err := keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed: 1,
			MaxContractRewards: []*sdk.DecCoin{&zeroDecCoin},
			ContractTrackingInfos: nil,
		},
	}})
	require.NoError(t, err, "We should be able to track new block")

	err = keeper.MarkEndOfTheBlock(ctx)
	require.NoError(t, err, "We should be able to end the block")

	testRewardKeeper := &TestRewardTransferKeeper{B: Log}
	testMintParamsKeeper := &TestMintParamsKeeper{B: Log}
	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// We should not have made any call to reward keeper
	require.Zero(t, testRewardKeeper.Logs, "No logs should be there as no need to make new calls")
	// We would have overwritten the TrackNewBlock obj
	blockGasTracking, err := keeper.GetCurrentBlockTrackingInfo(ctx)
	require.NoError(t, err, "We should be able to get new block gas tracking")
	require.Equal(t, gstTypes.BlockGasTracking{}, blockGasTracking, "We should have overwritten block gas tracking obj")

	err = keeper.AddNewContractMetadata(ctx, "1", gstTypes.ContractInstanceMetadata{
		RewardAddress:   "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		GasRebateToUser: false,
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddNewContractMetadata(ctx, "2", gstTypes.ContractInstanceMetadata{
		RewardAddress:   "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk",
		GasRebateToUser: false,
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddNewContractMetadata(ctx, "3", gstTypes.ContractInstanceMetadata{
		RewardAddress:   "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk",
		GasRebateToUser: false,
		CollectPremium: true,
		PremiumPercentageCharged: 50,
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	err = keeper.AddNewContractMetadata(ctx, "4", gstTypes.ContractInstanceMetadata{
		RewardAddress:   "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		GasRebateToUser: false,
		CollectPremium: true,
		PremiumPercentageCharged: 150,
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))

	err = keeper.MarkEndOfTheBlock(ctx)
	require.NoError(t, err, "We should be able to end the block")

	// Tracking new block with multiple tx tracking obj
	err = keeper.TrackNewBlock(ctx, gstTypes.BlockGasTracking{TxTrackingInfos: []*gstTypes.TransactionTracking{
		{
			MaxGasAllowed: 10,
			MaxContractRewards: []*sdk.DecCoin{&firstTxMaxContractReward[0], &firstTxMaxContractReward[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address: "1",
					GasConsumed: 2,
					IsEligibleForReward: true,
					Operation: gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address: "2",
					GasConsumed: 4,
					IsEligibleForReward: true,
					Operation: gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address: "3",
					GasConsumed: 1,
					IsEligibleForReward: true,
					Operation: gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address: "4",
					GasConsumed: 1,
					IsEligibleForReward: false,
					Operation: gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
			},
		},
		{
			MaxGasAllowed: 2,
			MaxContractRewards: []*sdk.DecCoin{&secondTxMaxContractReward[0], &secondTxMaxContractReward[1]},
			ContractTrackingInfos: []*gstTypes.ContractGasTracking{
				{
					Address: "2",
					GasConsumed: 2,
					IsEligibleForReward: true,
					Operation: gstTypes.ContractOperation_CONTRACT_OPERATION_SUDO,
				},
			},
		},
	}})
	require.NoError(t, err, "We should be able to track new block")

	err = keeper.MarkEndOfTheBlock(ctx)
	require.NoError(t, err, "We should be able to end the block")

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)

	// Inflation reward cap for a block is hardcoded at 20% of total inflation reward
	// So, for per block inflation of 765 (76500/100), we need to take 20% of it which is
	// 153.
	// Total gas across block is 10 (2+4+1+2+1)
	// So, inflation reward for
	// Contract "1" is: 30.6 (153 * 2 / 10)
	// Contract "2" is: 61.2 (153 * 4 / 10)
	// Contract "3" is: 15.3 (153 * 1 / 10)
	// Contract "2" is: 30.6 (153 * 2 / 10)
	// Contract "4" does not get any reward
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
	// For Contract "1": 30.8test (30.6 + 0.2) and 0.0666666test1
	// For Contract "2": 94.2test (61.2 + 30.6 + 0.4 + 2) and 0.6333333test1 (0.1333333 + 0.5)
	// For Contract "3": 15.45test (15.3 + 0.15) and 0.04999995test1
	// Reward distribution per address:
	// (for contract "1") "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt": 30.8test and 0.0666666test1
	// (for contract "2" and "3") "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk": 109.65test and 0.6666666test1
	// So, we should be fetching 141test (30.8 + 109.65 = 140.45 rounded to 141) and 1test1 (0.6333333 + 0.04999995 =  0.68333325 rounded to 1)
	// from the fee collector
	// Since, left over threshold is hard coded to 1, we should be transferring 30test to "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt" and
	// 109test to "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk" and left over rewards should be 0.8test,0.0666666test1 and 0.65test and 0.68333325test1
	// respectively.

	// Let's check reward keeper call logs first
	require.Equal(t, 3, len(testRewardKeeper.Logs))
	require.Equal(t, &RewardTransferKeeperCallLogs{
		Method:       "SendCoinsFromModuleToModule",
		senderModule: authTypes.FeeCollectorName,
		recipientModule: gstTypes.ContractRewardCollector,
		amt: sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(141)), sdk.NewCoin("test1", sdk.NewInt(1))),
	}, testRewardKeeper.Logs[0])
	require.Equal(t, &RewardTransferKeeperCallLogs{
		Method: "SendCoinsFromModuleToAccount",
		senderModule: gstTypes.ContractRewardCollector,
		recipientAddr: "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt",
		amt: sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(30))),
	}, testRewardKeeper.Logs[1])
	require.Equal(t, &RewardTransferKeeperCallLogs{
		Method: "SendCoinsFromModuleToAccount",
		senderModule: gstTypes.ContractRewardCollector,
		recipientAddr: "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk",
		amt: sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(109))),
	}, testRewardKeeper.Logs[2])

	// Let's check left-over balances
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, "archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt")
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, 2, len(leftOverEntry.ContractRewards))
	require.Equal(t, sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.8")), *leftOverEntry.ContractRewards[0])
	require.Equal(t, sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.066666666666666667")), *leftOverEntry.ContractRewards[1])

	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, "archway1j08452mqwadp8xu25kn9rleyl2gufgfjls8ekk")
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, 2, len(leftOverEntry.ContractRewards))
	require.Equal(t, sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.65")), *leftOverEntry.ContractRewards[0])
	require.Equal(t, sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("0.683333333333333333")), *leftOverEntry.ContractRewards[1])
}
