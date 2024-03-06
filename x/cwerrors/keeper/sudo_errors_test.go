package keeper_test

import (
	"fmt"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwerrors/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (s *KeeperTestSuite) TestSetError() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAddr2 := contractAddresses[1]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	testCases := []struct {
		testCase    string
		sudoError   types.SudoError
		expectError bool
		errorType   error
	}{
		{
			testCase: "FAIL: contract address is invalid",
			sudoError: types.SudoError{
				ContractAddress: "👻",
				ModuleName:      "test",
				ErrorCode:       1,
				InputPayload:    "test",
				ErrorMessage:    "test",
			},
			expectError: true,
			errorType:   fmt.Errorf("decoding bech32 failed: invalid bech32 string length 4"),
		},
		{
			testCase: "FAIL: module name is invalid",
			sudoError: types.SudoError{
				ContractAddress: contractAddr.String(),
				ModuleName:      "",
				ErrorCode:       1,
				InputPayload:    "test",
				ErrorMessage:    "test",
			},
			expectError: true,
			errorType:   types.ErrModuleNameMissing,
		},
		{
			testCase: "FAIL: contract does not exist",
			sudoError: types.SudoError{
				ContractAddress: contractAddr2.String(),
				ModuleName:      "test",
				ErrorCode:       1,
				InputPayload:    "test",
				ErrorMessage:    "test",
			},
			expectError: true,
			errorType:   types.ErrContractNotFound,
		},
		{
			testCase: "OK: successfully set error",
			sudoError: types.SudoError{
				ContractAddress: contractAddr.String(),
				ModuleName:      "test",
				ErrorCode:       1,
				InputPayload:    "test",
				ErrorMessage:    "test",
			},
			expectError: false,
		},
	}
	for _, tc := range testCases {
		s.Run(fmt.Sprintf("Case: %s", tc.testCase), func() {
			err := keeper.SetError(ctx, tc.sudoError)
			if tc.expectError {
				s.Require().Error(err)
				s.Assert().ErrorContains(err, tc.errorType.Error())
			} else {
				s.Require().NoError(err)

				getErrors, err := keeper.GetErrorsByContractAddress(ctx, sdk.MustAccAddressFromBech32(tc.sudoError.ContractAddress))
				s.Require().NoError(err)
				s.Require().Len(getErrors, 1)
				s.Require().Equal(tc.sudoError, getErrors[0])
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetErrorsByContractAddress() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAddr2 := contractAddresses[1]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	contractViewer.AddContractAdmin(
		contractAddr2.String(),
		contractAdminAcc.Address.String(),
	)

	// Set errors for block 1
	// 2 errors for contract1
	// 1 error for contract2
	contract1Err := types.SudoError{
		ContractAddress: contractAddr.String(),
		ModuleName:      "test",
	}
	contract2Err := types.SudoError{
		ContractAddress: contractAddr2.String(),
		ModuleName:      "test",
	}
	err := keeper.SetError(ctx, contract1Err)
	s.Require().NoError(err)
	err = keeper.SetError(ctx, contract1Err)
	s.Require().NoError(err)
	err = keeper.SetError(ctx, contract2Err)
	s.Require().NoError(err)

	// Check number of errors match
	sudoErrs, err := keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	s.Require().NoError(err)
	s.Require().Len(sudoErrs, 2)
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	s.Require().NoError(err)
	s.Require().Len(sudoErrs, 1)

	// Increment block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

	// Set errors for block 2
	// 1 error for contract1
	// 1 error for contract2
	err = keeper.SetError(ctx, contract1Err)
	s.Require().NoError(err)
	err = keeper.SetError(ctx, contract2Err)
	s.Require().NoError(err)

	// Check number of errors match
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	s.Require().NoError(err)
	s.Require().Len(sudoErrs, 3)
	sudoErrs, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	s.Require().NoError(err)
	s.Require().Len(sudoErrs, 2)
}

func (s *KeeperTestSuite) TestPruneErrorsByBlockHeight() {
	// Setting up chain and contract in mock wasm keeper
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	contractViewer := testutils.NewMockContractViewer()
	keeper.SetWasmKeeper(contractViewer)
	contractAddresses := e2eTesting.GenContractAddresses(3)
	contractAddr := contractAddresses[0]
	contractAddr2 := contractAddresses[1]
	contractAdminAcc := s.chain.GetAccount(0)
	contractViewer.AddContractAdmin(
		contractAddr.String(),
		contractAdminAcc.Address.String(),
	)
	contractViewer.AddContractAdmin(
		contractAddr2.String(),
		contractAdminAcc.Address.String(),
	)

	// Set errors for block 1
	contract1Err := types.SudoError{
		ContractAddress: contractAddr.String(),
		ModuleName:      "test",
	}
	contract2Err := types.SudoError{
		ContractAddress: contractAddr2.String(),
		ModuleName:      "test",
	}

	// Set errors for block 1
	// 1 errors for contract1
	// 1 error for contract2
	err := keeper.SetError(ctx, contract1Err)
	s.Require().NoError(err)
	err = keeper.SetError(ctx, contract2Err)
	s.Require().NoError(err)

	// Calculate height at which these errors are pruned
	params, err := keeper.GetParams(ctx)
	s.Require().NoError(err)
	pruneHeight := ctx.BlockHeight() + params.GetErrorStoredTime()

	// Increment block height
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	// Set errors for block 2
	// 1 error for contract1
	err = keeper.SetError(ctx, contract1Err)
	s.Require().NoError(err)

	// Check number of errors match
	getErrors, err := keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	s.Require().NoError(err)
	s.Require().Len(getErrors, 2)
	getErrors, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	s.Require().NoError(err)
	s.Require().Len(getErrors, 1)

	// Go to prune height and prune errors
	ctx = ctx.WithBlockHeight(pruneHeight)
	err = keeper.PruneErrorsCurrentBlock(ctx)
	s.Require().NoError(err)

	// Check number of errors match
	getErrors, err = keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	s.Require().NoError(err)
	s.Require().Len(getErrors, 1)
	getErrors, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	s.Require().NoError(err)
	s.Require().Len(getErrors, 0)

	// Increment block height + add error for contract 2 + prune
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	err = keeper.SetError(ctx, contract2Err)
	s.Require().NoError(err)
	err = keeper.PruneErrorsCurrentBlock(ctx)
	s.Require().NoError(err)

	// Check number of errors match
	getErrors, err = keeper.GetErrorsByContractAddress(ctx, contractAddr.Bytes())
	s.Require().NoError(err)
	s.Require().Len(getErrors, 0)
	getErrors, err = keeper.GetErrorsByContractAddress(ctx, contractAddr2.Bytes())
	s.Require().NoError(err)
	s.Require().Len(getErrors, 1)

}

func (s *KeeperTestSuite) TestGetErrorCount() {
	ctx, keeper := s.chain.GetContext(), s.chain.GetApp().Keepers.CWErrorsKeeper
	count, err := keeper.GetErrorCount(ctx)
	s.Require().NoError(err)
	s.Require().Equal(int64(0), count)

	err = keeper.ErrorsCount.Set(ctx, 6)
	s.Require().NoError(err)

	count, err = keeper.GetErrorCount(ctx)
	s.Require().NoError(err)
	s.Require().Equal(int64(6), count)
}
