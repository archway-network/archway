package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// TestWithdrawRewardsByLimit tests the withdraw operation using record limit mode.
func TestWithdrawRewardsByLimit(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	wk := testutils.NewMockContractViewer()
	k.SetContractInfoViewer(wk)
	accAddr := testutils.AccAddress()

	testData := []withdrawTestRecordData{
		{
			RecordID:    1,
			RewardsAddr: accAddr,
			Rewards:     sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)),
		},
		{
			RecordID:    2,
			RewardsAddr: accAddr,
			Rewards:     sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)),
		},
		{
			RecordID:    3,
			RewardsAddr: accAddr,
			Rewards:     sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150)),
		},
	}

	// Invalid inputs
	t.Run("Fail: limit is GT MaxWithdrawRecords", func(t *testing.T) {
		_, _, err := k.WithdrawRewardsByRecordsLimit(ctx, accAddr, rewardsTypes.MaxWithdrawRecordsParamLimit+1)
		require.ErrorIs(t, err, rewardsTypes.ErrInvalidRequest)
	})

	// Withdraw nothing
	t.Run("OK: withdraw empty rewards", func(t *testing.T) {
		totalRewardsReceived, recordsUsedReceived, err := k.WithdrawRewardsByRecordsLimit(ctx, accAddr, 1000)
		require.NoError(t, err)
		require.Empty(t, totalRewardsReceived)
		require.Empty(t, recordsUsedReceived)
	})

	// Setup environment
	err := SetupWithdrawTest(k, ctx, testData)
	require.NoError(t, err)
	_, err = rewards.EndBlocker(ctx, k)
	require.NoError(t, err)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// Withdraw the 1st half
	t.Run("OK: withdraw 1st half", func(t *testing.T) {
		CheckWithdrawResults(t, k, ctx,
			accAddr, testData[:2],
			func() (sdk.Coins, int, error) {
				return k.WithdrawRewardsByRecordsLimit(ctx, accAddr, 2)
			},
		)
	})

	// Withdraw the rest
	t.Run("OK: withdraw 2nd half", func(t *testing.T) {
		CheckWithdrawResults(t, k, ctx,
			accAddr, testData[2:],
			func() (sdk.Coins, int, error) {
				return k.WithdrawRewardsByRecordsLimit(ctx, accAddr, 0)
			},
		)
	})
}

// TestWithdrawRewardsByIDs tests the withdraw operation using record IDs mode.
func TestWithdrawRewardsByIDs(t *testing.T) {
	k, ctx, _ := testutils.RewardsKeeper(t)
	accAddr1, accAddr2 := testutils.AccAddress(), testutils.AccAddress()

	testData := []withdrawTestRecordData{
		{
			RecordID:    1,
			RewardsAddr: accAddr1,
			Rewards:     sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)),
		},
		{
			RecordID:    2,
			RewardsAddr: accAddr1,
			Rewards:     sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100)),
		},
		{
			RecordID:    3,
			RewardsAddr: accAddr1,
			Rewards:     sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 150)),
		},
		{
			RecordID:    4,
			RewardsAddr: accAddr2,
			Rewards:     sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 200)),
		},
	}

	// Override the default record limit
	{
		params := k.GetParams(ctx)
		params.MaxWithdrawRecords = 5
		err := k.Params.Set(ctx, params)
		require.NoError(t, err)
	}

	// Invalid inputs
	t.Run("Fail: limit is GT MaxWithdrawRecords", func(t *testing.T) {
		_, _, err := k.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{1, 2, 3, 4, 5, 6})
		require.ErrorIs(t, err, rewardsTypes.ErrInvalidRequest)
	})

	t.Run("Fail: non-existing IDs", func(t *testing.T) {
		_, _, err := k.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{1, 2, 3, 10})
		require.ErrorIs(t, err, rewardsTypes.ErrInvalidRequest)
	})

	t.Run("Fail: rewardsAddr mismatch", func(t *testing.T) {
		_, _, err := k.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{1, 2, 3, 4})
		require.ErrorIs(t, err, rewardsTypes.ErrInvalidRequest)
	})

	// Withdraw nothing
	t.Run("OK: withdraw empty rewards", func(t *testing.T) {
		totalRewardsReceived, recordsUsedReceived, err := k.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{})
		require.NoError(t, err)
		require.Empty(t, totalRewardsReceived)
		require.Empty(t, recordsUsedReceived)
	})

	// Setup environment
	err := SetupWithdrawTest(k, ctx, testData)
	require.NoError(t, err)

	// Withdraw for the 1st account
	t.Run("OK: withdraw 1st half for account1", func(t *testing.T) {
		CheckWithdrawResults(t, k, ctx,
			accAddr1, testData[:2],
			func() (sdk.Coins, int, error) {
				return k.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{1, 2})
			},
		)
	})

	t.Run("OK: withdraw 2nd half for account1", func(t *testing.T) {
		CheckWithdrawResults(t, k, ctx,
			accAddr1, testData[2:3],
			func() (sdk.Coins, int, error) {
				return k.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{3})
			},
		)
	})

	// Withdraw for the 2nd account
	t.Run("OK: withdraw all for account2", func(t *testing.T) {
		CheckWithdrawResults(t, k, ctx,
			accAddr2, testData[3:],
			func() (sdk.Coins, int, error) {
				return k.WithdrawRewardsByRecordIDs(ctx, accAddr2, []uint64{4})
			},
		)
	})
}
