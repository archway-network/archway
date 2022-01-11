package gastracker

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"testing"

	gstTypes "github.com/archway-network/archway/x/gastracker/types"
	"github.com/stretchr/testify/require"
)

func TestSetParams(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t, sdk.AccAddress{})

	// Initialize default values
	params := gstTypes.DefaultParams()
	keeper.SetParams(ctx, params)

	// Retrieve default values
	require.Equal(t, true, keeper.IsGasTrackingEnabled(ctx), "gas tracking is not default value")
	require.Equal(t, true, keeper.IsGasRebateEnabled(ctx), "gas rebate is not default value")
	require.Equal(t, true, keeper.IsGasRebateToUserEnabled(ctx), "gas rebate to user is not default value")
	require.Equal(t, true, keeper.IsContractPremiumEnabled(ctx), "contract premium is not default value")

	// Disable features
	params.GasTrackingSwitch = false
	params.GasRebateSwitch = false
	params.GasRebateToUserSwitch = false
	params.ContractPremiumSwitch = false
	keeper.SetParams(ctx, params)

	require.Equal(t, false, keeper.IsGasTrackingEnabled(ctx), "gas tracking was not updated ")
	require.Equal(t, false, keeper.IsGasRebateEnabled(ctx), "gas rebate was not updated")
	require.Equal(t, false, keeper.IsGasRebateToUserEnabled(ctx), "gas rebate to user was not updated")
	require.Equal(t, false, keeper.IsContractPremiumEnabled(ctx), "contract premium was not updated")
}
