package gastracker

import (
	"fmt"
	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	recipientAddr sdk.AccAddress
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
			recipientAddr: recipientAddr,
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
	})
	require.NoError(t, err, "We should be able to add new contract metadata")

	firstTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(1)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(3)))
	secondTxMaxContractReward := sdk.NewDecCoins(sdk.NewDecCoinFromDec("test", sdk.NewDec(2)), sdk.NewDecCoinFromDec("test1", sdk.NewDec(1).QuoInt64(2)))
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
					GasConsumed: 2,
					IsEligibleForReward: true,
					Operation: gstTypes.ContractOperation_CONTRACT_OPERATION_INSTANTIATION,
				},
				{
					Address: "3",
					GasConsumed: 2,
					IsEligibleForReward: true,
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

	BeginBlock(ctx, types.RequestBeginBlock{}, keeper, testRewardKeeper, testMintParamsKeeper)
	// TODO: Verify the results
}
