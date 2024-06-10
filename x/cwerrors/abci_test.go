package cwerrors_test

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwerrors"
	"github.com/archway-network/archway/x/cwerrors/types"
)

func TestEndBlocker(t *testing.T) {
	keeper, ctx := testutils.CWErrorsKeeper(t)
	wasmKeeper := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(wasmKeeper)

	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAddr2 := contractAddresses[1]
	contractAdminAcc := testutils.AccAddress()
	wasmKeeper.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.String(),
	)
	wasmKeeper.AddContractAdmin(
		contractAddr2.String(),
		contractAdminAcc.String(),
	)
	params := types.Params{
		ErrorStoredTime:    5,
		SubscriptionFee:    sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(0)),
		SubscriptionPeriod: 5,
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1) //chain.NextBlock(1)

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
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1) //chain.NextBlock(1)

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

	// Go to prune height & execute endblockers
	ctx = ctx.WithBlockHeight(pruneHeight)
	cwerrors.EndBlocker(ctx, keeper, wasmKeeper)

	// Check number of errors match
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 1)
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 2)

	// Go to next block & execute endblockers
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	cwerrors.EndBlocker(ctx, keeper, wasmKeeper)

	// Check number of errors match
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 0)
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 0)

	// Setup subscription
	expiryTime, err := keeper.SetSubscription(ctx, contractAdminAcc, contractAddr, sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))
	require.NoError(t, err)
	require.Equal(t, ctx.BlockHeight()+params.SubscriptionPeriod, expiryTime)

	// Go to next block
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// Set an error which should be called as callback
	err = keeper.SetError(ctx, contract1Err)
	require.NoError(t, err)
	// Set an error for a contract which has no subscription
	err = keeper.SetError(ctx, contract2Err)
	require.NoError(t, err)

	// Should be empty as the is stored for error callback
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 0)
	// Second error should still be stored in state
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 1)

	// Should be queued for callback
	sudoErrs = keeper.GetAllSudoErrorCallbacks(ctx)
	require.Len(t, sudoErrs, 1)

	// Ensure old errors in state persist and are not purged
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 1)
}
