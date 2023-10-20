package keeper_test

import (
	"testing"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

type KeeperTestSuite struct {
	suite.Suite

	chain *e2eTesting.TestChain
}

// withdrawTestRecordData is a helper struct to store RewardsRecord data for Withdraw tests.
type withdrawTestRecordData struct {
	RecordID    uint64         // expected recordID to be created
	RewardsAddr sdk.AccAddress // rewards address
	Rewards     sdk.Coins      // record rewards
}

func (s *KeeperTestSuite) SetupTest() {
	s.chain = e2eTesting.NewTestChain(s.T(), 1)
}

// SetupWithdrawTest is a helper function to setup the test environment for Withdraw tests.
func (s *KeeperTestSuite) SetupWithdrawTest(testData []withdrawTestRecordData) {
	// Create test records
	for _, testRecord := range testData {
		ctx := s.chain.GetContext()
		keepers := s.chain.GetApp().Keepers

		// Get rid of the current inflationary rewards for the current block (otherwise the invariant fails)
		blockRewards, err := keepers.RewardsKeeper.BlockRewards.Get(ctx, uint64(ctx.BlockHeight()))
		s.Require().NoError(err)
		s.Require().NoError(keepers.BankKeeper.SendCoinsFromModuleToModule(ctx, rewardsTypes.ContractRewardCollector, rewardsTypes.TreasuryCollector, sdk.Coins{blockRewards.InflationRewards}))

		err = keepers.RewardsKeeper.BlockRewards.Set(ctx, uint64(ctx.BlockHeight()), rewardsTypes.BlockRewards{
			Height:           ctx.BlockHeight(),
			InflationRewards: sdk.NewCoin(sdk.DefaultBondDenom, sdk.ZeroInt()),
			MaxGas:           0,
		})
		s.Require().NoError(err)

		// Mint rewards for the current record
		rewardsToMint := testRecord.Rewards
		s.Require().NoError(keepers.MintKeeper.MintCoins(ctx, rewardsToMint))
		s.Require().NoError(keepers.BankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, rewardsTypes.ContractRewardCollector, rewardsToMint))

		// Create the record
		keepers.RewardsKeeper.GetState().RewardsRecord(ctx).
			CreateRewardsRecord(
				testRecord.RewardsAddr,
				testRecord.Rewards,
				ctx.BlockHeight(), ctx.BlockTime(),
			)

		// Switch to the next block
		s.chain.NextBlock(0)
	}
}

// CheckWithdrawResults is a helper function to check the results of a Withdraw operation.
func (s *KeeperTestSuite) CheckWithdrawResults(rewardsAddr sdk.AccAddress, recordsUsed []withdrawTestRecordData, withdraw func() (sdk.Coins, int, error)) {
	// Estimate the expected rewards amount and get the current account balance
	totalRewardsExpected := sdk.NewCoins()
	for _, testRecord := range recordsUsed {
		totalRewardsExpected = totalRewardsExpected.Add(testRecord.Rewards...)
	}

	accBalanceBefore := s.chain.GetBalance(rewardsAddr)

	// Withdraw and check the output
	totalRewardsReceived, recordsUsedReceived, err := withdraw()
	s.Require().NoError(err)
	s.Assert().Equal(totalRewardsExpected.String(), totalRewardsReceived.String())
	s.Assert().EqualValues(len(recordsUsed), recordsUsedReceived)

	// Check the account balance diff
	accBalanceAfter := s.chain.GetBalance(rewardsAddr)
	s.Assert().Equal(totalRewardsExpected.String(), accBalanceAfter.Sub(accBalanceBefore...).String())

	// Check records pruning
	for _, testRecord := range recordsUsed {
		_, err := s.chain.GetApp().Keepers.RewardsKeeper.RewardsRecords.Get(s.chain.GetContext(), testRecord.RecordID)
		s.ErrorIs(err, collections.ErrNotFound)
	}
}

func TestRewardsKeeper(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
