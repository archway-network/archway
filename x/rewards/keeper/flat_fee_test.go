package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestSetFlatFee(t *testing.T) {
	k, ctx, _, wk := testutils.RewardsKeeper(t)
	contractAdminAcc := testutils.AccAddress()

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	fee := sdk.NewInt64Coin("test", 10)

	t.Run("Fail: non-existing contract metadata", func(t *testing.T) {
		err := k.SetFlatFee(ctx, contractAdminAcc, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         fee,
		})
		require.ErrorIs(t, err, rewardsTypes.ErrMetadataNotFound)
	})

	wk.AddContractAdmin(contractAddr.String(), contractAdminAcc.String())
	var metaCurrent rewardsTypes.ContractMetadata
	metaCurrent.ContractAddress = contractAddr.String()
	metaCurrent.OwnerAddress = contractAdminAcc.String()
	err := k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, metaCurrent)
	require.NoError(t, err)

	t.Run("Fail: rewards address not set", func(t *testing.T) {
		err := k.SetFlatFee(ctx, contractAdminAcc, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         fee,
		})
		require.ErrorIs(t, err, rewardsTypes.ErrMetadataNotFound)
	})

	metaCurrent.RewardsAddress = contractAdminAcc.String()
	err = k.SetContractMetadata(ctx, contractAdminAcc, contractAddr, metaCurrent)
	require.NoError(t, err)

	t.Run("OK: set flat fee", func(t *testing.T) {
		err := k.SetFlatFee(ctx, contractAdminAcc, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         fee,
		})
		require.NoError(t, err)

		flatFee, ok := k.GetFlatFee(ctx, contractAddr)
		require.True(t, ok)
		require.Equal(t, fee, flatFee)
	})

	t.Run("OK: remove flat fee", func(t *testing.T) {
		err := k.SetFlatFee(ctx, contractAdminAcc, rewardsTypes.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         sdk.NewInt64Coin("test", 0),
		})
		require.NoError(t, err)

		flatFee, ok := k.GetFlatFee(ctx, contractAddr)
		require.False(t, ok)
		require.Equal(t, sdk.Coin{}, flatFee)
	})
}
