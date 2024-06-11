package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards"
	"github.com/archway-network/archway/x/rewards/keeper"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

type KeeperTestSuite struct {
	suite.Suite

	keeper     keeper.Keeper
	ctx        sdk.Context
	bankKeeper keeper.BankKeeperExpected
	wasmKeeper testutils.MockContractViewer
}

// withdrawTestRecordData is a helper struct to store RewardsRecord data for Withdraw tests.
type withdrawTestRecordData struct {
	RecordID    uint64         // expected recordID to be created
	RewardsAddr sdk.AccAddress // rewards address
	Rewards     sdk.Coins      // record rewards
}

func (s *KeeperTestSuite) SetupTest() {
	keeper, ctx, bk, wk := testutils.RewardsKeeper(s.T())
	s.keeper = keeper
	s.ctx = ctx
	s.bankKeeper = bk
	s.wasmKeeper = wk
}

// SetupWithdrawTest is a helper function to setup the test environment for Withdraw tests.
func (s *KeeperTestSuite) SetupWithdrawTest(testData []withdrawTestRecordData) {
	// Create test records
	for _, testRecord := range testData {
		// // Get rid of the current inflationary rewards for the current block (otherwise the invariant fails)
		// blockRewards, err := s.keeper.BlockRewards.Get(s.ctx, uint64(s.ctx.BlockHeight()))
		// s.Require().NoError(err)
		// s.Require().NoError(s.bankKeeper.SendCoinsFromModuleToModule(s.ctx, rewardsTypes.ContractRewardCollector, rewardsTypes.TreasuryCollector, sdk.Coins{blockRewards.InflationRewards}))

		err := s.keeper.BlockRewards.Set(s.ctx, uint64(s.ctx.BlockHeight()), rewardsTypes.BlockRewards{
			Height:           s.ctx.BlockHeight(),
			InflationRewards: sdk.NewCoin(sdk.DefaultBondDenom, math.ZeroInt()),
			MaxGas:           0,
		})
		s.Require().NoError(err)

		// // Mint rewards for the current record
		// rewardsToMint := testRecord.Rewards
		// s.Require().NoError(keepers.MintKeeper.MintCoins(s.ctx, rewardsToMint))
		// s.Require().NoError(s.bankKeeper.SendCoinsFromModuleToModule(s.ctx, mintTypes.ModuleName, rewardsTypes.ContractRewardCollector, rewardsToMint))

		// Create the record
		_, err = s.keeper.CreateRewardsRecord(
			s.ctx,
			testRecord.RewardsAddr,
			testRecord.Rewards,
			s.ctx.BlockHeight(), s.ctx.BlockTime(),
		)
		s.Require().NoError(err)

		// Switch to the next block
		rewards.EndBlocker(s.ctx, s.keeper)
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
	}
}

// CheckWithdrawResults is a helper function to check the results of a Withdraw operation.
func (s *KeeperTestSuite) CheckWithdrawResults(rewardsAddr sdk.AccAddress, recordsUsed []withdrawTestRecordData, withdraw func() (sdk.Coins, int, error)) {
	// Estimate the expected rewards amount and get the current account balance
	totalRewardsExpected := sdk.NewCoins()
	for _, testRecord := range recordsUsed {
		totalRewardsExpected = totalRewardsExpected.Add(testRecord.Rewards...)
	}

	accBalanceBefore := s.bankKeeper.GetAllBalances(s.ctx, rewardsAddr)

	// Withdraw and check the output
	totalRewardsReceived, recordsUsedReceived, err := withdraw()
	s.Require().NoError(err)
	s.Assert().Equal(totalRewardsExpected.String(), totalRewardsReceived.String())
	s.Assert().EqualValues(len(recordsUsed), recordsUsedReceived)

	// Check the account balance diff
	accBalanceAfter := s.bankKeeper.GetAllBalances(s.ctx, rewardsAddr)
	s.Assert().Equal(totalRewardsExpected.String(), accBalanceAfter.Sub(accBalanceBefore...).String())

	// Check records pruning
	for _, testRecord := range recordsUsed {
		_, err := s.keeper.RewardsRecords.Get(s.ctx, testRecord.RecordID)
		s.ErrorIs(err, collections.ErrNotFound)
	}
}

func TestRewardsKeeper(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
