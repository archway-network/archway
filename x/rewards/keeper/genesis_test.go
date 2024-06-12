package keeper_test

import (
	"testing"
	"time"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards/types"
)

// TestGenesisImportExport check genesis import/export.
// Test updates the initial state with new records and checks that they were merged.
func TestGenesisImportExport(t *testing.T) {
	k, ctx, _, _ := testutils.RewardsKeeper(t)
	contractAddrs := e2eTesting.GenContractAddresses(2)
	accAddrs, _ := e2eTesting.GenAccounts(2)

	var genesisStateInitial types.GenesisState
	t.Run("Check export of the initial genesis", func(t *testing.T) {
		genesisState := k.ExportGenesis(ctx)
		require.NotNil(t, genesisState)

		require.Equal(t, types.DefaultParams(), genesisState.Params)
		require.Empty(t, genesisState.ContractsMetadata)
		require.Empty(t, genesisState.BlockRewards)
		require.Empty(t, genesisState.TxRewards)
		require.Empty(t, genesisState.RewardsRecordLastId)
		require.Empty(t, genesisState.RewardsRecords)
		require.Empty(t, genesisState.FlatFees)

		genesisStateInitial = *genesisState
	})

	newParams := types.NewParams(
		math.LegacyNewDecWithPrec(99, 2),
		math.LegacyNewDecWithPrec(98, 2),
		1001,
		types.DefaultMinPriceOfGas,
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
			InflationRewards: sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(100)},
			MaxGas:           1000,
		},
		{
			Height:           200,
			InflationRewards: sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(200)},
			MaxGas:           2000,
		},
	}

	newTxRewards := []types.TxRewards{
		{
			TxId:   110,
			Height: 100,
			FeeRewards: []sdk.Coin{
				{Denom: sdk.DefaultBondDenom, Amount: math.NewInt(150)},
			},
		},
		{
			TxId:   210,
			Height: 200,
			FeeRewards: []sdk.Coin{
				{Denom: "uarch", Amount: math.NewInt(250)},
			},
		},
	}

	newMinConsFee := sdk.NewDecCoin("uarch", math.NewInt(100))

	newRewardsRecords := []types.RewardsRecord{
		{
			Id:               1,
			RewardsAddress:   accAddrs[0].String(),
			Rewards:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100))),
			CalculatedHeight: ctx.BlockHeight(),
			CalculatedTime:   ctx.BlockTime(),
		},
		{
			Id:               2,
			RewardsAddress:   accAddrs[1].String(),
			Rewards:          sdk.NewCoins(sdk.NewCoin("uarch", math.NewInt(1))),
			CalculatedHeight: ctx.BlockHeight() + 1,
			CalculatedTime:   ctx.BlockTime().Add(5 * time.Second),
		},
	}

	newFlatFees := []types.FlatFee{
		{
			ContractAddress: contractAddrs[0].String(),
			FlatFee:         sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(100)),
		},
		{
			ContractAddress: contractAddrs[1].String(),
			FlatFee:         sdk.NewCoin("uarch", math.NewInt(1)),
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
	t.Run("Check import of an updated genesis", func(t *testing.T) {
		k.InitGenesis(ctx, genesisStateImported)

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

		genesisStateReceived := k.ExportGenesis(ctx)
		require.NotNil(t, genesisStateReceived)
		require.Equal(t, genesisStateExpected.Params, genesisStateReceived.Params)
		require.ElementsMatch(t, genesisStateExpected.ContractsMetadata, genesisStateReceived.ContractsMetadata)
		require.ElementsMatch(t, genesisStateExpected.BlockRewards, genesisStateReceived.BlockRewards)
		require.ElementsMatch(t, genesisStateExpected.TxRewards, genesisStateReceived.TxRewards)
		require.Equal(t, genesisStateExpected.MinConsensusFee.String(), genesisStateReceived.MinConsensusFee.String())
		require.Equal(t, genesisStateExpected.RewardsRecordLastId, genesisStateReceived.RewardsRecordLastId)
		require.ElementsMatch(t, genesisStateExpected.RewardsRecords, genesisStateReceived.RewardsRecords)
		require.ElementsMatch(t, genesisStateExpected.FlatFees, genesisStateReceived.FlatFees)
	})
}
