package integration

import (
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/gastracker"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRewardsCollection(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	// TODO: this test can be done better but for the sake of simplicity lets keep it like this for now
	const blocks int64 = 2
	var inflation = sdk.NewInt64Coin("stake", 103)

	params, err := gastracker.NewQueryClient(chain.Client()).Params(chain.GetContext().Context(), &gastracker.QueryParamsRequest{})
	require.NoError(t, err)

	totalInflation := sdk.NewCoin(
		inflation.Denom, (inflation.Amount.ToDec().Mul(params.Params.DappInflationRewardsRatio)).MulInt64(blocks).TruncateInt().SubRaw(1)) // we're subbing a meaningless residual due to loss of precision

	gasTrackerBalance := chain.GetBalance(authtypes.NewModuleAddress(gastracker.ModuleName))
	require.Equal(t,
		gasTrackerBalance.String(),
		sdk.NewCoins(totalInflation).String(),
	)
}
