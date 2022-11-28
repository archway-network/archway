package rewards_test

import (
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	archPkg "github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/wasmbinding/pkg"
	"github.com/archway-network/archway/wasmbinding/rewards"
	rewardsWbTypes "github.com/archway-network/archway/wasmbinding/rewards/types"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// TestRewardsWASMBindings tests the custom querier and custom message handler for the x/rewards WASM bindings.
func TestRewardsWASMBindings(t *testing.T) {
	// Setup
	chain := e2eTesting.NewTestChain(t, 1)
	acc := chain.GetAccount(0)

	// Set mock wasmd contract info viewer to emulate a contract being deployed
	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	contractViewer := testutils.NewMockContractViewer()
	contractViewer.AddContractAdmin(contractAddr.String(), acc.Address.String())
	chain.GetApp().RewardsKeeper.SetContractInfoViewer(contractViewer)
	ctx, keeper := chain.GetContext(), chain.GetApp().RewardsKeeper

	// Create custom plugins
	queryPlugin := rewards.NewQueryHandler(keeper)
	msgPlugin := rewards.NewRewardsMsgHandler(keeper)

	// Invalid inputs
	t.Run("Invalid query input", func(t *testing.T) {
		query := rewardsWbTypes.Query{}

		_, err := queryPlugin.DispatchQuery(ctx, query)
		assert.ErrorIs(t, err, rewardsTypes.ErrInvalidRequest)
	})

	// Query empty / non-existing data
	t.Run("Query non-existing metadata", func(t *testing.T) {
		query := rewardsWbTypes.Query{
			Metadata: &rewardsWbTypes.ContractMetadataRequest{
				ContractAddress: contractAddr.String(),
			},
		}

		_, err := queryPlugin.DispatchQuery(ctx, query)
		assert.ErrorIs(t, err, rewardsTypes.ErrMetadataNotFound)
	})

	t.Run("Query empty rewards", func(t *testing.T) {
		query := rewardsWbTypes.Query{
			RewardsRecords: &rewardsWbTypes.RewardsRecordsRequest{
				RewardsAddress: contractAddr.String(),
			},
		}

		resObj, err := queryPlugin.DispatchQuery(ctx, query)
		require.NoError(t, err)

		res, ok := resObj.(rewardsWbTypes.RewardsRecordsResponse)
		require.True(t, ok)
		assert.Empty(t, res.Records)
	})

	t.Run("Update invalid metadata", func(t *testing.T) {
		msg := rewardsWbTypes.UpdateContractMetadataRequest{
			OwnerAddress: "invalid",
		}

		_, _, err := msgPlugin.UpdateContractMetadata(ctx, contractAddr, msg)
		assert.ErrorContains(t, err, "ownerAddress: parsing: decoding bech32 failed")
	})

	// Handle no-op msg
	t.Run("Update non-existing metadata (unauthorized create operation)", func(t *testing.T) {
		msg := rewardsWbTypes.UpdateContractMetadataRequest{
			OwnerAddress:   acc.Address.String(),
			RewardsAddress: acc.Address.String(),
		}

		_, _, err := msgPlugin.UpdateContractMetadata(ctx, contractAddr, msg)
		assert.ErrorIs(t, err, rewardsTypes.ErrUnauthorized)
	})

	t.Run("Withdraw invalid request", func(t *testing.T) {
		msg := rewardsWbTypes.WithdrawRewardsRequest{
			RecordsLimit: archPkg.Uint64Ptr(1000),
			RecordIDs:    []uint64{1, 0},
		}

		_, _, err := msgPlugin.WithdrawContractRewards(ctx, contractAddr, msg)
		assert.ErrorContains(t, err, "one of (RecordsLimit, RecordIDs) fields must be set")
	})

	t.Run("Withdraw empty rewards", func(t *testing.T) {
		msg := rewardsWbTypes.WithdrawRewardsRequest{
			RecordsLimit: archPkg.Uint64Ptr(1000),
		}

		_, resData, err := msgPlugin.WithdrawContractRewards(ctx, contractAddr, msg)
		require.NoError(t, err)
		require.Len(t, resData, 1)

		var res rewardsWbTypes.WithdrawRewardsResponse
		require.NoError(t, json.Unmarshal(resData[0], &res))
		assert.Empty(t, res.TotalRewards)
	})

	// Create metadata with contractAddr as owner (for a contract to be able to modify it)
	err := keeper.SetContractMetadata(ctx, acc.Address, contractAddr, rewardsTypes.ContractMetadata{
		OwnerAddress:   contractAddr.String(),
		RewardsAddress: acc.Address.String(),
	})
	require.NoError(t, err)

	// Update metadata
	t.Run("Update metadata (set contractAddr as the rewardsAddr)", func(t *testing.T) {
		msg := rewardsWbTypes.UpdateContractMetadataRequest{
			OwnerAddress:   contractAddr.String(),
			RewardsAddress: contractAddr.String(),
		}

		_, _, err := msgPlugin.UpdateContractMetadata(ctx, contractAddr, msg)
		require.NoError(t, err)
	})

	t.Run("Check metadata updated", func(t *testing.T) {
		query := rewardsWbTypes.Query{
			Metadata: &rewardsWbTypes.ContractMetadataRequest{
				ContractAddress: contractAddr.String(),
			},
		}

		resObj, err := queryPlugin.DispatchQuery(ctx, query)
		require.NoError(t, err)

		res, ok := resObj.(rewardsWbTypes.ContractMetadataResponse)
		require.True(t, ok)
		assert.Equal(t, contractAddr.String(), res.OwnerAddress)
		assert.Equal(t, contractAddr.String(), res.RewardsAddress)
	})

	// Add some rewards to withdraw (create new records and mint tokens)
	record1RewardsExpected := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 25))
	record2RewardsExpected := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 75))
	record3RewardsExpected := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	recordsRewards := record1RewardsExpected.Add(record2RewardsExpected...).Add(record3RewardsExpected...)

	keeper.GetState().RewardsRecord(ctx).CreateRewardsRecord(contractAddr, record1RewardsExpected, ctx.BlockHeight(), ctx.BlockTime())
	keeper.GetState().RewardsRecord(ctx).CreateRewardsRecord(contractAddr, record2RewardsExpected, ctx.BlockHeight(), ctx.BlockTime())
	keeper.GetState().RewardsRecord(ctx).CreateRewardsRecord(contractAddr, record3RewardsExpected, ctx.BlockHeight(), ctx.BlockTime())
	require.NoError(t, chain.GetApp().MintKeeper.MintCoins(ctx, recordsRewards))
	require.NoError(t, chain.GetApp().BankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, rewardsTypes.ContractRewardCollector, recordsRewards))

	// Query available rewards
	t.Run("Query new rewards", func(t *testing.T) {
		query := rewardsWbTypes.Query{
			RewardsRecords: &rewardsWbTypes.RewardsRecordsRequest{
				RewardsAddress: contractAddr.String(),
				Pagination: &pkg.PageRequest{
					CountTotal: true,
				},
			},
		}

		resObj, err := queryPlugin.DispatchQuery(ctx, query)
		require.NoError(t, err)

		res, ok := resObj.(rewardsWbTypes.RewardsRecordsResponse)
		require.True(t, ok)

		require.Len(t, res.Records, 3)
		// Record 1
		assert.EqualValues(t, 1, res.Records[0].ID)
		assert.Equal(t, contractAddr.String(), res.Records[0].RewardsAddress)
		assert.Equal(t, ctx.BlockHeight(), res.Records[0].CalculatedHeight)
		assert.Equal(t, ctx.BlockTime().Format(time.RFC3339Nano), res.Records[0].CalculatedTime)
		record1RewardsReceived, err := pkg.WasmCoinsToSDK(res.Records[0].Rewards)
		require.NoError(t, err)
		assert.Equal(t, record1RewardsExpected.String(), record1RewardsReceived.String())
		// Record 2
		assert.EqualValues(t, 2, res.Records[1].ID)
		assert.Equal(t, contractAddr.String(), res.Records[1].RewardsAddress)
		assert.Equal(t, ctx.BlockHeight(), res.Records[1].CalculatedHeight)
		assert.Equal(t, ctx.BlockTime().Format(time.RFC3339Nano), res.Records[1].CalculatedTime)
		record2RewardsReceived, err := pkg.WasmCoinsToSDK(res.Records[1].Rewards)
		require.NoError(t, err)
		assert.Equal(t, record2RewardsExpected.String(), record2RewardsReceived.String())
		// Record 3
		assert.EqualValues(t, 3, res.Records[2].ID)
		assert.Equal(t, contractAddr.String(), res.Records[2].RewardsAddress)
		assert.Equal(t, ctx.BlockHeight(), res.Records[2].CalculatedHeight)
		assert.Equal(t, ctx.BlockTime().Format(time.RFC3339Nano), res.Records[2].CalculatedTime)
		record3RewardsReceived, err := pkg.WasmCoinsToSDK(res.Records[2].Rewards)
		require.NoError(t, err)
		assert.Equal(t, record3RewardsExpected.String(), record3RewardsReceived.String())

		assert.EqualValues(t, 3, res.Pagination.Total)
	})

	// Withdraw rewards using the limit mode
	t.Run("Withdraw 1st reward using limit", func(t *testing.T) {
		msg := rewardsWbTypes.WithdrawRewardsRequest{
			RecordsLimit: archPkg.Uint64Ptr(1),
		}

		_, resData, err := msgPlugin.WithdrawContractRewards(ctx, contractAddr, msg)
		require.NoError(t, err)
		require.Len(t, resData, 1)

		var res rewardsWbTypes.WithdrawRewardsResponse
		require.NoError(t, json.Unmarshal(resData[0], &res))

		assert.EqualValues(t, 1, res.RecordsNum)
		totalRewardsReceived, err := pkg.WasmCoinsToSDK(res.TotalRewards)
		require.NoError(t, err)
		assert.EqualValues(t, record1RewardsExpected.String(), totalRewardsReceived.String())

		assert.Equal(t, record1RewardsExpected.String(), chain.GetBalance(contractAddr).String())
	})

	// Withdraw rewards using the record IDs mode
	t.Run("Withdraw 2nd reward using record ID", func(t *testing.T) {
		msg := rewardsWbTypes.WithdrawRewardsRequest{
			RecordIDs: []uint64{2},
		}

		_, resData, err := msgPlugin.WithdrawContractRewards(ctx, contractAddr, msg)
		require.NoError(t, err)
		require.Len(t, resData, 1)

		var res rewardsWbTypes.WithdrawRewardsResponse
		require.NoError(t, json.Unmarshal(resData[0], &res))

		assert.EqualValues(t, 1, res.RecordsNum)
		totalRewardsReceived, err := pkg.WasmCoinsToSDK(res.TotalRewards)
		require.NoError(t, err)
		assert.EqualValues(t, record2RewardsExpected.String(), totalRewardsReceived.String())

		assert.Equal(t, record1RewardsExpected.Add(record2RewardsExpected...).String(), chain.GetBalance(contractAddr).String())
	})

	// Withdraw rewards using the limit mode with default limit
	t.Run("Withdraw 3rd reward using default limit", func(t *testing.T) {
		msg := rewardsWbTypes.WithdrawRewardsRequest{
			RecordsLimit: archPkg.Uint64Ptr(0),
		}

		_, resData, err := msgPlugin.WithdrawContractRewards(ctx, contractAddr, msg)
		require.NoError(t, err)
		require.Len(t, resData, 1)

		var res rewardsWbTypes.WithdrawRewardsResponse
		require.NoError(t, json.Unmarshal(resData[0], &res))

		assert.EqualValues(t, 1, res.RecordsNum)
		totalRewardsReceived, err := pkg.WasmCoinsToSDK(res.TotalRewards)
		require.NoError(t, err)
		assert.EqualValues(t, record3RewardsExpected.String(), totalRewardsReceived.String())

		assert.Equal(t, recordsRewards.String(), chain.GetBalance(contractAddr).String())
	})
}
