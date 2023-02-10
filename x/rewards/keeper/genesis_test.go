package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/rewards/types"
)

// TestGenesisImportExport check genesis import/export.
// Test updates the initial state with new records and checks that they were merged.
func (s *KeeperTestSuite) TestGenesisImportExport() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().RewardsKeeper

	contractAddrs := e2eTesting.GenContractAddresses(2)
	accAddrs, _ := e2eTesting.GenAccounts(2)

	var genesisStateInitial types.GenesisState
	s.Run("Check export of the initial genesis", func() {
		genesisState := keeper.ExportGenesis(ctx)
		s.Require().NotNil(genesisState)

		s.Assert().Equal(types.DefaultParams(), genesisState.Params)
		s.Assert().Empty(genesisState.ContractsMetadata)
		s.Assert().NotEmpty(genesisState.BlockRewards) // height is 2 so we have some inflation rewards already
		s.Assert().Empty(genesisState.TxRewards)
		s.Assert().Empty(genesisState.RewardsRecordLastId)
		s.Assert().Empty(genesisState.RewardsRecords)
		s.Assert().Empty(genesisState.FlatFees)

		genesisStateInitial = *genesisState
	})

	newParams := types.NewParams(
		sdk.NewDecWithPrec(99, 2),
		sdk.NewDecWithPrec(98, 2),
		1001,
	)

	newMetadata := []types.ContractMetadata{
		{
			ContractAddress: contractAddrs[0].String(),
			OwnerAddress:    accAddrs[0].String(),
		},
		{
			ContractAddress: contractAddrs[1].String(),
			OwnerAddress:    accAddrs[1].String(),
			RewardsAddress:  accAddrs[1].String(),
		},
	}

	newBlockRewards := []types.BlockRewards{
		{
			Height:           100,
			InflationRewards: sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(100)},
			MaxGas:           1000,
		},
		{
			Height:           200,
			InflationRewards: sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(200)},
			MaxGas:           2000,
		},
	}

	newTxRewards := []types.TxRewards{
		{
			TxId:   110,
			Height: 100,
			FeeRewards: []sdk.Coin{
				{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(150)},
			},
		},
		{
			TxId:   210,
			Height: 200,
			FeeRewards: []sdk.Coin{
				{Denom: "uarch", Amount: sdk.NewInt(250)},
			},
		},
	}

	newMinConsFee := sdk.NewDecCoin("uarch", sdk.NewInt(100))

	newRewardsRecords := []types.RewardsRecord{
		{
			Id:               1,
			RewardsAddress:   accAddrs[0].String(),
			Rewards:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100))),
			CalculatedHeight: ctx.BlockHeight(),
			CalculatedTime:   ctx.BlockTime(),
		},
		{
			Id:               2,
			RewardsAddress:   accAddrs[1].String(),
			Rewards:          sdk.NewCoins(sdk.NewCoin("uarch", sdk.NewInt(1))),
			CalculatedHeight: ctx.BlockHeight() + 1,
			CalculatedTime:   ctx.BlockTime().Add(5 * time.Second),
		},
	}

	newFlatFees := []types.FlatFee{
		{
			ContractAddress: contractAddrs[0].String(),
			FlatFee:         sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
		},
		{
			ContractAddress: contractAddrs[1].String(),
			FlatFee:         sdk.NewCoin("uarch", sdk.NewInt(1)),
		},
	}

	genesisStateImported := types.NewGenesisState(
		newParams,
		newMetadata,
		newBlockRewards,
		newTxRewards,
		newMinConsFee,
		newRewardsRecords[len(newRewardsRecords)-1].Id,
		newRewardsRecords,
		newFlatFees,
	)
	s.Run("Check import of an updated genesis", func() {
		keeper.InitGenesis(ctx, genesisStateImported)

		genesisStateExpected := types.GenesisState{
			Params:              newParams,
			ContractsMetadata:   append(genesisStateInitial.ContractsMetadata, newMetadata...),
			BlockRewards:        append(genesisStateInitial.BlockRewards, newBlockRewards...),
			TxRewards:           append(genesisStateInitial.TxRewards, newTxRewards...),
			MinConsensusFee:     newMinConsFee,
			RewardsRecordLastId: newRewardsRecords[len(newRewardsRecords)-1].Id,
			RewardsRecords:      append(genesisStateInitial.RewardsRecords, newRewardsRecords...),
			FlatFees:            append(genesisStateInitial.FlatFees, newFlatFees...),
		}

		genesisStateReceived := keeper.ExportGenesis(ctx)
		s.Require().NotNil(genesisStateReceived)
		s.Assert().Equal(genesisStateExpected.Params, genesisStateReceived.Params)
		s.Assert().ElementsMatch(genesisStateExpected.ContractsMetadata, genesisStateReceived.ContractsMetadata)
		s.Assert().ElementsMatch(genesisStateExpected.BlockRewards, genesisStateReceived.BlockRewards)
		s.Assert().ElementsMatch(genesisStateExpected.TxRewards, genesisStateReceived.TxRewards)
		s.Assert().Equal(genesisStateExpected.MinConsensusFee.String(), genesisStateReceived.MinConsensusFee.String())
		s.Assert().Equal(genesisStateExpected.RewardsRecordLastId, genesisStateReceived.RewardsRecordLastId)
		s.Assert().ElementsMatch(genesisStateExpected.RewardsRecords, genesisStateReceived.RewardsRecords)
		s.Assert().ElementsMatch(genesisStateExpected.FlatFees, genesisStateReceived.FlatFees)
	})
}
