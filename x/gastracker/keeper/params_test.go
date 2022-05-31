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
	params := gstTypes.DefaultParams(ctx)
	keeper.SetParams(ctx, params)

	// Retrieve default values
	require.Equal(t, true, keeper.IsGasTrackingEnabled(ctx), "gas tracking is not default value")
	require.Equal(t, true, keeper.IsGasRebateToContractEnabled(ctx), "gas rebate is not default value")
	require.Equal(t, true, keeper.IsGasRebateToUserEnabled(ctx), "gas rebate to user is not default value")
	require.Equal(t, true, keeper.IsContractPremiumEnabled(ctx), "contract premium is not default value")
	require.Equal(t, (ctx.BlockGasMeter().Limit()*5)/100, keeper.GetMaxGasForContractFeeGrant(ctx), "max gas for contract fee grant is not default value")
	require.Equal(t, (ctx.BlockGasMeter().Limit()*40)/100, keeper.GetMaxGasForGlobalFeeGrant(ctx), "max gas for global fee grant is not default value")

	// Disable features
	params.GasTrackingSwitch = false
	params.GasRebateSwitch = false
	params.GasRebateToUserSwitch = false
	params.ContractPremiumSwitch = false
	params.MaxGasForGlobalFeeGrant = 2
	params.MaxGasForContractFeeGrant = 1
	keeper.SetParams(ctx, params)

	require.Equal(t, false, keeper.IsGasTrackingEnabled(ctx), "gas tracking was not updated ")
	require.Equal(t, false, keeper.IsGasRebateToContractEnabled(ctx), "gas rebate was not updated")
	require.Equal(t, false, keeper.IsGasRebateToUserEnabled(ctx), "gas rebate to user was not updated")
	require.Equal(t, false, keeper.IsContractPremiumEnabled(ctx), "contract premium was not updated")
	require.Equal(t, uint64(2), keeper.GetMaxGasForGlobalFeeGrant(ctx), "max gas for global fee grant is not updated")
	require.Equal(t, uint64(1), keeper.GetMaxGasForContractFeeGrant(ctx), "max gas for contract fee grant is not updated")
}
