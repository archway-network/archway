package wasmbinding_test

import (
	"encoding/json"
	"testing"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
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
	buildQuery := func(metaReq *wasmBindingsTypes.ContractMetadataRequest, rewardsReq *wasmBindingsTypes.CurrentRewardsRequest) []byte {
		query := wasmBindingsTypes.Query{
			Metadata:       metaReq,
			CurrentRewards: rewardsReq,
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
			&wasmBindingsTypes.CurrentRewardsRequest{
				RewardsAddress: contractAddr.String(),
			},
		)

		resBz, err := queryPlugin.DispatchQuery(ctx, queryBz)
		require.NoError(t, err)

		var res wasmBindingsTypes.CurrentRewardsResponse
		require.NoError(t, json.Unmarshal(resBz, &res))
		assert.Empty(t, res.Rewards)
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
			&wasmBindingsTypes.WithdrawRewardsRequest{},
		)

		_, resData, err := msgPlugin.DispatchMsg(ctx, contractAddr, "", msg)
		require.NoError(t, err)

		require.Len(t, resData, 1)
		var res wasmBindingsTypes.WithdrawRewardsResponse
		require.NoError(t, json.Unmarshal(resData[0], &res))
		assert.Empty(t, res.Rewards)
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

	// Add some rewards to withdraw (create a new record and mint tokens)
	rewards := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 100))
	keeper.GetState().RewardsRecord(ctx).CreateRewardsRecord(contractAddr, rewards, ctx.BlockHeight(), ctx.BlockTime())
	require.NoError(t, chain.GetApp().MintKeeper.MintCoins(ctx, rewards))
	require.NoError(t, chain.GetApp().BankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, rewardsTypes.ModuleName, rewards))

	// Query available rewards
	t.Run("Query new rewards", func(t *testing.T) {
		queryBz := buildQuery(
			nil,
			&wasmBindingsTypes.CurrentRewardsRequest{
				RewardsAddress: contractAddr.String(),
			},
		)

		resBz, err := queryPlugin.DispatchQuery(ctx, queryBz)
		require.NoError(t, err)

		var res wasmBindingsTypes.CurrentRewardsResponse
		require.NoError(t, json.Unmarshal(resBz, &res))
		assert.Equal(t, rewards.String(), res.Rewards)
	})

	// Withdraw rewards
	t.Run("Withdraw new rewards", func(t *testing.T) {
		msg := buildMsg(
			nil,
			&wasmBindingsTypes.WithdrawRewardsRequest{},
		)

		_, resData, err := msgPlugin.DispatchMsg(ctx, contractAddr, "", msg)
		require.NoError(t, err)

		require.Len(t, resData, 1)
		var res wasmBindingsTypes.WithdrawRewardsResponse
		require.NoError(t, json.Unmarshal(resData[0], &res))
		assert.Equal(t, rewards.String(), res.Rewards)

		assert.Equal(t, rewards.String(), chain.GetBalance(contractAddr).String())
	})
}
