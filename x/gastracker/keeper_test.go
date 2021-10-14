package gastracker

import (
	"github.com/archway-network/archway/x/gastracker/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"
	"testing"
	"time"
)

func CreateTestKeeperAndContext(t *testing.T) (sdk.Context, GasTrackingKeeper) {
	memDB := db.NewMemDB()
	ms := store.NewCommitMultiStore(memDB)
	storeKey := sdk.NewKVStoreKey("TestStore")
	ms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, memDB)
	err := ms.LoadLatestVersion()
	require.NoError(t, err, "Loading latest version should not fail")
	encodingConfig := simapp.MakeTestEncodingConfig()

	keeper := Keeper{
		key:      storeKey,
		appCodec: encodingConfig.Marshaler,
	}

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, log.NewNopLogger())

	return ctx, &keeper
}

func TestContractMetadataHandling(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)
	// Should return appropriate error
	_, err := keeper.GetNewContractMetadata(ctx, "1")
	require.EqualError(
		t,
		err,
		types.ErrContractInstanceMetadataNotFound.Error(),
		"We should get not found error when try to get non existent contract metadata",
	)

	newMetadata := types.ContractInstanceMetadata{
		RewardAddress:   "2",
		GasRebateToUser: false,
	}

	// Should be successful
	err = keeper.AddNewContractMetadata(ctx, "1", newMetadata)

	// Should be able to get the new stored contract metadata
	metadata, err := keeper.GetNewContractMetadata(ctx, "1")
	// Error must be nil
	require.NoError(t, err, "We should be able to get already existing metadata")
	// Metadata must match the one we stored
	require.Equal(t, newMetadata, metadata)

	// Should be successful (we should be able to overwrite the existing metadata
	updatedMetadata := types.ContractInstanceMetadata{
		RewardAddress:   "3",
		GasRebateToUser: true,
	}
	err = keeper.AddNewContractMetadata(ctx, "1", updatedMetadata)
	require.NoError(t, err, "We should be able to overwrite existing metadata")

	// Should be able to get the new stored contract metadata
	metadata, err = keeper.GetNewContractMetadata(ctx, "1")
	// Error must be nil
	require.NoError(t, err, "We should be able to get already existing metadata")
	// Metadata must match the one we stored last
	require.Equal(t, updatedMetadata, metadata)
}

func TestCreateOrMergeLeftOverRewardEntry(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)

	rewardCoins := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1).QuoInt64(3)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))
	rewardCoins.Sort()

	expectedWholeCoins := sdk.NewCoins()
	wholeCoins, err := keeper.CreateOrMergeLeftOverRewardEntry(ctx, "1", rewardCoins, 1)
	require.NoError(t, err, "Creating new reward entry should not fail")
	require.Equal(t, expectedWholeCoins, wholeCoins)

	expectedLeftOverRewards := rewardCoins
	// Left over rewards same as reward coins
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, "1")
	require.NoError(t, err, "Getting left over reward entry should not fail")
	require.Equal(t, len(expectedLeftOverRewards), 2)
	require.Equal(t, len(expectedLeftOverRewards), len(leftOverEntry.ContractRewards))
	require.Equal(t, expectedLeftOverRewards[0], *leftOverEntry.ContractRewards[0])
	require.Equal(t, expectedLeftOverRewards[1], *leftOverEntry.ContractRewards[1])

	// Test1 reward will be 0.5+0.5 = 1 which is greater than or equal to left over threshold
	expectedWholeCoins = sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(1)))
	wholeCoins, err = keeper.CreateOrMergeLeftOverRewardEntry(ctx, "1", rewardCoins, 1)
	require.NoError(t, err, "Creating new reward entry should not fail")
	require.Equal(t, expectedWholeCoins, wholeCoins)

	// Left over reward will only contain test denomination with value of 0.6666 (0.33333+0.33333)
	expectedLeftOverRewards = sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", rewardCoins[0].Amount.MulInt64(2)))
	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, "1")
	require.NoError(t, err, "Getting left over reward entry should not fail")
	require.Equal(t, len(expectedLeftOverRewards), 1)
	require.Equal(t, len(expectedLeftOverRewards), len(leftOverEntry.ContractRewards))
	require.Equal(t, expectedLeftOverRewards[0], *leftOverEntry.ContractRewards[0])

	rewardCoins = sdk.NewDecCoins(
		sdk.NewDecCoinFromDec("test", sdk.NewDec(11).QuoInt64(2)),
		sdk.NewDecCoinFromDec("test1", sdk.NewDec(7).QuoInt64(2)),
		sdk.NewDecCoinFromDec("test2", sdk.NewDec(11).QuoInt64(2)),
	)
	// Whole coins would be 6test (5.5 + 0.666 = 6.16666), 3test1 (3.5), 5test2 (11/2)
	expectedWholeCoins = sdk.NewCoins(sdk.NewInt64Coin("test", 6), sdk.NewInt64Coin("test1", 3), sdk.NewInt64Coin("test2", 5))
	wholeCoins, err = keeper.CreateOrMergeLeftOverRewardEntry(ctx, "1", rewardCoins, 1)
	require.NoError(t, err, "Merging left over reward entry should not result in error")
	require.Equal(t, expectedWholeCoins, wholeCoins, "Wholecoins should be same")


	// Left over rewards are: 0.16666test, 0.5test1, 0.5test2
	expectedLeftOverRewards = []sdk.DecCoin{
		sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.166666666666666666")),
		sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)),
		sdk.NewDecCoinFromDec("test2", sdk.NewDec(1).QuoInt64(2)),
	}
	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, "1")
	require.NoError(t, err, "We should be able to get left over entry without an error")
	require.Equal(t, len(expectedLeftOverRewards), len(leftOverEntry.ContractRewards))
	require.Equal(t, expectedLeftOverRewards[0], *leftOverEntry.ContractRewards[0])
	require.Equal(t, expectedLeftOverRewards[1], *leftOverEntry.ContractRewards[1])
	require.Equal(t, expectedLeftOverRewards[2], *leftOverEntry.ContractRewards[2])

	// Now, let's change the leftOverThreshold to 2
	// The wholecoin we would get is 3test2 (2.5 + 0.5 = 3 > 2)
	// test1 and test2 denomination won't be part of wholecoins (1 + 0.1666 = 1.16666test1 and 1 + 0.5 = 1.5test2)
	rewardCoins = sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test2", sdk.NewDec(5).QuoInt64(2)))
	rewardCoins.Sort()
	wholeCoins, err = keeper.CreateOrMergeLeftOverRewardEntry(ctx, "1", rewardCoins, 2)
	require.NoError(t, err, "We should be able to merge left over reward entry")
	require.Equal(t, wholeCoins, sdk.NewCoins(sdk.NewCoin("test2", sdk.NewInt(3))))

	// Left over coins would be: 1.166666test1 and 1.5test2
	expectedLeftOverRewards = sdk.NewDecCoins(
		sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("1.166666666666666666")),
		sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("1.5")),
	)
	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, "1")
	require.NoError(t, err, "We should be able to get left over reward entry")
	require.Equal(t, len(expectedLeftOverRewards), len(leftOverEntry.ContractRewards))
	require.Equal(t, expectedLeftOverRewards[0], *leftOverEntry.ContractRewards[0])
	require.Equal(t, expectedLeftOverRewards[1], *leftOverEntry.ContractRewards[1])

	// Now, changing back leftOverThreshold to 1 both test and test1 denomination will be released
	expectedWholeCoins = sdk.NewCoins(
		sdk.NewCoin("test", sdk.NewInt(1)),
		sdk.NewCoin("test1", sdk.NewInt(1)),
	)
	wholeCoins, err = keeper.CreateOrMergeLeftOverRewardEntry(ctx, "1", sdk.NewDecCoins(), 1)
	require.NoError(t, err, "We should be able to merge empty rewards without an error")
	require.Equal(t, expectedWholeCoins, wholeCoins)

	// Left over entry for test would be (1.166666666666666666 - 1 = 0.166666666666666666test, 1.5 - 1 = 0.5test1)
	expectedLeftOverRewards = sdk.NewDecCoins(
		sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.166666666666666666")),
		sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)),
	)
	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, "1")
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expectedLeftOverRewards), len(leftOverEntry.ContractRewards))
	require.Equal(t, expectedLeftOverRewards[0], *leftOverEntry.ContractRewards[0])
	require.Equal(t, expectedLeftOverRewards[1], *leftOverEntry.ContractRewards[1])
}


func TestAddContractGasUsage(t *testing.T) {
	ctx, keeper := CreateTestKeeperAndContext(t)

	err := keeper.TrackContractGasUsage(ctx, "1", 1, types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION, false)
	require.EqualError(t, err, types.ErrBlockTrackingDataNotFound.Error(),"We cannot track contract gas since block tracking does not exists")

	err = keeper.TrackNewBlock(ctx, types.BlockGasTracking{})
	require.NoError(t, err, "We should be able to track new block")

	err = keeper.TrackContractGasUsage(ctx, "1", 1, types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION, false)
	require.EqualError(t, err, types.ErrTxTrackingDataNotFound.Error(),"We cannot track contract gas since tx tracking does not exists")

	// Let's track one tx with one contract gas usage
	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{}, 5)
	require.NoError(t, err, "We should be able to track new transaction")
	err = keeper.TrackContractGasUsage(ctx, "1", 1, types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION, false)
	require.NoError(t, err, "We should be able to track contract gas since block tracking obj and tx tracking obj exists")

	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{}, 6)
	require.NoError(t, err, "We should be able to track new transaction")
	err = keeper.TrackContractGasUsage(ctx, "2", 2, types.ContractOperation_CONTRACT_OPERATION_REPLY, true)
	require.NoError(t, err, "We should be able to track contract gas since block tracking obj and tx tracking obj exists")
	err = keeper.TrackContractGasUsage(ctx, "3", 3, types.ContractOperation_CONTRACT_OPERATION_SUDO, true)
	require.NoError(t, err, "We should be able to track contract gas since block tracking obj and tx tracking obj exists")

	blockTrackingObj, err := keeper.GetCurrentBlockTrackingInfo(ctx)
	require.NoError(t, err, "We should be able to get block tracking object")
	require.Equal(t, 2, len(blockTrackingObj.TxTrackingInfos))
	require.Equal(t, types.TransactionTracking{
		MaxGasAllowed:         5,
		MaxContractRewards:    nil,
		ContractTrackingInfos: []*types.ContractGasTracking{
			{
				Address:             "1",
				GasConsumed:         1,
				IsEligibleForReward: false,
				Operation:           types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
			},
		},
	}, *blockTrackingObj.TxTrackingInfos[0])
	require.Equal(t, types.TransactionTracking{
		MaxGasAllowed:         6,
		MaxContractRewards:    nil,
		ContractTrackingInfos: []*types.ContractGasTracking{
			{
				Address:             "2",
				GasConsumed:         2,
				IsEligibleForReward: true,
				Operation:           types.ContractOperation_CONTRACT_OPERATION_REPLY,
			},
			{
				Address:             "3",
				GasConsumed:         3,
				IsEligibleForReward: true,
				Operation:           types.ContractOperation_CONTRACT_OPERATION_SUDO,
			},
		},
	}, *blockTrackingObj.TxTrackingInfos[1])

	err = keeper.TrackNewBlock(ctx, types.BlockGasTracking{})
	require.NoError(t, err, "We should be able to track new block")

	blockTrackingObj, err = keeper.GetCurrentBlockTrackingInfo(ctx)
	require.NoError(t,err, "We should be able to get the block tracking obj")
	// It should be empty
	require.Equal(t, types.BlockGasTracking{}, blockTrackingObj)
}
