package keeper

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	gstTypes "github.com/archway-network/archway/x/gastracker"
	"github.com/stretchr/testify/require"
)

func TestSetParams(t *testing.T) {
	ctx, keeper := createTestBaseKeeperAndContext(t, sdk.AccAddress{})

	// Initialize default values
	params := gstTypes.DefaultParams()
	fmt.Println(fmt.Sprintf("%+v", params))
	keeper.SetParams(ctx, params)

	// Retrieve default values
	require.Equal(t, true, keeper.IsGasTrackingEnabled(ctx), "gas tracking is not default value")
	require.Equal(t, true, keeper.IsGasRebateToContractEnabled(ctx), "gas rebate is not default value")
	require.Equal(t, true, keeper.IsGasRebateToUserEnabled(ctx), "gas rebate to user is not default value")
	require.Equal(t, true, keeper.IsContractPremiumEnabled(ctx), "contract premium is not default value")
	require.Equal(t, gstTypes.DefaultInflationRewardQuotaPercentage, keeper.InflationRewardQuotaPercentage(ctx), "inflation reward quota percentage is not default")
	require.Equal(t, gstTypes.DefaultGasRebatePercentage, keeper.GasRebatePercentage(ctx), "gas rebate percentage is not default")

	// Disable features
	params.GasTrackingSwitch = false
	params.GasRebateSwitch = false
	params.GasRebateToUserSwitch = false
	params.ContractPremiumSwitch = false
	params.InflationRewardQuotaPercentage = 40
	params.GasRebatePercentage = 40

	keeper.SetParams(ctx, params)

	require.Equal(t, false, keeper.IsGasTrackingEnabled(ctx), "gas tracking was not updated ")
	require.Equal(t, false, keeper.IsGasRebateToContractEnabled(ctx), "gas rebate was not updated")
	require.Equal(t, false, keeper.IsGasRebateToUserEnabled(ctx), "gas rebate to user was not updated")
	require.Equal(t, false, keeper.IsContractPremiumEnabled(ctx), "contract premium was not updated")

	require.Equal(t, uint64(40), keeper.InflationRewardQuotaPercentage(ctx), "inflation reward quota percentage is not default")
	require.Equal(t, uint64(40), keeper.GasRebatePercentage(ctx), "gas rebate percentage is not default")
}
