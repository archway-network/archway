package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// TestWithdrawRewardsByLimit tests the withdraw operation using record limit mode.
func (s *KeeperTestSuite) TestWithdrawRewardsByLimit() {
	keeper := s.chain.GetApp().RewardsKeeper
	accAddr := s.chain.GetAccount(0).Address

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
	s.Run("Fail: limit is GT MaxWithdrawRecords", func() {
		ctx := s.chain.GetContext()
		_, _, err := keeper.WithdrawRewardsByRecordsLimit(ctx, accAddr, rewardsTypes.MaxWithdrawRecordsParamLimit+1)
		s.Assert().ErrorIs(err, rewardsTypes.ErrInvalidRequest)
	})

	// Withdraw nothing
	s.Run("OK: withdraw empty rewards", func() {
		ctx := s.chain.GetContext()
		totalRewardsReceived, recordsUsedReceived, err := keeper.WithdrawRewardsByRecordsLimit(ctx, accAddr, 1000)
		s.Require().NoError(err)
		s.Assert().Empty(totalRewardsReceived)
		s.Assert().Empty(recordsUsedReceived)
	})

	// Setup environment
	s.SetupWithdrawTest(testData)

	// Withdraw the 1st half
	s.Run("OK: withdraw 1st half", func() {
		s.CheckWithdrawResults(
			accAddr, testData[:2],
			func() (sdk.Coins, int, error) {
				ctx := s.chain.GetContext()
				return keeper.WithdrawRewardsByRecordsLimit(ctx, accAddr, 2)
			},
		)
	})

	// Withdraw the rest
	s.Run("OK: withdraw 2nd half", func() {
		s.CheckWithdrawResults(
			accAddr, testData[2:],
			func() (sdk.Coins, int, error) {
				ctx := s.chain.GetContext()
				return keeper.WithdrawRewardsByRecordsLimit(ctx, accAddr, 0)
			},
		)
	})
}

// TestWithdrawRewardsByIDs tests the withdraw operation using record IDs mode.
func (s *KeeperTestSuite) TestWithdrawRewardsByIDs() {
	keeper := s.chain.GetApp().RewardsKeeper
	accAddr1, accAddr2 := s.chain.GetAccount(0).Address, s.chain.GetAccount(1).Address

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
		ctx := s.chain.GetContext()
		params := keeper.GetParams(ctx)
		params.MaxWithdrawRecords = 5
		keeper.SetParams(ctx, params)
	}

	// Invalid inputs
	s.Run("Fail: limit is GT MaxWithdrawRecords", func() {
		ctx := s.chain.GetContext()
		_, _, err := keeper.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{1, 2, 3, 4, 5, 6})
		s.Assert().ErrorIs(err, rewardsTypes.ErrInvalidRequest)
	})

	s.Run("Fail: non-existing IDs", func() {
		ctx := s.chain.GetContext()
		_, _, err := keeper.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{1, 2, 3, 10})
		s.Assert().ErrorIs(err, rewardsTypes.ErrInvalidRequest)
	})

	s.Run("Fail: rewardsAddr mismatch", func() {
		ctx := s.chain.GetContext()
		_, _, err := keeper.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{1, 2, 3, 4})
		s.Assert().ErrorIs(err, rewardsTypes.ErrInvalidRequest)
	})

	// Withdraw nothing
	s.Run("OK: withdraw empty rewards", func() {
		ctx := s.chain.GetContext()
		totalRewardsReceived, recordsUsedReceived, err := keeper.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{})
		s.Require().NoError(err)
		s.Assert().Empty(totalRewardsReceived)
		s.Assert().Empty(recordsUsedReceived)
	})

	// Setup environment
	s.SetupWithdrawTest(testData)

	// Withdraw for the 1st account
	s.Run("OK: withdraw 1st half for account1", func() {
		s.CheckWithdrawResults(
			accAddr1, testData[:2],
			func() (sdk.Coins, int, error) {
				ctx := s.chain.GetContext()
				return keeper.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{1, 2})
			},
		)
	})

	s.Run("OK: withdraw 2nd half for account1", func() {
		s.CheckWithdrawResults(
			accAddr1, testData[2:3],
			func() (sdk.Coins, int, error) {
				ctx := s.chain.GetContext()
				return keeper.WithdrawRewardsByRecordIDs(ctx, accAddr1, []uint64{3})
			},
		)
	})

	// Withdraw for the 2nd account
	s.Run("OK: withdraw all for account2", func() {
		s.CheckWithdrawResults(
			accAddr2, testData[3:],
			func() (sdk.Coins, int, error) {
				ctx := s.chain.GetContext()
				return keeper.WithdrawRewardsByRecordIDs(ctx, accAddr2, []uint64{4})
			},
		)
	})
}
