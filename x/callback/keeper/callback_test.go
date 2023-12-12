package keeper_test

import (
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/callback/types"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestSaveCallback() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CallbackKeeper
	rewardsKeeper := s.chain.GetApp().Keepers.RewardsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	notContractAdminAcc := s.chain.GetAccount(1)
	contractOwnerAcc := s.chain.GetAccount(2)

	s.Run("FAIL: contract address is invalid", func() {
		err := keeper.SaveCallback(ctx, types.Callback{
			ContractAddress: "👻",
			JobId:           1,
			CallbackHeight:  101,
			ReservedBy:      contractAddr.String(),
			FeeSplit: &types.CallbackFeesFeeSplit{
				TransactionFees:       &validCoin,
				BlockReservationFees:  &validCoin,
				FutureReservationFees: &validCoin,
				SurplusFees:           &validCoin,
			},
		})
		s.Assert().ErrorContains(err, "decoding bech32 failed: invalid bech32 string length 4")
	})

	s.Run("FAIL: contract does not exist", func() {
		err := keeper.SaveCallback(ctx, types.Callback{
			ContractAddress: e2eTesting.GenContractAddresses(1)[0].String(),
			JobId:           1,
			CallbackHeight:  101,
			ReservedBy:      contractAddr.String(),
			FeeSplit: &types.CallbackFeesFeeSplit{
				TransactionFees:       &validCoin,
				BlockReservationFees:  &validCoin,
				FutureReservationFees: &validCoin,
				SurplusFees:           &validCoin,
			},
		})
		s.Assert().ErrorIs(err, types.ErrContractNotFound)
	})

	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)

	var metaCurrent rewardsTypes.ContractMetadata
	metaCurrent.ContractAddress = contractAddr.String()
	metaCurrent.OwnerAddress = contractOwnerAcc.Address.String()
	metaCurrent.RewardsAddress = contractOwnerAcc.Address.String()
	rewardsKeeper.SetContractInfoViewer(contractViewer)
	err := rewardsKeeper.SetContractMetadata(ctx, contractAdminAcc.Address, contractAddr, metaCurrent)
	s.Require().NoError(err)

	s.Run("FAIL: sender not authorized to modify", func() {
		err := keeper.SaveCallback(ctx, types.Callback{
			ContractAddress: contractAddr.String(),
			JobId:           1,
			CallbackHeight:  101,
			ReservedBy:      notContractAdminAcc.Address.String(),
			FeeSplit: &types.CallbackFeesFeeSplit{
				TransactionFees:       &validCoin,
				BlockReservationFees:  &validCoin,
				FutureReservationFees: &validCoin,
				SurplusFees:           &validCoin,
			},
		})
		s.Assert().ErrorIs(err, types.ErrUnauthorized)
	})

	s.Run("FAIL: callback height is in the past", func() {
		err := keeper.SaveCallback(ctx, types.Callback{
			ContractAddress: contractAddr.String(),
			JobId:           1,
			CallbackHeight:  99,
			ReservedBy:      contractAddr.String(),
			FeeSplit: &types.CallbackFeesFeeSplit{
				TransactionFees:       &validCoin,
				BlockReservationFees:  &validCoin,
				FutureReservationFees: &validCoin,
				SurplusFees:           &validCoin,
			},
		})
		s.Assert().ErrorIs(err, types.ErrCallbackHeightNotinFuture)
	})

	s.Run("OK: save callback - sender is contract", func() {
		callback := types.Callback{
			ContractAddress: contractAddr.String(),
			JobId:           1,
			CallbackHeight:  101,
			ReservedBy:      contractAddr.String(),
			FeeSplit: &types.CallbackFeesFeeSplit{
				TransactionFees:       &validCoin,
				BlockReservationFees:  &validCoin,
				FutureReservationFees: &validCoin,
				SurplusFees:           &validCoin,
			},
		}
		err := keeper.SaveCallback(ctx, callback)
		s.Assert().NoError(err)
		callbackFound, err := keeper.GetCallback(ctx, 101, contractAddr.String(), 1)
		s.Assert().NoError(err)
		s.Assert().Equal(callback, callbackFound)
	})

	s.Run("OK: save callback - sender is contract metadata owner", func() {
		callback := types.Callback{
			ContractAddress: contractAddr.String(),
			JobId:           2,
			CallbackHeight:  101,
			ReservedBy:      contractOwnerAcc.Address.String(),
			FeeSplit: &types.CallbackFeesFeeSplit{
				TransactionFees:       &validCoin,
				BlockReservationFees:  &validCoin,
				FutureReservationFees: &validCoin,
				SurplusFees:           &validCoin,
			},
		}
		err := keeper.SaveCallback(ctx, callback)
		s.Assert().NoError(err)
		callbackFound, err := keeper.GetCallback(ctx, 101, contractAddr.String(), 2)
		s.Assert().NoError(err)
		s.Assert().Equal(callback, callbackFound)
	})

	s.Run("OK: save callback - sender is contract admin", func() {
		callback := types.Callback{
			ContractAddress: contractAddr.String(),
			JobId:           3,
			CallbackHeight:  101,
			ReservedBy:      contractAdminAcc.Address.String(),
			FeeSplit: &types.CallbackFeesFeeSplit{
				TransactionFees:       &validCoin,
				BlockReservationFees:  &validCoin,
				FutureReservationFees: &validCoin,
				SurplusFees:           &validCoin,
			},
		}
		err := keeper.SaveCallback(ctx, callback)
		s.Assert().NoError(err)
		callbackFound, err := keeper.GetCallback(ctx, 101, contractAddr.String(), 3)
		s.Assert().NoError(err)
		s.Assert().Equal(callback, callbackFound)
	})

	s.Run("FAIL: callback already exists", func() {
		err := keeper.SaveCallback(ctx, types.Callback{
			ContractAddress: contractAddr.String(),
			JobId:           1,
			CallbackHeight:  101,
			ReservedBy:      contractAddr.String(),
			FeeSplit: &types.CallbackFeesFeeSplit{
				TransactionFees:       &validCoin,
				BlockReservationFees:  &validCoin,
				FutureReservationFees: &validCoin,
				SurplusFees:           &validCoin,
			},
		})
		s.Assert().Error(err)
	})
}

// func (s *KeeperTestSuite) TestDeleteCallback() {
// 	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.RewardsKeeper
// 	contractAdminAcc := s.chain.GetAccount(0)
// 	contractViewer := testutils.NewMockContractViewer()
// 	keeper.SetContractInfoViewer(contractViewer)

// 	contractAddr := e2eTesting.GenContractAddresses(1)[0]
// 	fee := sdk.NewInt64Coin("test", 10)

// 	s.Run()
// }
