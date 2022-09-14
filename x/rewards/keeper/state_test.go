package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/rewards/types"
)

// TestStates tests ContractMetadata, BlockRewards, TxRewards and RewardsRecord state storages.
// Test append multiple objects for different blocks to make sure there are no namespace
// collisions (prefixed store keys) and state indexes work as expected.
// Final test stage is the cascade delete of reward objects.
func (s *KeeperTestSuite) TestStates() {
	type testBlockData struct {
		BlockRewards types.BlockRewards
		TxRewards    []types.TxRewards
	}

	type testData struct {
		Metadata       []types.ContractMetadata
		Blocks         []testBlockData
		RewardsRecords []types.RewardsRecord
	}

	chain := s.chain
	ctx, keeper := chain.GetContext(), chain.GetApp().RewardsKeeper
	metaState := keeper.GetState().ContractMetadataState(ctx)
	blockState := keeper.GetState().BlockRewardsState(ctx)
	txState := keeper.GetState().TxRewardsState(ctx)
	rewardsRecordState := keeper.GetState().RewardsRecord(ctx)

	// Fixtures
	startBlock, startTime := ctx.BlockHeight(), ctx.BlockTime()
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
		RewardsRecords: []types.RewardsRecord{
			{
				Id:               1,
				RewardsAddress:   accAddrs[0].String(),
				Rewards:          []sdk.Coin{coin1},
				CalculatedHeight: startBlock,
				CalculatedTime:   startTime,
			},
			{
				Id:               2,
				RewardsAddress:   accAddrs[1].String(),
				Rewards:          []sdk.Coin{coin1},
				CalculatedHeight: startBlock + 1,
				CalculatedTime:   startTime.Add(5 * time.Second),
			},
			{
				Id:               3,
				RewardsAddress:   accAddrs[1].String(),
				Rewards:          []sdk.Coin{coin2},
				CalculatedHeight: startBlock + 1,
				CalculatedTime:   startTime.Add(5 * time.Second),
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
	rewardsRecordState.Import(
		testDataExpected.RewardsRecords[len(testDataExpected.RewardsRecords)-1].Id,
		testDataExpected.RewardsRecords,
	)

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
	s.Run("Check non-existing RewardsRecord", func() {
		_, recordFound := rewardsRecordState.GetRewardsRecord(10)
		s.Assert().False(recordFound)
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

		for i, recordExpected := range testDataExpected.RewardsRecords {
			recordReceived, found := rewardsRecordState.GetRewardsRecord(recordExpected.Id)
			s.Require().True(found, "RewardsRecord [%d]: not found", i)
			s.Assert().Equal(recordExpected, recordReceived, "RewardsRecord [%d]: wrong value", i)
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

		// 3rd block (non-existing)
		{
			height := testDataExpected.Blocks[1].BlockRewards.Height + 1

			s.Assert().Empty(txState.GetTxRewardsByBlock(height))
		}
	})

	// Check RewardsRecord search via RewardsAddress index
	s.Run("Check RewardsRecord RewardsAddress index", func() {
		// 1st address
		{
			addr := accAddrs[0]
			recordExpected := testDataExpected.RewardsRecords[:1]

			recordReceived := rewardsRecordState.GetRewardsRecordByRewardsAddress(addr)
			s.Assert().ElementsMatch(recordExpected, recordReceived, "RewardsRecordsByAddress (%s): wrong value", addr)
		}

		// 2nd address
		{
			addr := accAddrs[1]
			recordExpected := testDataExpected.RewardsRecords[1:3]

			recordReceived := rewardsRecordState.GetRewardsRecordByRewardsAddress(addr)
			s.Assert().ElementsMatch(recordExpected, recordReceived, "RewardsRecordsByAddress (%s): wrong value", addr)
		}

		// 3rd address (non-existing)
		{
			addr := accAddrs[2]

			s.Assert().Empty(rewardsRecordState.GetRewardsRecordByRewardsAddress(addr))
		}
	})

	// Check RewardsRecord search via RewardsAddress index with pagination
	// We don't cover all the pagination cases here because the pagination is tested already
	s.Run("Check RewardsRecord RewardsAddress index with pagination", func() {
		// 2nd address
		addr := accAddrs[1]

		// Limit 1
		{
			page := &query.PageRequest{
				Limit:      1,
				CountTotal: true,
			}
			recordExpected := testDataExpected.RewardsRecords[1:2]

			recordReceived, pageResp, err := rewardsRecordState.GetRewardsRecordByRewardsAddressPaginated(addr, page)
			s.Require().NoError(err)
			s.Assert().ElementsMatch(recordExpected, recordReceived)

			s.Require().NotNil(pageResp)
			s.Assert().NotNil(pageResp.NextKey)
			s.Assert().EqualValues(2, pageResp.Total)
		}

		// Limit 1, Offset 1
		{
			page := &query.PageRequest{
				Offset:     1,
				Limit:      1,
				CountTotal: true,
			}
			recordExpected := testDataExpected.RewardsRecords[2:3]

			recordReceived, pageResp, err := rewardsRecordState.GetRewardsRecordByRewardsAddressPaginated(addr, page)
			s.Require().NoError(err)
			s.Assert().ElementsMatch(recordExpected, recordReceived)

			s.Require().NotNil(pageResp)
			s.Assert().Nil(pageResp.NextKey)
			s.Assert().EqualValues(2, pageResp.Total)
		}

		// Limit 1, Using NextKey
		{
			page := &query.PageRequest{
				Limit: 1,
			}
			recordExpected := testDataExpected.RewardsRecords[2:3]

			_, pageResp, err := rewardsRecordState.GetRewardsRecordByRewardsAddressPaginated(addr, page)
			s.Require().NoError(err)
			s.Require().NotNil(pageResp)
			s.Assert().NotNil(pageResp.NextKey)

			page.Key = pageResp.NextKey
			recordReceived, pageResp, err := rewardsRecordState.GetRewardsRecordByRewardsAddressPaginated(addr, page)
			s.Require().NoError(err)
			s.Require().NotNil(pageResp)
			s.Assert().Nil(pageResp.NextKey)

			s.Assert().ElementsMatch(recordExpected, recordReceived)
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

	// Check records removal
	s.Run("Check rewards records removal", func() {
		rewardsRecordState.DeleteRewardsRecords(testDataExpected.RewardsRecords...)

		for i, recordExpected := range testDataExpected.RewardsRecords {
			_, found := rewardsRecordState.GetRewardsRecord(recordExpected.Id)
			s.Assert().False(found, "RewardsRecord [%d]: found", i)
		}
	})
}
