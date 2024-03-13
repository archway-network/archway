package cwerrors_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwerrors/types"
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
		ErrorStoredTime:    5,
		SubscriptionFee:    sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(0)),
		SubscriptionPeriod: 5,
	}
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	chain.NextBlock(1)

	// Set errors for block 1
	contract1Err := types.SudoError{
		ContractAddress: contractAddr.String(),
		ModuleName:      "test",
	}
	contract2Err := types.SudoError{
		ContractAddress: contractAddr2.String(),
		ModuleName:      "test",
	}
	err = keeper.SetError(chain.GetContext(), contract1Err)
	require.NoError(t, err)
	err = keeper.SetError(chain.GetContext(), contract1Err)
	require.NoError(t, err)
	err = keeper.SetError(chain.GetContext(), contract2Err)
	require.NoError(t, err)

	pruneHeight := chain.GetContext().BlockHeight() + params.ErrorStoredTime

	// Increment block height
	chain.NextBlock(1)

	// Set errors for block 2
	err = keeper.SetError(chain.GetContext(), contract1Err)
	require.NoError(t, err)
	err = keeper.SetError(chain.GetContext(), contract2Err)
	require.NoError(t, err)
	err = keeper.SetError(chain.GetContext(), contract2Err)
	require.NoError(t, err)

	// Check number of errors match
	sudoErrs, err := keeper.GetErrorsByContractAddress(chain.GetContext(), contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 3)
	sudoErrs, err = keeper.GetErrorsByContractAddress(chain.GetContext(), contractAddr2.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 3)

	// Go to prune height & execute being&endblockers
	chain.GoToHeight(pruneHeight, time.Duration(pruneHeight))

	// Check number of errors match
	sudoErrs, err = keeper.GetErrorsByContractAddress(chain.GetContext(), contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 1)
	sudoErrs, err = keeper.GetErrorsByContractAddress(chain.GetContext(), contractAddr2.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 2)

	// Go to next block & execute being&endblockers
	chain.NextBlock(1)

	// Check number of errors match
	sudoErrs, err = keeper.GetErrorsByContractAddress(chain.GetContext(), contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 0)
	sudoErrs, err = keeper.GetErrorsByContractAddress(chain.GetContext(), contractAddr2.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 0)

	// Setup subscription
	expiryTime, err := keeper.SetSubscription(chain.GetContext(), contractAdminAcc.Address, contractAddr, sdk.NewInt64Coin(sdk.DefaultBondDenom, 0))
	require.NoError(t, err)
	require.Equal(t, chain.GetContext().BlockHeight()+params.SubscriptionPeriod, expiryTime)

	// Go to next block
	chain.NextBlock(1)

	// Set an error which should be called as callback
	err = keeper.SetError(chain.GetContext(), contract1Err)
	require.NoError(t, err)

	// Should be empty as the is stored for error callback
	sudoErrs, err = keeper.GetErrorsByContractAddress(chain.GetContext(), contractAddr.Bytes())
	require.NoError(t, err)
	require.Len(t, sudoErrs, 0)

	// Should be queued for callback
	sudoErrs = keeper.GetAllSudoErrorCallbacks(chain.GetContext())
	require.Len(t, sudoErrs, 1)

	// Execute endblocker & execute being&endblockers
	chain.NextBlock(1)

	// Check number of errors match
	sudoErrs = keeper.GetAllSudoErrorCallbacks(chain.GetContext())
	require.Len(t, sudoErrs, 0)
}
