package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards/keeper"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper     keeper.Keeper
	ctx        sdk.Context
	bankKeeper keeper.BankKeeperExpected
}

// withdrawTestRecordData is a helper struct to store RewardsRecord data for Withdraw tests.
type withdrawTestRecordData struct {
	RecordID    uint64         // expected recordID to be created
	RewardsAddr sdk.AccAddress // rewards address
	Rewards     sdk.Coins      // record rewards
}

func (s *KeeperTestSuite) SetupTest() {
	keeper, ctx, bk := testutils.RewardsKeeper(s.T())
	s.keeper = keeper
	s.ctx = ctx
	s.bankKeeper = bk
}

// SetupWithdrawTest is a helper function to setup the test environment for Withdraw tests.
func SetupWithdrawTest(k keeper.Keeper, ctx sdk.Context, testData []withdrawTestRecordData) error {
	// Create test records
	for _, testRecord := range testData {
		// // Get rid of the current inflationary rewards for the current block (otherwise the invariant fails)
		// blockRewards, err := s.keeper.BlockRewards.Get(s.ctx, uint64(s.ctx.BlockHeight()))
		// s.Require().NoError(err)
		// s.Require().NoError(s.bankKeeper.SendCoinsFromModuleToModule(s.ctx, rewardsTypes.ContractRewardCollector, rewardsTypes.TreasuryCollector, sdk.Coins{blockRewards.InflationRewards}))

		err := k.BlockRewards.Set(ctx, uint64(ctx.BlockHeight()), rewardsTypes.BlockRewards{
			Height:           ctx.BlockHeight(),
			InflationRewards: sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
			MaxGas:           0,
		})
		if err != nil {
			return err
		}

		// // Mint rewards for the current record
		// rewardsToMint := testRecord.Rewards
		// s.Require().NoError(keepers.MintKeeper.MintCoins(s.ctx, rewardsToMint))
		// s.Require().NoError(s.bankKeeper.SendCoinsFromModuleToModule(s.ctx, mintTypes.ModuleName, rewardsTypes.ContractRewardCollector, rewardsToMint))

		// Create the record
		_, err = k.CreateRewardsRecord(
			ctx,
			testRecord.RewardsAddr,
			testRecord.Rewards,
			ctx.BlockHeight(), ctx.BlockTime(),
		)
		if err != nil {
			return err
		}

		// // Switch to the next block
		// rewards.EndBlocker(ctx, keeper)
		// //s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	}
	return nil
}

// CheckWithdrawResults is a helper function to check the results of a Withdraw operation.
func CheckWithdrawResults(t *testing.T, k keeper.Keeper, ctx sdk.Context, rewardsAddr sdk.AccAddress, recordsUsed []withdrawTestRecordData, withdraw func() (sdk.Coins, int, error)) {
	// // Estimate the expected rewards amount and get the current account balance
	totalRewardsExpected := sdk.NewCoins()
	for _, testRecord := range recordsUsed {
		totalRewardsExpected = totalRewardsExpected.Add(testRecord.Rewards...)
	}

	// accBalanceBefore := s.bankKeeper.GetAllBalances(s.ctx, rewardsAddr)

	// // Withdraw and check the output
	totalRewardsReceived, recordsUsedReceived, err := withdraw()
	require.NoError(t, err)
	require.Equal(t, totalRewardsExpected.String(), totalRewardsReceived.String())
	require.EqualValues(t, len(recordsUsed), recordsUsedReceived)

	// // Check the account balance diff
	// accBalanceAfter := s.bankKeeper.GetAllBalances(s.ctx, rewardsAddr)
	// s.Assert().Equal(totalRewardsExpected.String(), accBalanceAfter.Sub(accBalanceBefore...).String())

	// Check records pruning
	for _, testRecord := range recordsUsed {
		_, err := k.RewardsRecords.Get(ctx, testRecord.RecordID)
		require.ErrorIs(t, err, collections.ErrNotFound)
	}
}

func TestRewardsKeeper(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
