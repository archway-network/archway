package gastracker

import (
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"os"
	"testing"
	"time"

	"github.com/archway-network/archway/x/gastracker/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmLog "github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	db "github.com/tendermint/tm-db"

	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	paramsTypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

type TestContractInfoView struct {
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

var _ ContractInfoView = &TestContractInfoView{}

type subspace struct {
	space map[string]bool
}

func (s *subspace) SetParamSet(ctx sdk.Context, paramset paramsTypes.ParamSet) {
	params, ok := paramset.(*gstTypes.Params)
	if !ok {
		panic("[mock subspace]: invalid params type")
	}
	s.space[string(gstTypes.KeyGasTrackingSwitch)] = params.GasTrackingSwitch
	s.space[string(gstTypes.KeyDappInflationRewards)] = params.GasDappInflationRewardsSwitch
	s.space[string(gstTypes.KeyGasRebateSwitch)] = params.GasRebateSwitch
	s.space[string(gstTypes.KeyGasRebateToUserSwitch)] = params.GasRebateToUserSwitch
	s.space[string(gstTypes.KeyContractPremiumSwitch)] = params.ContractPremiumSwitch

}
func (s *subspace) Get(ctx sdk.Context, key []byte, ptr interface{}) {
	x, ok := ptr.(*bool)
	if !ok {
		panic("[mock subspace]: ptr is invalid type")
	}
	*x = s.space[string(key)]
}

func createTestBaseKeeperAndContext(t *testing.T, contractAdmin sdk.AccAddress) (sdk.Context, *Keeper) {
	memDB := db.NewMemDB()
	ms := store.NewCommitMultiStore(memDB)
	storeKey := sdk.NewKVStoreKey("TestStore")
	ms.MountStoreWithDB(storeKey, sdk.StoreTypeIAVL, memDB)
	err := ms.LoadLatestVersion()
	require.NoError(t, err, "Loading latest version should not fail")
	encodingConfig := simapp.MakeTestEncodingConfig()
	appCodec := encodingConfig.Marshaler

	subspace := subspace{space: make(map[string]bool)}

	keeper := Keeper{
		key:              storeKey,
		appCodec:         appCodec,
		paramSpace:       &subspace,
		contractInfoView: NewTestContractInfoView(contractAdmin.String()),
	}

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, false, tmLog.NewTMLogger(os.Stdout))

	params := gstTypes.DefaultParams()
	subspace.SetParamSet(ctx, &params)
	return ctx, &keeper
}
func CreateTestKeeperAndContext(t *testing.T, contractAdmin sdk.AccAddress) (sdk.Context, GasTrackingKeeper) {
	return createTestBaseKeeperAndContext(t, contractAdmin)
}

// Test various conditions in handling contract metadata
func TestContractMetadataHandling(t *testing.T) {
	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	ctx, keeper := CreateTestKeeperAndContext(t, spareAddress[0])
	// Should return appropriate error when contract metadata is not found
	_, err := keeper.GetContractMetadata(ctx, spareAddress[1])
	require.EqualError(
		t,
		err,
		types.ErrContractInstanceMetadataNotFound.Error(),
		"We should get not found error when try to get non existent contract metadata",
	)

	// No developer and reward address
	incompleteMetadata := types.ContractInstanceMetadata{
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 3,
	}

	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[1], incompleteMetadata)
	require.EqualError(t, err, gstTypes.ErrInvalidSetContractMetadataRequest.Error(), "We should not be able to set metadata")

	// No developer address
	incompleteMetadata = types.ContractInstanceMetadata{
		RewardAddress:            spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 3,
	}

	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[1], incompleteMetadata)
	require.EqualError(t, err, gstTypes.ErrInvalidSetContractMetadataRequest.Error(), "We should not be able to set metadata")

	// No reward address
	incompleteMetadata = types.ContractInstanceMetadata{
		DeveloperAddress:         spareAddress[5].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 3,
	}

	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[1], incompleteMetadata)
	require.EqualError(t, err, gstTypes.ErrInvalidSetContractMetadataRequest.Error(), "We should not be able to set metadata")

	newMetadata := types.ContractInstanceMetadata{
		RewardAddress:            spareAddress[2].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 3,
		DeveloperAddress:         spareAddress[0].String(),
	}

	// Should be successful
	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[1], newMetadata)
	require.NoError(t, err, "We should be able to set metadata")

	// You should be able to omit either developer address or reward address now

	// Test to omit Reward address
	newMetadata = types.ContractInstanceMetadata{
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 3,
		DeveloperAddress:         spareAddress[1].String(),
	}

	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[1], newMetadata)
	require.NoError(t, err, "We should be able to set metadata")

	retrievedMetadata, err := keeper.GetContractMetadata(ctx, spareAddress[1])
	require.NoError(t, err, "We should be able to get metadata")

	require.Equal(t, spareAddress[2].String(), retrievedMetadata.RewardAddress, "The reward address must be the same")
	require.Equal(t, spareAddress[1].String(), retrievedMetadata.DeveloperAddress, "Developer address must be changed")

	// Test to omit Developer address
	newMetadata = types.ContractInstanceMetadata{
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 3,
		RewardAddress:            spareAddress[5].String(),
	}

	err = keeper.SetContractMetadata(ctx, spareAddress[1], spareAddress[1], newMetadata)
	require.NoError(t, err, "We should be able to set metadata")

	retrievedMetadata, err = keeper.GetContractMetadata(ctx, spareAddress[1])
	require.NoError(t, err, "We should be able to get metadata")

	require.Equal(t, spareAddress[5].String(), retrievedMetadata.RewardAddress, "The reward address must be changed")
	require.Equal(t, spareAddress[1].String(), retrievedMetadata.DeveloperAddress, "Developer address must be same")

	// Test to omit both
	newMetadata = types.ContractInstanceMetadata{
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 3,
	}

	err = keeper.SetContractMetadata(ctx, spareAddress[1], spareAddress[1], newMetadata)
	require.NoError(t, err, "We should be able to set metadata")

	retrievedMetadata, err = keeper.GetContractMetadata(ctx, spareAddress[1])
	require.NoError(t, err, "We should be able to get metadata")

	require.Equal(t, spareAddress[5].String(), retrievedMetadata.RewardAddress, "The reward address must be same")
	require.Equal(t, spareAddress[1].String(), retrievedMetadata.DeveloperAddress, "Developer address must be same")

	// Sender validation check

	// Right now default admin is senderAddress[0], passing anything else should not work
	metadata := types.ContractInstanceMetadata{
		DeveloperAddress:         spareAddress[6].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 3,
		RewardAddress:            spareAddress[7].String(),
	}

	err = keeper.SetContractMetadata(ctx, spareAddress[5], spareAddress[2], metadata)
	require.EqualError(t, err, gstTypes.ErrNoPermissionToSetMetadata.Error(), "keeper should not allow metadata change")

	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[2], metadata)
	require.NoError(t, err, "We should be able to set metadata")

	// Now that we already set the metadata and developer address is set to spareAddress[6], we would not be able to change
	// metadata
	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[2], metadata)
	require.EqualError(t, err, gstTypes.ErrNoPermissionToSetMetadata.Error(), "keeper should not allow metadata change")

	metadata = types.ContractInstanceMetadata{
		DeveloperAddress:         spareAddress[8].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 3,
		RewardAddress:            spareAddress[7].String(),
	}

	err = keeper.SetContractMetadata(ctx, spareAddress[6], spareAddress[2], metadata)
	require.NoError(t, err, "We should be able to set metadata")

	metadata = types.ContractInstanceMetadata{
		DeveloperAddress:         spareAddress[9].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 3,
	}

	// Both admin and the previous developer should not be able to set the metadata
	err = keeper.SetContractMetadata(ctx, spareAddress[6], spareAddress[2], metadata)
	require.EqualError(t, err, gstTypes.ErrNoPermissionToSetMetadata.Error(), "keeper should not allow metadata change")

	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[2], metadata)
	require.EqualError(t, err, gstTypes.ErrNoPermissionToSetMetadata.Error(), "keeper should not allow metadata change")

	// Current developer should be able to set the metadata
	err = keeper.SetContractMetadata(ctx, spareAddress[8], spareAddress[2], metadata)
	require.NoError(t, err, "We should be able to set metadata")
}

// Extensive testing of keeper function that merges incoming rewards and stores left over reward
func TestCreateOrMergeLeftOverRewardEntry(t *testing.T) {
	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	ctx, keeper := CreateTestKeeperAndContext(t, spareAddress[0])

	_, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[1])
	require.EqualError(t, err, types.ErrRewardEntryNotFound.Error(), "Getting left over reward entry should fail")

	rewardCoins := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1).QuoInt64(3)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))
	rewardCoins.Sort()

	expectedWholeCoins := sdk.NewCoins()
	wholeCoins, err := keeper.CreateOrMergeLeftOverRewardEntry(ctx, spareAddress[1], rewardCoins, 1)
	require.NoError(t, err, "Creating new reward entry should not fail")
	require.Equal(t, expectedWholeCoins, wholeCoins)

	expectedLeftOverRewards := rewardCoins
	// Left over rewards same as reward coins
	leftOverEntry, err := keeper.GetLeftOverRewardEntry(ctx, spareAddress[1])
	require.NoError(t, err, "Getting left over reward entry should not fail")
	require.Equal(t, len(expectedLeftOverRewards), 2)
	require.Equal(t, len(expectedLeftOverRewards), len(leftOverEntry.ContractRewards))
	require.Equal(t, expectedLeftOverRewards[0], *leftOverEntry.ContractRewards[0])
	require.Equal(t, expectedLeftOverRewards[1], *leftOverEntry.ContractRewards[1])

	// Test1 reward will be 0.5+0.5 = 1 which is greater than or equal to left over threshold
	expectedWholeCoins = sdk.NewCoins(sdk.NewCoin("test1", sdk.NewInt(1)))
	wholeCoins, err = keeper.CreateOrMergeLeftOverRewardEntry(ctx, spareAddress[1], rewardCoins, 1)
	require.NoError(t, err, "Creating new reward entry should not fail")
	require.Equal(t, expectedWholeCoins, wholeCoins)

	// Left over reward will only contain test denomination with value of 0.6666 (0.33333+0.33333)
	expectedLeftOverRewards = sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", rewardCoins[0].Amount.MulInt64(2)))
	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[1])
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
	wholeCoins, err = keeper.CreateOrMergeLeftOverRewardEntry(ctx, spareAddress[1], rewardCoins, 1)
	require.NoError(t, err, "Merging left over reward entry should not result in error")
	require.Equal(t, expectedWholeCoins, wholeCoins, "Wholecoins should be same")

	// Left over rewards are: 0.16666test, 0.5test1, 0.5test2
	expectedLeftOverRewards = []sdk.DecCoin{
		sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.166666666666666666")),
		sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)),
		sdk.NewDecCoinFromDec("test2", sdk.NewDec(1).QuoInt64(2)),
	}
	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[1])
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
	wholeCoins, err = keeper.CreateOrMergeLeftOverRewardEntry(ctx, spareAddress[1], rewardCoins, 2)
	require.NoError(t, err, "We should be able to merge left over reward entry")
	require.Equal(t, wholeCoins, sdk.NewCoins(sdk.NewCoin("test2", sdk.NewInt(3))))

	// Left over coins would be: 1.166666test1 and 1.5test2
	expectedLeftOverRewards = sdk.NewDecCoins(
		sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("1.166666666666666666")),
		sdk.NewDecCoinFromDec("test1", sdk.MustNewDecFromStr("1.5")),
	)
	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[1])
	require.NoError(t, err, "We should be able to get left over reward entry")
	require.Equal(t, len(expectedLeftOverRewards), len(leftOverEntry.ContractRewards))
	require.Equal(t, expectedLeftOverRewards[0], *leftOverEntry.ContractRewards[0])
	require.Equal(t, expectedLeftOverRewards[1], *leftOverEntry.ContractRewards[1])

	// Now, changing back leftOverThreshold to 1 both test and test1 denomination will be released
	expectedWholeCoins = sdk.NewCoins(
		sdk.NewCoin("test", sdk.NewInt(1)),
		sdk.NewCoin("test1", sdk.NewInt(1)),
	)
	wholeCoins, err = keeper.CreateOrMergeLeftOverRewardEntry(ctx, spareAddress[1], sdk.NewDecCoins(), 1)
	require.NoError(t, err, "We should be able to merge empty rewards without an error")
	require.Equal(t, expectedWholeCoins, wholeCoins)

	// Left over entry for test would be (1.166666666666666666 - 1 = 0.166666666666666666test, 1.5 - 1 = 0.5test1)
	expectedLeftOverRewards = sdk.NewDecCoins(
		sdk.NewDecCoinFromDec("test", sdk.MustNewDecFromStr("0.166666666666666666")),
		sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)),
	)
	leftOverEntry, err = keeper.GetLeftOverRewardEntry(ctx, spareAddress[1])
	require.NoError(t, err, "We should be able to get left over entry")
	require.Equal(t, len(expectedLeftOverRewards), len(leftOverEntry.ContractRewards))
	require.Equal(t, expectedLeftOverRewards[0], *leftOverEntry.ContractRewards[0])
	require.Equal(t, expectedLeftOverRewards[1], *leftOverEntry.ContractRewards[1])
}

func TestCalculateUpdatedGas(t *testing.T) {
	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	ctx, keeper := CreateTestKeeperAndContext(t, spareAddress[0])

	// No change in updated gas when contract's metadata does not exists
	gasRecord := wasmTypes.ContractGasRecord{
		OperationId:     wasmTypes.ContractOperationIbcChannelOpen,
		ContractAddress: spareAddress[1].String(),
		GasConsumed:     5,
	}
	updatedGas, err := keeper.CalculateUpdatedGas(ctx, gasRecord)
	require.NoError(t, err, "Calculation of updated gas should be succeed")
	require.Equal(t, gasRecord.GasConsumed, updatedGas)

	// Checking gas rebate calculation
	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		DeveloperAddress:         spareAddress[0].String(),
		RewardAddress:            spareAddress[1].String(),
		GasRebateToUser:          true,
		CollectPremium:           false,
		PremiumPercentageCharged: 0,
	})
	require.NoError(t, err, "SetContractMetadata should be successful")

	gasRecord = wasmTypes.ContractGasRecord{
		OperationId:     wasmTypes.ContractOperationIbcChannelOpen,
		ContractAddress: spareAddress[1].String(),
		GasConsumed:     7,
	}
	updatedGas, err = keeper.CalculateUpdatedGas(ctx, gasRecord)
	require.NoError(t, err, "Calculation of updated gas should be succeed")
	require.Equal(t, gasRecord.GasConsumed/2, updatedGas)

	// Checking premium percentage calculation
	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[1], gstTypes.ContractInstanceMetadata{
		DeveloperAddress:         spareAddress[0].String(),
		RewardAddress:            spareAddress[1].String(),
		GasRebateToUser:          false,
		CollectPremium:           true,
		PremiumPercentageCharged: 50,
	})
	require.NoError(t, err, "SetContractMetadata should be successful")

	gasRecord = wasmTypes.ContractGasRecord{
		OperationId:     wasmTypes.ContractOperationIbcChannelOpen,
		ContractAddress: spareAddress[1].String(),
		GasConsumed:     10,
	}
	updatedGas, err = keeper.CalculateUpdatedGas(ctx, gasRecord)
	require.NoError(t, err, "Calculation of updated gas should be succeed")
	require.Equal(t, gasRecord.GasConsumed+(gasRecord.GasConsumed*50)/100, updatedGas)
}

func TestIngestionOfGasRecords(t *testing.T) {
	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	ctx, keeper := CreateTestKeeperAndContext(t, spareAddress[0])

	err := keeper.TrackNewBlock(ctx)
	require.NoError(t, err, "We should be able to track new block")

	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{}, 5)
	require.NoError(t, err, "We should be able to track new tx")

	// Ingest gas record should be successful, but should skip the entry
	// since there is no contract metadata
	err = keeper.IngestGasRecord(ctx, []wasmTypes.ContractGasRecord{
		{
			OperationId:     wasmTypes.ContractOperationInstantiate,
			ContractAddress: spareAddress[1].String(),
			GasConsumed:     2,
		},
	})
	require.NoError(t, err, "IngestGasRecord should be successful")

	blockTracking, err := keeper.GetCurrentBlockTrackingInfo(ctx)
	require.NoError(t, err, "We should be able to get Current block tracking info")

	require.Equal(t, 1, len(blockTracking.TxTrackingInfos))

	require.Equal(t, 0, len(blockTracking.TxTrackingInfos[0].ContractTrackingInfos))

	// Let's add the metadata and call IngestGasRecord again
	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[2], gstTypes.ContractInstanceMetadata{
		DeveloperAddress: spareAddress[0].String(),
		RewardAddress:    spareAddress[0].String(),
	})
	require.NoError(t, err, "We should be able to set contract metadata")

	err = keeper.SetContractMetadata(ctx, spareAddress[0], spareAddress[3], gstTypes.ContractInstanceMetadata{
		DeveloperAddress: spareAddress[0].String(),
		RewardAddress:    spareAddress[0].String(),
		GasRebateToUser:  true,
	})
	require.NoError(t, err, "We should be able to set contract metadata")

	// First entry is ignored, but since second contract address's metadata
	// exists, contract tracking entry will be added.
	err = keeper.IngestGasRecord(ctx, []wasmTypes.ContractGasRecord{
		{
			OperationId:     wasmTypes.ContractOperationInstantiate,
			ContractAddress: spareAddress[1].String(),
			GasConsumed:     1,
		},
		{
			OperationId:     wasmTypes.ContractOperationIbcPacketReceive,
			ContractAddress: spareAddress[2].String(),
			GasConsumed:     2,
		},
		{
			OperationId:     wasmTypes.ContractOperationMigrate,
			ContractAddress: spareAddress[3].String(),
			GasConsumed:     3,
		},
	})
	require.NoError(t, err, "IngestGasRecord should be successful")

	blockTracking, err = keeper.GetCurrentBlockTrackingInfo(ctx)
	require.NoError(t, err, "We should be able to get Current block tracking info")

	require.Equal(t, 1, len(blockTracking.TxTrackingInfos))

	require.Equal(t, 2, len(blockTracking.TxTrackingInfos[0].ContractTrackingInfos))

	require.Equal(t, &gstTypes.ContractGasTracking{
		Address:             spareAddress[3].String(),
		GasConsumed:         3,
		IsEligibleForReward: false,
		Operation:           gstTypes.ContractOperation_CONTRACT_OPERATION_MIGRATE,
	}, blockTracking.TxTrackingInfos[0].ContractTrackingInfos[1])

	require.Equal(t, &gstTypes.ContractGasTracking{
		Address:             spareAddress[2].String(),
		GasConsumed:         2,
		IsEligibleForReward: true,
		Operation:           gstTypes.ContractOperation_CONTRACT_OPERATION_IBC,
	}, blockTracking.TxTrackingInfos[0].ContractTrackingInfos[0])

}

// Test storing and retrieving contract gas usage
func TestAddContractGasUsage(t *testing.T) {
	var spareAddress = make([]sdk.AccAddress, 10)
	for i := 0; i < 10; i++ {
		spareAddress[i] = GenerateRandomAccAddress()
	}

	ctx, keeper := createTestBaseKeeperAndContext(t, spareAddress[0])

	err := keeper.TrackContractGasUsage(ctx, spareAddress[1], 1, types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION, false)
	require.EqualError(t, err, types.ErrBlockTrackingDataNotFound.Error(), "We cannot track contract gas since block tracking does not exists")

	err = keeper.TrackNewBlock(ctx)
	require.NoError(t, err, "We should be able to track new block")

	err = keeper.TrackContractGasUsage(ctx, spareAddress[1], 1, types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION, false)
	require.EqualError(t, err, types.ErrTxTrackingDataNotFound.Error(), "We cannot track contract gas since tx tracking does not exists")

	// Let's track one tx with one contract gas usage
	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{}, 5)
	require.NoError(t, err, "We should be able to track new transaction")
	err = keeper.TrackContractGasUsage(ctx, spareAddress[1], 1, types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION, false)
	require.NoError(t, err, "We should be able to track contract gas since block tracking obj and tx tracking obj exists")

	err = keeper.TrackNewTx(ctx, []*sdk.DecCoin{}, 6)
	require.NoError(t, err, "We should be able to track new transaction")
	err = keeper.TrackContractGasUsage(ctx, spareAddress[2], 2, types.ContractOperation_CONTRACT_OPERATION_REPLY, true)
	require.NoError(t, err, "We should be able to track contract gas since block tracking obj and tx tracking obj exists")
	err = keeper.TrackContractGasUsage(ctx, spareAddress[3], 3, types.ContractOperation_CONTRACT_OPERATION_SUDO, true)
	require.NoError(t, err, "We should be able to track contract gas since block tracking obj and tx tracking obj exists")

	blockTrackingObj, err := keeper.GetCurrentBlockTrackingInfo(ctx)
	require.NoError(t, err, "We should be able to get block tracking object")
	require.Equal(t, 2, len(blockTrackingObj.TxTrackingInfos))
	require.Equal(t, types.TransactionTracking{
		MaxGasAllowed:      5,
		MaxContractRewards: nil,
		ContractTrackingInfos: []*types.ContractGasTracking{
			{
				Address:             spareAddress[1].String(),
				GasConsumed:         1,
				IsEligibleForReward: false,
				Operation:           types.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
			},
		},
	}, *blockTrackingObj.TxTrackingInfos[0])
	require.Equal(t, types.TransactionTracking{
		MaxGasAllowed:      6,
		MaxContractRewards: nil,
		ContractTrackingInfos: []*types.ContractGasTracking{
			{
				Address:             spareAddress[2].String(),
				GasConsumed:         2,
				IsEligibleForReward: true,
				Operation:           types.ContractOperation_CONTRACT_OPERATION_REPLY,
			},
			{
				Address:             spareAddress[3].String(),
				GasConsumed:         3,
				IsEligibleForReward: true,
				Operation:           types.ContractOperation_CONTRACT_OPERATION_SUDO,
			},
		},
	}, *blockTrackingObj.TxTrackingInfos[1])

	err = keeper.TrackNewBlock(ctx)
	require.NoError(t, err, "We should be able to track new block")

	blockTrackingObj, err = keeper.GetCurrentBlockTrackingInfo(ctx)
	require.NoError(t, err, "We should be able to get the block tracking obj")
	// It should be empty
	require.Equal(t, types.BlockGasTracking{}, blockTrackingObj)
}

// Test initialization of block tracking data for new block as well as marking end of the block for current block tracking
// data
func TestBlockTrackingReadWrite(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t, sdk.AccAddress{})

	dummyTxTracking1 := types.TransactionTracking{
		MaxGasAllowed: 500,
	}

	dummyTxTracking2 := types.TransactionTracking{
		MaxGasAllowed: 1000,
	}

	err := keeper.TrackNewBlock(ctx)
	require.NoError(t, err, "We should be able to track new block")

	CreateTestBlockEntry(ctx, keeper.key, keeper.appCodec, types.BlockGasTracking{TxTrackingInfos: []*types.TransactionTracking{&dummyTxTracking1}})

	// We should be able to retrieve the block tracking info
	currentBlockTrackingInfo, err := keeper.GetCurrentBlockTrackingInfo(ctx)
	require.NoError(t, err, "We should be able to get current block tracking")
	require.Equal(t, len(currentBlockTrackingInfo.TxTrackingInfos), 1)
	require.Equal(t, dummyTxTracking1, *currentBlockTrackingInfo.TxTrackingInfos[0])

	err = keeper.TrackNewBlock(ctx)
	require.NoError(t, err, "We should be able to track new block in any case")

	CreateTestBlockEntry(ctx, keeper.key, keeper.appCodec, types.BlockGasTracking{TxTrackingInfos: []*types.TransactionTracking{&dummyTxTracking2}})

	currentBlockTrackingInfo, err = keeper.GetCurrentBlockTrackingInfo(ctx)
	require.NoError(t, err, "We should be able to get current block")
	require.Equal(t, len(currentBlockTrackingInfo.TxTrackingInfos), 1)
	require.Equal(t, dummyTxTracking2, *currentBlockTrackingInfo.TxTrackingInfos[0])
}
