package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gstTypes "github.com/archway-network/archway/x/gastracker"
	"github.com/stretchr/testify/require"
)

func TestSetParams(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t, sdk.AccAddress{})

	// Initialize default values
	params := gstTypes.DefaultParams()
	keeper.SetParams(ctx, params)

	// Retrieve default values
	require.Equal(t, gstTypes.DefaultGasTrackingSwitch, keeper.IsGasTrackingEnabled(ctx), "gas tracking is not default value")
	require.Equal(t, gstTypes.DefaultGasRebateSwitch, keeper.IsGasRebateToContractEnabled(ctx), "gas rebate is not default value")
	require.Equal(t, gstTypes.DefaultGasRebateToUserSwitch, keeper.IsGasRebateToUserEnabled(ctx), "gas rebate to user is not default value")
	require.Equal(t, gstTypes.DefaultContractPremiumSwitch, keeper.IsContractPremiumEnabled(ctx), "contract premium is not default value")
	require.Equal(t, gstTypes.DefaultInflationRewardCapPercentage, keeper.InflationRewardCapPercentage(ctx), "inflation reward cap percentage is not default")
	require.Equal(t, gstTypes.DefaultInflationRewardCapSwitch, keeper.IsInflationRewardCapped(ctx), "inflation reward cap switch is not default")
	require.Equal(t, gstTypes.DefaultInflationRewardQuotaPercentage, keeper.InflationRewardQuotaPercentage(ctx), "inflation reward quota percentage is not default")

	// Disable features
	params.GasTrackingSwitch = false
	params.GasRebateSwitch = false
	params.GasRebateToUserSwitch = false
	params.ContractPremiumSwitch = false
	params.InflationRewardQuotaPercentage = 40
	params.InflationRewardCapSwitch = false
	params.InflationRewardCapPercentage = 50

	keeper.SetParams(ctx, params)

	require.Equal(t, false, keeper.IsGasTrackingEnabled(ctx), "gas tracking was not updated ")
	require.Equal(t, false, keeper.IsGasRebateToContractEnabled(ctx), "gas rebate was not updated")
	require.Equal(t, false, keeper.IsGasRebateToUserEnabled(ctx), "gas rebate to user was not updated")
	require.Equal(t, false, keeper.IsContractPremiumEnabled(ctx), "contract premium was not updated")
	require.Equal(t, uint64(40), keeper.InflationRewardQuotaPercentage(ctx), "inflation reward quota percentage was not updated")
	require.Equal(t, false, keeper.IsInflationRewardCapped(ctx), "inflation reward cap switch was not updated")
	require.Equal(t, uint64(50), keeper.InflationRewardCapPercentage(ctx), "inflation reward cap percentage was not updated")
}
