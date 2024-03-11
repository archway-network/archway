package cwerrors_test

import (
	"testing"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	cwerrorsabci "github.com/archway-network/archway/x/cwerrors"
	"github.com/archway-network/archway/x/cwerrors/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestEndBlocker(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAddr2 := contractAddresses[1]
	contractAdminAcc := chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	contractViewer.AddContractAdmin(
		contractAddr2.String(),
		contractAdminAcc.Address.String(),
	)
	params := types.Params{
		ErrorStoredTime:       5,
		DisableErrorCallbacks: true,
		SubscriptionFee:       sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)),
		SubscriptionPeriod:    5,
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	// Set errors for block 1
	contract1Err := types.SudoError{
		ContractAddress: contractAddr.String(),
		ModuleName:      "test",
	}
	contract2Err := types.SudoError{
		ContractAddress: contractAddr2.String(),
		ModuleName:      "test",
	}
	err = keeper.SetError(ctx, contract1Err)
	require.NoError(t, err)
	err = keeper.SetError(ctx, contract1Err)
	require.NoError(t, err)
	err = keeper.SetError(ctx, contract2Err)
	require.NoError(t, err)

	pruneHeight := ctx.BlockHeight() + params.ErrorStoredTime

	// Increment block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// Set errors for block 2
	err = keeper.SetError(ctx, contract1Err)
	require.NoError(t, err)
	err = keeper.SetError(ctx, contract2Err)
	require.NoError(t, err)
	err = keeper.SetError(ctx, contract2Err)
	require.NoError(t, err)

	// Check number of errors match
	sudoErrs, err := keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 3)
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 3)

	// Go to prune height
	ctx = ctx.WithBlockHeight(pruneHeight)

	_ = cwerrorsabci.EndBlocker(ctx, keeper, contractViewer)

	// Check number of errors match
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 1)
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 2)

	// Go to next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	_ = cwerrorsabci.EndBlocker(ctx, keeper, contractViewer)

	// Check number of errors match
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 0)
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 0)

	// Setup subscription
	expiryTime, err := keeper.SetSubscription(ctx, contractAddr, sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))
	require.NoError(t, err)
	require.Equal(t, ctx.BlockHeight()+params.SubscriptionPeriod, expiryTime)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	err = keeper.SetError(ctx, contract1Err)
	require.NoError(t, err)

	// Should be empty as the is stored for error callback
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 0)

	// Should be queued for callback
	sudoErrs = keeper.GetAllSudoErrorCallbacks(ctx)
	require.Len(t, sudoErrs, 1)

	// Execute endblocker
	_ = cwerrorsabci.EndBlocker(ctx, keeper, contractViewer)
}
