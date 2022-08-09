package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/rewards/types"
)

// TestStates tests ContractMetadata, BlockRewards and TxRewards state storages.
// Test append multiple objects for different blocks to make sure there are no namespace
// collisions (prefixed store keys) and state indexes work as expected.
// Final test stage is the cascade delete of reward objects.
func (s *KeeperTestSuite) TestStates() {
	type testBlockData struct {
		BlockRewards types.BlockRewards
		TxRewards    []types.TxRewards
	}

	type testData struct {
		Metadata []types.ContractMetadata
		Blocks   []testBlockData
	}

	chain := s.chain
	ctx, keeper := chain.GetContext(), chain.GetApp().RewardsKeeper
	metaState, blockState, txState := keeper.GetState().ContractMetadataState(ctx), keeper.GetState().BlockRewardsState(ctx), keeper.GetState().TxRewardsState(ctx)

	// Fixtures
	startBlock := ctx.BlockHeight()
	contractAddrs := e2eTesting.GenContractAddresses(3)
	accAddrs, _ := e2eTesting.GenAccounts(3)
	coin1 := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(100)}
	coin2 := sdk.Coin{Denom: "uarch", Amount: sdk.NewInt(200)}

	testDataExpected := testData{
		Metadata: []types.ContractMetadata{
			// Metadata 1
			{
				ContractAddress: contractAddrs[0].String(),
				OwnerAddress:    accAddrs[0].String(),
			},
			// Metadata 2
			{
				ContractAddress: contractAddrs[1].String(),
				OwnerAddress:    accAddrs[1].String(),
				RewardsAddress:  accAddrs[2].String(),
			},
		},
		Blocks: []testBlockData{
			// Block 1 (no gas, no rewards)
			{
				BlockRewards: types.BlockRewards{
					Height: startBlock,
				},
				TxRewards: []types.TxRewards{
					// Tx 1 (no rewards)
					{
						TxId:   1,
						Height: startBlock,
					},
					// Tx 2
					{
						TxId:       2,
						Height:     startBlock,
						FeeRewards: []sdk.Coin{coin1},
					},
				},
			},
			// Block 2
			{
				BlockRewards: types.BlockRewards{
					Height:           startBlock + 1,
					InflationRewards: sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(100)},
					MaxGas:           1000,
				},
				TxRewards: []types.TxRewards{
					// Tx 3
					{
						TxId:       3,
						Height:     startBlock + 1,
						FeeRewards: []sdk.Coin{coin1, coin2},
					},
				},
			},
		},
	}

	// Upload fixtures
	for _, metadata := range testDataExpected.Metadata {
		metaState.SetContractMetadata(metadata.MustGetContractAddress(), metadata)
	}
	for _, blockData := range testDataExpected.Blocks {
		blockRewards := blockData.BlockRewards
		blockState.CreateBlockRewards(blockRewards.Height, blockRewards.InflationRewards, blockRewards.MaxGas)

		for _, txRewards := range blockData.TxRewards {
			txState.CreateTxRewards(txRewards.TxId, txRewards.Height, txRewards.FeeRewards)
		}
	}

	// Check non-existing records
	s.Run("Check non-existing metadata record", func() {
		_, metaFound := metaState.GetContractMetadata(contractAddrs[2])
		s.Assert().False(metaFound)
	})
	s.Run("Check non-existing BlockRewards and TxRewards records", func() {
		_, blockRewardsFound := blockState.GetBlockRewards(startBlock + 10)
		s.Assert().False(blockRewardsFound)

		_, txRewardsFound := txState.GetTxRewards(10)
		s.Assert().False(txRewardsFound)
	})

	// Check that the states are as expected
	s.Run("Check objects one by one", func() {
		for _, metadataExpected := range testDataExpected.Metadata {
			metaReceived, found := metaState.GetContractMetadata(metadataExpected.MustGetContractAddress())
			s.Require().True(found)
			s.Assert().Equal(metadataExpected, metaReceived)
		}

		for i, blockData := range testDataExpected.Blocks {
			blockRewardsExpected := blockData.BlockRewards
			blockRewardsReceived, found := blockState.GetBlockRewards(blockRewardsExpected.Height)
			s.Require().True(found, "BlockRewards [%d]: not found", i)

			// Modify the expected coin because proto.Unmarshal creates a coin with zero amount (not nil)
			if blockRewardsExpected.InflationRewards.Amount.IsNil() {
				blockRewardsExpected.InflationRewards.Amount = sdk.ZeroInt()
			}
			s.Assert().Equal(blockRewardsExpected, blockRewardsReceived, "BlockRewards [%d]: wrong value", i)

			for j, txRewardsExpected := range blockData.TxRewards {
				txRewardsReceived, found := txState.GetTxRewards(txRewardsExpected.TxId)
				s.Require().True(found, "TxRewards [%d][%d]: not found", i, j)
				s.Assert().Equal(txRewardsExpected, txRewardsReceived, "TxRewards [%d][%d]: wrong value", i, j)
			}
		}
	})

	// Check TxRewards search via block index
	s.Run("Check TxRewards block index", func() {
		// 1st block
		{
			height := testDataExpected.Blocks[0].BlockRewards.Height
			txRewardsExpected := testDataExpected.Blocks[0].TxRewards

			txRewardsReceived := txState.GetTxRewardsByBlock(height)
			s.Assert().ElementsMatch(txRewardsExpected, txRewardsReceived, "TxRewardsByBlock (%d): wrong value", height)
		}

		// 2nd block
		{
			height := testDataExpected.Blocks[1].BlockRewards.Height
			txRewardsExpected := testDataExpected.Blocks[1].TxRewards

			txRewardsReceived := txState.GetTxRewardsByBlock(height)
			s.Assert().ElementsMatch(txRewardsExpected, txRewardsReceived, "TxRewardsByBlock (%d): wrong value", height)
		}
	})

	// Check rewards removal
	s.Run("Check rewards removal for the 1st block", func() {
		height1, height2 := testDataExpected.Blocks[0].BlockRewards.Height, testDataExpected.Blocks[1].BlockRewards.Height
		txs2 := testDataExpected.Blocks[1].TxRewards

		keeper.GetState().DeleteBlockRewardsCascade(ctx, height1)

		block1Txs := txState.GetTxRewardsByBlock(height1)
		s.Assert().Empty(block1Txs)

		block2Txs := txState.GetTxRewardsByBlock(height2)
		s.Assert().Len(block2Txs, len(txs2))

		_, block1Found := blockState.GetBlockRewards(height1)
		s.Assert().False(block1Found)

		_, block2Found := blockState.GetBlockRewards(height2)
		s.Assert().True(block2Found)
	})

	s.Run("Check rewards removal for the 2nd block", func() {
		height2 := testDataExpected.Blocks[1].BlockRewards.Height

		keeper.GetState().DeleteBlockRewardsCascade(ctx, height2)

		block2Txs := txState.GetTxRewardsByBlock(height2)
		s.Assert().Empty(block2Txs)

		_, block2Found := blockState.GetBlockRewards(height2)
		s.Assert().False(block2Found)
	})
}
