package wasmbinding_test

import (
	"encoding/json"
	"testing"
	"time"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/pkg/testutils"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
	wasmBindings "github.com/archway-network/archway/x/rewards/wasmbinding"
	wasmBindingsTypes "github.com/archway-network/archway/x/rewards/wasmbinding/types"
)

// TestWASMBindings tests the custom querier and custom message handler for WASM bindings.
func TestWASMBindings(t *testing.T) {
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
	queryPlugin := wasmBindings.NewQueryPlugin(keeper)
	msgPlugin := wasmBindings.NewMsgPlugin(testutils.NewMockMessenger(), keeper)

	// Helpers
	buildQuery := func(metaReq *wasmBindingsTypes.ContractMetadataRequest, rewardsReq *wasmBindingsTypes.RewardsRecordsRequest) []byte {
		query := wasmBindingsTypes.Query{
			Metadata:       metaReq,
			RewardsRecords: rewardsReq,
		}
		bz, err := json.Marshal(query)
		require.NoError(t, err)
		return bz
	}

	buildMsg := func(updateMetaReq *wasmBindingsTypes.UpdateMetadataRequest, withdrawReq *wasmBindingsTypes.WithdrawRewardsRequest) wasmVmTypes.CosmosMsg {
		msg := wasmBindingsTypes.Msg{
			UpdateMetadata:  updateMetaReq,
			WithdrawRewards: withdrawReq,
		}
		msgBz, err := json.Marshal(msg)
		require.NoError(t, err)

		return wasmVmTypes.CosmosMsg{
			Custom: msgBz,
		}
	}

	// Invalid inputs
	t.Run("Invalid query input", func(t *testing.T) {
		queryBz := buildQuery(nil, nil)

		_, err := queryPlugin.DispatchQuery(ctx, queryBz)
		assert.ErrorIs(t, err, rewardsTypes.ErrInvalidRequest)
	})

	t.Run("Invalid msg input", func(t *testing.T) {
		msg := buildMsg(nil, nil)

		_, _, err := msgPlugin.DispatchMsg(ctx, contractAddr, "", msg)
		assert.ErrorIs(t, err, rewardsTypes.ErrInvalidRequest)
	})

	// Query empty / non-existing data
	t.Run("Query non-existing metadata", func(t *testing.T) {
		queryBz := buildQuery(
			&wasmBindingsTypes.ContractMetadataRequest{
				ContractAddress: contractAddr.String(),
			},
			nil,
		)

		_, err := queryPlugin.DispatchQuery(ctx, queryBz)
		assert.ErrorIs(t, err, rewardsTypes.ErrMetadataNotFound)
	})

	t.Run("Query empty rewards", func(t *testing.T) {
		queryBz := buildQuery(
			nil,
			&wasmBindingsTypes.RewardsRecordsRequest{
				RewardsAddress: contractAddr.String(),
			},
		)

		resBz, err := queryPlugin.DispatchQuery(ctx, queryBz)
		require.NoError(t, err)

		var res wasmBindingsTypes.RewardsRecordsResponse
		require.NoError(t, json.Unmarshal(resBz, &res))
		assert.Empty(t, res.Records)
	})

	// Handle no-op msg
	t.Run("Update non-existing metadata (unauthorized create operation)", func(t *testing.T) {
		msg := buildMsg(
			&wasmBindingsTypes.UpdateMetadataRequest{
				OwnerAddress:   acc.Address.String(),
				RewardsAddress: acc.Address.String(),
			},
			nil,
		)

		_, _, err := msgPlugin.DispatchMsg(ctx, contractAddr, "", msg)
		assert.ErrorIs(t, err, rewardsTypes.ErrUnauthorized)
	})

	t.Run("Withdraw empty rewards", func(t *testing.T) {
		msg := buildMsg(
			nil,
			&wasmBindingsTypes.WithdrawRewardsRequest{
				RecordsLimit: 1000,
			},
		)

		_, resData, err := msgPlugin.DispatchMsg(ctx, contractAddr, "", msg)
		require.NoError(t, err)

		require.Len(t, resData, 1)
		var res wasmBindingsTypes.WithdrawRewardsResponse
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
		msg := buildMsg(
			&wasmBindingsTypes.UpdateMetadataRequest{
				OwnerAddress:   contractAddr.String(),
				RewardsAddress: contractAddr.String(),
			},
			nil,
		)

		_, _, err := msgPlugin.DispatchMsg(ctx, contractAddr, "", msg)
		require.NoError(t, err)
	})

	t.Run("Check metadata updated", func(t *testing.T) {
		queryBz := buildQuery(
			&wasmBindingsTypes.ContractMetadataRequest{
				ContractAddress: contractAddr.String(),
			},
			nil,
		)

		resBz, err := queryPlugin.DispatchQuery(ctx, queryBz)
		require.NoError(t, err)

		var res wasmBindingsTypes.ContractMetadataResponse
		require.NoError(t, json.Unmarshal(resBz, &res))

		assert.Equal(t, contractAddr.String(), res.OwnerAddress)
		assert.Equal(t, contractAddr.String(), res.RewardsAddress)
	})

	// Add some rewards to withdraw (create new records and mint tokens)
	record1RewardsExpected := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 25))
	record2RewardsExpected := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 75))
	recordsRewards := record1RewardsExpected.Add(record2RewardsExpected...)

	keeper.GetState().RewardsRecord(ctx).CreateRewardsRecord(contractAddr, record1RewardsExpected, ctx.BlockHeight(), ctx.BlockTime())
	keeper.GetState().RewardsRecord(ctx).CreateRewardsRecord(contractAddr, record2RewardsExpected, ctx.BlockHeight(), ctx.BlockTime())
	require.NoError(t, chain.GetApp().MintKeeper.MintCoins(ctx, recordsRewards))
	require.NoError(t, chain.GetApp().BankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, rewardsTypes.ContractRewardCollector, recordsRewards))

	// Query available rewards
	t.Run("Query new rewards", func(t *testing.T) {
		queryBz := buildQuery(
			nil,
			&wasmBindingsTypes.RewardsRecordsRequest{
				RewardsAddress: contractAddr.String(),
			},
		)

		resBz, err := queryPlugin.DispatchQuery(ctx, queryBz)
		require.NoError(t, err)

		var res wasmBindingsTypes.RewardsRecordsResponse
		require.NoError(t, json.Unmarshal(resBz, &res))

		require.Len(t, res.Records, 2)
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
	})

	// Withdraw rewards using the limit mode
	t.Run("Withdraw 1st reward using limit", func(t *testing.T) {
		msg := buildMsg(
			nil,
			&wasmBindingsTypes.WithdrawRewardsRequest{
				RecordsLimit: 1,
			},
		)

		_, resData, err := msgPlugin.DispatchMsg(ctx, contractAddr, "", msg)
		require.NoError(t, err)

		require.Len(t, resData, 1)
		var res wasmBindingsTypes.WithdrawRewardsResponse
		require.NoError(t, json.Unmarshal(resData[0], &res))

		assert.EqualValues(t, 1, res.RecordsNum)
		totalRewardsReceived, err := pkg.WasmCoinsToSDK(res.TotalRewards)
		require.NoError(t, err)
		assert.EqualValues(t, record1RewardsExpected.String(), totalRewardsReceived.String())

		assert.Equal(t, record1RewardsExpected.String(), chain.GetBalance(contractAddr).String())
	})

	// Withdraw rewards using the record IDs mode
	t.Run("Withdraw 2nd reward using record ID", func(t *testing.T) {
		msg := buildMsg(
			nil,
			&wasmBindingsTypes.WithdrawRewardsRequest{
				RecordIDs: []uint64{2},
			},
		)

		_, resData, err := msgPlugin.DispatchMsg(ctx, contractAddr, "", msg)
		require.NoError(t, err)

		require.Len(t, resData, 1)
		var res wasmBindingsTypes.WithdrawRewardsResponse
		require.NoError(t, json.Unmarshal(resData[0], &res))

		assert.EqualValues(t, 1, res.RecordsNum)
		totalRewardsReceived, err := pkg.WasmCoinsToSDK(res.TotalRewards)
		require.NoError(t, err)
		assert.EqualValues(t, record2RewardsExpected.String(), totalRewardsReceived.String())

		assert.Equal(t, recordsRewards.String(), chain.GetBalance(contractAddr).String())
	})
}
