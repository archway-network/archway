package ante_test

import (
	"errors"
	"testing"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards/ante"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestRewardsMinFeeAnteHandler(t *testing.T) {
	type testCase struct {
		name string
		// Inputs
		txFees     string // transaction fees [sdk.Coins]
		txGasLimit uint64 // transaction gas limit
		minConsFee string // min consensus fee [sdk.DecCoin]
		// Output expected
		errExpected error // concrete error expected (or nil if no error expected)
	}

	testCases := []testCase{
		{
			name:       "OK: 200stake fee > 100stake min fee",
			txFees:     "200stake",
			txGasLimit: 1000,
			minConsFee: "0.1stake",
		},
		{
			name:       "OK: 100stake fee == 100stake min fee",
			txFees:     "100stake",
			txGasLimit: 1000,
			minConsFee: "0.1stake",
		},
		{
			name:        "Fail: 99stake fee < 100stake min fee",
			txFees:      "99stake",
			txGasLimit:  1000,
			minConsFee:  "0.1stake",
			errExpected: sdkErrors.ErrInsufficientFee,
		},
		{
			name:       "OK: min consensus fee is zero",
			txFees:     "100stake",
			txGasLimit: 1000,
			minConsFee: "0stake",
		},
		{
			name:       "OK: expected fee is too low (zero)",
			txFees:     "1stake",
			txGasLimit: 1000,
			minConsFee: "0.000000000001stake",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create chain
			chain := e2eTesting.NewTestChain(t, 1)
			ctx := chain.GetContext()

			// Set min consensus fee
			minConsFee, err := sdk.ParseDecCoin(tc.minConsFee)
			require.NoError(t, err)

			chain.GetApp().RewardsKeeper.GetState().MinConsensusFee(ctx).SetFee(minConsFee)

			// Build transaction
			txFees, err := sdk.ParseCoinsNormalized(tc.txFees)
			require.NoError(t, err)

			tx := testutils.NewMockFeeTx(
				testutils.WithMockFeeTxFees(txFees),
				testutils.WithMockFeeTxGas(tc.txGasLimit),
			)

			// Call the Ante handler manually
			anteHandler := ante.NewMinFeeDecorator(chain.GetAppCodec(), chain.GetApp().RewardsKeeper)
			_, err = anteHandler.AnteHandle(ctx, tx, false, testutils.NoopAnteHandler)
			if tc.errExpected != nil {
				assert.ErrorIs(t, err, tc.errExpected)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestRewardsContractFlatFeeAnteHandler(t *testing.T) {
	// Create chain
	chain := e2eTesting.NewTestChain(t, 1)
	ctx := chain.GetContext()

	// Set min consensus fee
	minConsFee, err := sdk.ParseDecCoin("0.1stake")
	require.NoError(t, err)
	chain.GetApp().RewardsKeeper.GetState().MinConsensusFee(ctx).SetFee(minConsFee)

	contractAdminAcc := chain.GetAccount(0)
	contractViewer := testutils.NewMockContractViewer()
	chain.GetApp().RewardsKeeper.SetContractInfoViewer(contractViewer)
	contractAddrs := e2eTesting.GenContractAddresses(3)

	// Test contract address which dosent have flatfee set
	contractFlatFeeNotSet := contractAddrs[0]
	// Test contract address which has flat fee set which is different denom than minConsensusFee
	contractFlatFeeDiffDenomSet := contractAddrs[1]
	contractViewer.AddContractAdmin(contractFlatFeeDiffDenomSet.String(), contractAdminAcc.Address.String())
	var metaCurrentDiff rewardsTypes.ContractMetadata
	metaCurrentDiff.ContractAddress = contractFlatFeeDiffDenomSet.String()
	metaCurrentDiff.OwnerAddress = contractAdminAcc.Address.String()
	err = chain.GetApp().RewardsKeeper.SetContractMetadata(ctx, contractAdminAcc.Address, contractFlatFeeDiffDenomSet, metaCurrentDiff)
	require.NoError(t, err)
	flatFeeDiff := sdk.NewInt64Coin("test", 10)
	err = chain.GetApp().RewardsKeeper.SetFlatFee(ctx, contractAdminAcc.Address, rewardsTypes.FlatFee{
		ContractAddress: contractFlatFeeDiffDenomSet.String(),
		FlatFee:         flatFeeDiff,
	})
	require.NoError(t, err)

	// Test contract address which has flat fee set which is same denom as minConsensusFee
	contractFlatFeeSameDenomSet := contractAddrs[2]
	contractViewer.AddContractAdmin(contractFlatFeeSameDenomSet.String(), contractAdminAcc.Address.String())
	var metaCurrentSame rewardsTypes.ContractMetadata
	metaCurrentSame.ContractAddress = contractFlatFeeSameDenomSet.String()
	metaCurrentSame.OwnerAddress = contractAdminAcc.Address.String()
	err = chain.GetApp().RewardsKeeper.SetContractMetadata(ctx, contractAdminAcc.Address, contractFlatFeeSameDenomSet, metaCurrentSame)
	require.NoError(t, err)
	flatFeeSame := sdk.NewInt64Coin("stake", 10)
	err = chain.GetApp().RewardsKeeper.SetFlatFee(ctx, contractAdminAcc.Address, rewardsTypes.FlatFee{
		ContractAddress: contractFlatFeeSameDenomSet.String(),
		FlatFee:         flatFeeSame,
	})
	require.NoError(t, err)

	type testCase struct {
		name string
		// Inputs
		txFees    string    // transaction fees [sdk.Coins]
		txMsgs    []sdk.Msg // transaction msgs
		wrapAuthz bool      // wrap the given transaction in authz.MsgExec type
		// Output expected
		errExpected error // concrete error expected (or nil if no error expected)
	}

	testCases := []testCase{
		{
			name:   "Fail: Invalid contract address",
			txFees: "100stake",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{},
			},
			errExpected: errors.New("empty address string is not allowed"),
		},
		{
			name:   "OK: Contract flat fee not set",
			txFees: "100stake",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeNotSet.String(),
				},
			},
			errExpected: nil,
		},
		{
			name:   "Fail: Contract flat fee set + but tx doesnt send fee (diff denoms)",
			txFees: "100stake",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeDiffDenomSet.String(),
				},
			},
			errExpected: sdkErrors.ErrInsufficientFee,
		},
		{
			name:   "OK: Contract flat fee set + tx sends fee (diff denoms)",
			txFees: "100stake,10test",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeDiffDenomSet.String(),
				},
			},
			errExpected: nil,
		},
		{
			name:   "Fail: Contract flat fee set + tx sends insufficient fee (same denoms)",
			txFees: "100stake",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeSameDenomSet.String(),
				},
			},
			errExpected: sdkErrors.ErrInsufficientFee,
		},
		{
			name:   "OK: Contract flat fee set + tx sends sufficient fee (same denoms)",
			txFees: "110stake",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeSameDenomSet.String(),
				},
			},
			errExpected: nil,
		},
		{
			name:   "Fail: Contract flat fee set + tx sends insufficient fee (same&diff denoms)",
			txFees: "100stake,10test",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeSameDenomSet.String(),
				},
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeDiffDenomSet.String(),
				},
			},
			errExpected: sdkErrors.ErrInsufficientFee,
		},
		{
			name:   "OK: Contract flat fee set + tx sends sufficient fee (same&diff denoms)",
			txFees: "110stake,10test",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeSameDenomSet.String(),
				},
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeDiffDenomSet.String(),
				},
			},
			errExpected: nil,
		},
		{
			name:   "Fail: Contract flat fee set + tx doesnt send enough fee + msg is authz.MsgExec",
			txFees: "100stake",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeDiffDenomSet.String(),
				},
			},
			wrapAuthz:   true,
			errExpected: sdkErrors.ErrInsufficientFee,
		},
		{
			name:   "OK: Contract flat fee set + tx sends fee + msg is authz.MsgExec",
			txFees: "100stake,10test",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeDiffDenomSet.String(),
				},
			},
			wrapAuthz:   true,
			errExpected: nil,
		},
		{
			name:   "Fail: Contract flat fee set + tx sends insufficient fee (same&diff denoms) + msg is authz.MsgExec",
			txFees: "100stake,10test",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeSameDenomSet.String(),
				},
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeDiffDenomSet.String(),
				},
			},
			wrapAuthz:   true,
			errExpected: sdkErrors.ErrInsufficientFee,
		},
		{
			name:   "OK: Contract flat fee set + tx sends sufficient fee (same&diff denoms) + msg is authz.MsgExec",
			txFees: "110stake,10test",
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeSameDenomSet.String(),
				},
				&wasmdTypes.MsgExecuteContract{
					Contract: contractFlatFeeDiffDenomSet.String(),
				},
			},
			wrapAuthz:   true,
			errExpected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			txFees, err := sdk.ParseCoinsNormalized(tc.txFees)
			require.NoError(t, err)
			msgs := tc.txMsgs
			if tc.wrapAuthz {
				authzMsg := authz.NewMsgExec(sdk.AccAddress{}, tc.txMsgs)
				authzMsgs := []sdk.Msg{
					testutils.NewMockMsg(),
					&authzMsg,
				}
				msgs = authzMsgs
			}
			tx := testutils.NewMockFeeTx(
				testutils.WithMockFeeTxFees(txFees),
				testutils.WithMockFeeTxGas(1000),
				testutils.WithMockFeeTxMsgs(msgs...),
			)
			anteHandler := ante.NewMinFeeDecorator(chain.GetAppCodec(), chain.GetApp().RewardsKeeper)

			_, err = anteHandler.AnteHandle(ctx, tx, false, testutils.NoopAnteHandler)

			if tc.errExpected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err, tc.errExpected)
			}
		})
	}
}
