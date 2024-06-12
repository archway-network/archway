package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestSetContractMetadata(t *testing.T) {
	k, ctx, _, wk := testutils.RewardsKeeper(t)
	contractAdminAcc, otherAcc := testutils.AccAddress(), testutils.AccAddress()
	rewardAddr := sdk.AccAddress{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	t.Run("Fail: non-existing contract", func(t *testing.T) {
		err := k.SetContractMetadata(ctx, otherAcc, contractAddr, rewardsTypes.ContractMetadata{})
		require.ErrorIs(t, err, rewardsTypes.ErrContractNotFound)
	})

	// Set contract admin
	wk.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())

	var metaCurrent rewardsTypes.ContractMetadata
	t.Run("OK: create", func(t *testing.T) {
		metaCurrent.ContractAddress = contractAddr.String()
		metaCurrent.OwnerAddress = contractAdminAcc.String()
		metaCurrent.RewardsAddress = rewardAddr.String()

		err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, metaCurrent)
		require.NoError(t, err)

		metaReceived := k.GetContractMetadata(ctx, contractAddr)
		require.NotNil(t, metaReceived)
		require.Equal(t, metaCurrent, *metaReceived)
	})

	t.Run("Fail: not a contract admin", func(t *testing.T) {
		metaCurrent := metaCurrent
		metaCurrent.OwnerAddress = otherAcc.String()
		err := k.SetContractMetadata(ctx, otherAcc, contractAddr, metaCurrent)
		require.ErrorIs(t, err, rewardsTypes.ErrUnauthorized)
	})

	t.Run("OK: set RewardsAddr", func(t *testing.T) {
		metaCurrent.RewardsAddress = otherAcc.String()

		err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, metaCurrent)
		require.NoError(t, err)

		metaReceived := k.GetContractMetadata(ctx, contractAddr)
		require.NotNil(t, metaReceived)
		require.Equal(t, metaCurrent, *metaReceived)
	})

	t.Run("OK: update OwnerAddr (change ownership)", func(t *testing.T) {
		metaCurrent.OwnerAddress = otherAcc.String()

		err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, metaCurrent)
		require.NoError(t, err)
		metaReceived := k.GetContractMetadata(ctx, contractAddr)
		require.NotNil(t, metaReceived)
		require.Equal(t, metaCurrent, *metaReceived)
	})

	t.Run("Fail: try to regain ownership", func(t *testing.T) {
		metaCurrent.OwnerAddress = contractAdminAcc.String()

		err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, metaCurrent)
		require.ErrorIs(t, err, rewardsTypes.ErrUnauthorized)
	})

	t.Run("Fail: unable to set reward address to a module account", func(t *testing.T) {
		metaCurrent.RewardsAddress = authtypes.NewModuleAddress("distribution").String()
		err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, metaCurrent)
		require.ErrorIs(t, err, rewardsTypes.ErrInvalidRequest)
	})
}
