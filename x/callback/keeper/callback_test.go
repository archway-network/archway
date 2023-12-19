package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/callback/types"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
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

	s.Run("FAIL: callback height is current height", func() {
		err := keeper.SaveCallback(ctx, types.Callback{
			ContractAddress: contractAddr.String(),
			JobId:           1,
			CallbackHeight:  ctx.BlockHeight(),
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
		s.Assert().ErrorIs(err, types.ErrCallbackExists)
	})
}

func (s *KeeperTestSuite) TestDeleteCallback() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CallbackKeeper
	rewardsKeeper := s.chain.GetApp().Keepers.RewardsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	notContractAdminAcc := s.chain.GetAccount(1)
	contractOwnerAcc := s.chain.GetAccount(2)

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
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 2
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 3
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	s.Run("FAIL: Invalid contract address", func() {
		err := keeper.DeleteCallback(ctx, contractAddr.String(), 101, "👻", 0)
		s.Assert().ErrorContains(err, "decoding bech32 failed: invalid bech32 string length 4")
	})

	s.Run("FAIL: Not authorized to delete callback", func() {
		err := keeper.DeleteCallback(ctx, notContractAdminAcc.Address.String(), 101, contractAddr.String(), 0)
		s.Assert().ErrorIs(err, types.ErrUnauthorized)
	})

	s.Run("FAIL: Callback does not exist", func() {
		err := keeper.DeleteCallback(ctx, contractAddr.String(), 101, contractAddr.String(), 0)
		s.Assert().ErrorIs(err, types.ErrCallbackNotFound)
	})

	s.Run("OK: Success delete - sender is contract", func() {
		exists, err := keeper.ExistsCallback(ctx, 101, contractAddr.String(), 1)
		s.Assert().NoError(err)
		s.Assert().True(exists)

		err = keeper.DeleteCallback(ctx, contractAddr.String(), 101, contractAddr.String(), 1)
		s.Assert().NoError(err)

		exists, err = keeper.ExistsCallback(ctx, 101, contractAddr.String(), 1)
		s.Assert().NoError(err)
		s.Assert().False(exists)
	})

	s.Run("OK: Success delete - sender is contract admin", func() {
		exists, err := keeper.ExistsCallback(ctx, 101, contractAddr.String(), 2)
		s.Assert().NoError(err)
		s.Assert().True(exists)

		err = keeper.DeleteCallback(ctx, contractAdminAcc.Address.String(), 101, contractAddr.String(), 2)
		s.Assert().NoError(err)

		exists, err = keeper.ExistsCallback(ctx, 101, contractAddr.String(), 2)
		s.Assert().NoError(err)
		s.Assert().False(exists)
	})

	s.Run("OK: Success delete - sender is contract owner", func() {
		exists, err := keeper.ExistsCallback(ctx, 101, contractAddr.String(), 3)
		s.Assert().NoError(err)
		s.Assert().True(exists)

		err = keeper.DeleteCallback(ctx, contractOwnerAcc.Address.String(), 101, contractAddr.String(), 3)
		s.Assert().NoError(err)

		exists, err = keeper.ExistsCallback(ctx, 101, contractAddr.String(), 3)
		s.Assert().NoError(err)
		s.Assert().False(exists)
	})
}

func (s *KeeperTestSuite) TestGetCallbacksByHeight() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CallbackKeeper
	rewardsKeeper := s.chain.GetApp().Keepers.RewardsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	contractOwnerAcc := s.chain.GetAccount(2)

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

	callbackHeight := int64(101)

	callback := types.Callback{
		ContractAddress: contractAddr.String(),
		JobId:           1,
		CallbackHeight:  callbackHeight,
		ReservedBy:      contractAddr.String(),
		FeeSplit: &types.CallbackFeesFeeSplit{
			TransactionFees:       &validCoin,
			BlockReservationFees:  &validCoin,
			FutureReservationFees: &validCoin,
			SurplusFees:           &validCoin,
		},
	}
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 2
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 3
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	s.Run("OK: Get all three existing callbacks at height 101", func() {
		callbacks, err := keeper.GetCallbacksByHeight(ctx, callbackHeight)
		s.Assert().NoError(err)
		s.Assert().Equal(3, len(callbacks))
	})
	s.Run("OK: Get zero existing callbacks at height 102", func() {
		callbacks, err := keeper.GetCallbacksByHeight(ctx, callbackHeight+1)
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(callbacks))
	})
}

func (s *KeeperTestSuite) TestGetAllCallbacks() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CallbackKeeper
	rewardsKeeper := s.chain.GetApp().Keepers.RewardsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	contractOwnerAcc := s.chain.GetAccount(2)

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

	callbackHeight := int64(105)

	s.Run("OK: Get zero existing callbacks", func() {
		callbacks, err := keeper.GetAllCallbacks(ctx)
		s.Assert().NoError(err)
		s.Assert().Equal(0, len(callbacks))
	})

	callback := types.Callback{
		ContractAddress: contractAddr.String(),
		JobId:           1,
		CallbackHeight:  callbackHeight,
		ReservedBy:      contractAddr.String(),
		FeeSplit: &types.CallbackFeesFeeSplit{
			TransactionFees:       &validCoin,
			BlockReservationFees:  &validCoin,
			FutureReservationFees: &validCoin,
			SurplusFees:           &validCoin,
		},
	}
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 2
	callback.CallbackHeight = callbackHeight + 1
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 3
	callback.CallbackHeight = callbackHeight + 2
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	s.Run("OK: Get all existing callbacks - 3", func() {
		callbacks, err := keeper.GetAllCallbacks(ctx)
		s.Assert().NoError(err)
		s.Assert().Equal(3, len(callbacks))
	})
}

func (s *KeeperTestSuite) TestIterateCallbacksByHeight() {
	ctx, keeper := s.chain.GetContext().WithBlockHeight(100), s.chain.GetApp().Keepers.CallbackKeeper
	rewardsKeeper := s.chain.GetApp().Keepers.RewardsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	validCoin := sdk.NewInt64Coin("stake", 10)

	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	contractAdminAcc := s.chain.GetAccount(0)
	contractOwnerAcc := s.chain.GetAccount(2)

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

	callbackHeight := int64(101)

	callback := types.Callback{
		ContractAddress: contractAddr.String(),
		JobId:           1,
		CallbackHeight:  callbackHeight,
		ReservedBy:      contractAddr.String(),
		FeeSplit: &types.CallbackFeesFeeSplit{
			TransactionFees:       &validCoin,
			BlockReservationFees:  &validCoin,
			FutureReservationFees: &validCoin,
			SurplusFees:           &validCoin,
		},
	}
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 2
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 3
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	callback.JobId = 4
	callback.CallbackHeight = callbackHeight + 1
	err = keeper.SaveCallback(ctx, callback)
	s.Require().NoError(err)

	s.Run("OK: Get all three existing callbacks at height 101", func() {
		count := 0
		keeper.IterateCallbacksByHeight(ctx, callbackHeight, func(callback types.Callback) bool {
			count++
			return false
		})
		s.Assert().Equal(3, count)
	})

	s.Run("OK: Get one existing callbacks at height 102", func() {
		count := 0
		keeper.IterateCallbacksByHeight(ctx, callbackHeight+1, func(callback types.Callback) bool {
			count++
			return false
		})
		s.Assert().Equal(1, count)
	})
}
