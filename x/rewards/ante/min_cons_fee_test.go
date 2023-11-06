package ante_test

import (
	"errors"
	"testing"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
		minPoG     string
		// Output expected
		errExpected error // concrete error expected (or nil if no error expected)
	}

	testCases := []testCase{
		{
			name:       "OK: 200stake fee > 100stake min fee",
			txFees:     "200stake",
			txGasLimit: 1000,
			minConsFee: "0.1stake",
			minPoG:     "0stake",
		},
		{
			name:       "OK: 100stake fee == 100stake min fee",
			txFees:     "100stake",
			txGasLimit: 1000,
			minConsFee: "0.1stake",
			minPoG:     "0stake",
		},
		{
			name:        "Fail: 99stake fee < 100stake min fee",
			txFees:      "99stake",
			txGasLimit:  1000,
			minConsFee:  "0.1stake",
			minPoG:      "0stake",
			errExpected: sdkErrors.ErrInsufficientFee,
		},
		{
			name:       "OK: min consensus fee is zero",
			txFees:     "100stake",
			txGasLimit: 1000,
			minConsFee: "0stake",
			minPoG:     "0stake",
		},
		{
			name:       "OK: expected fee is too low (zero)",
			txFees:     "1stake",
			txGasLimit: 1000,
			minConsFee: "0.000000000001stake",
			minPoG:     "0stake",
		},
		{
			name:        "OK: min PoG used, min cons fee not set",
			txFees:      "1000stake",
			txGasLimit:  1000,
			minConsFee:  "0stake",
			minPoG:      "1stake",
			errExpected: nil,
		},
		{
			name:        "OK: min PoG used, min cons fee lower",
			txFees:      "1000stake",
			txGasLimit:  1000,
			minConsFee:  "0.1stake",
			minPoG:      "1stake",
			errExpected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create chain
			chain := e2eTesting.NewTestChain(t, 1)
			ctx := chain.GetContext()
			keepers := chain.GetApp().Keepers

			// Set min consensus fee
			minConsFee, err := sdk.ParseDecCoin(tc.minConsFee)
			require.NoError(t, err)

			keepers.RewardsKeeper.GetState().MinConsensusFee(ctx).SetFee(minConsFee)
			params := keepers.RewardsKeeper.GetParams(ctx)
			coin, err := sdk.ParseDecCoin(tc.minPoG)
			require.NoError(t, err)
			params.MinPriceOfGas = coin
			err = keepers.RewardsKeeper.SetParams(ctx, params)
			require.NoError(t, err)

			// Build transaction
			txFees, err := sdk.ParseCoinsNormalized(tc.txFees)
			require.NoError(t, err)

			tx := testutils.NewMockFeeTx(
				testutils.WithMockFeeTxFees(txFees),
				testutils.WithMockFeeTxGas(tc.txGasLimit),
			)

			// Call the Ante handler manually
			anteHandler := ante.NewMinFeeDecorator(chain.GetAppCodec(), keepers.RewardsKeeper)
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
	keepers := chain.GetApp().Keepers

	// Set min consensus fee
	minConsFee, err := sdk.ParseDecCoin("0.1stake")
	require.NoError(t, err)
	keepers.RewardsKeeper.GetState().MinConsensusFee(ctx).SetFee(minConsFee)

	contractAdminAcc := chain.GetAccount(0)
	contractViewer := testutils.NewMockContractViewer()
	keepers.RewardsKeeper.SetContractInfoViewer(contractViewer)
	contractAddrs := e2eTesting.GenContractAddresses(3)

	// Test contract address which dosent have flatfee set
	contractFlatFeeNotSet := contractAddrs[0]
	// Test contract address which has flat fee set which is different denom than minConsensusFee
	contractFlatFeeDiffDenomSet := contractAddrs[1]
	contractViewer.AddContractAdmin(contractFlatFeeDiffDenomSet.String(), contractAdminAcc.Address.String())
	var metaCurrentDiff rewardsTypes.ContractMetadata
	metaCurrentDiff.ContractAddress = contractFlatFeeDiffDenomSet.String()
	metaCurrentDiff.RewardsAddress = contractAdminAcc.Address.String()
	metaCurrentDiff.OwnerAddress = contractAdminAcc.Address.String()
	err = keepers.RewardsKeeper.SetContractMetadata(ctx, contractAdminAcc.Address, contractFlatFeeDiffDenomSet, metaCurrentDiff)
	require.NoError(t, err)
	flatFeeDiff := sdk.NewInt64Coin("test", 10)
	err = keepers.RewardsKeeper.SetFlatFee(ctx, contractAdminAcc.Address, rewardsTypes.FlatFee{
		ContractAddress: contractFlatFeeDiffDenomSet.String(),
		FlatFee:         flatFeeDiff,
	})
	require.NoError(t, err)

	// Test contract address which has flat fee set which is same denom as minConsensusFee
	contractFlatFeeSameDenomSet := contractAddrs[2]
	contractViewer.AddContractAdmin(contractFlatFeeSameDenomSet.String(), contractAdminAcc.Address.String())
	var metaCurrentSame rewardsTypes.ContractMetadata
	metaCurrentSame.ContractAddress = contractFlatFeeSameDenomSet.String()
	metaCurrentSame.RewardsAddress = contractAdminAcc.Address.String()
	metaCurrentSame.OwnerAddress = contractAdminAcc.Address.String()
	err = keepers.RewardsKeeper.SetContractMetadata(ctx, contractAdminAcc.Address, contractFlatFeeSameDenomSet, metaCurrentSame)
	require.NoError(t, err)
	flatFeeSame := sdk.NewInt64Coin("stake", 10)
	err = keepers.RewardsKeeper.SetFlatFee(ctx, contractAdminAcc.Address, rewardsTypes.FlatFee{
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
			anteHandler := ante.NewMinFeeDecorator(chain.GetAppCodec(), keepers.RewardsKeeper)

			_, err = anteHandler.AnteHandle(ctx, tx, false, testutils.NoopAnteHandler)

			if tc.errExpected == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err, tc.errExpected)
			}
		})
	}
}

func TestAuthzDecodeAntehandler(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	keepers := chain.GetApp().Keepers
	minConsFee, _ := sdk.ParseDecCoin("0.1stake") // Set min consensus fee
	keepers.RewardsKeeper.GetState().MinConsensusFee(chain.GetContext()).SetFee(minConsFee)
	txFees, _ := sdk.ParseCoinsNormalized("100stake")

	// Making a wrapped MsgDelegate
	authzMsg := authz.NewMsgExec(sdk.AccAddress{}, []sdk.Msg{&stakingTypes.MsgDelegate{
		DelegatorAddress: e2eTesting.TestAccountAddr.String(),
		ValidatorAddress: sdk.ValAddress(e2eTesting.TestAccountAddr).String(),
		Amount:           sdk.NewInt64Coin("stake", 10),
	}})

	tx := testutils.NewMockFeeTx(
		testutils.WithMockFeeTxFees(txFees),
		testutils.WithMockFeeTxGas(1000),
		testutils.WithMockFeeTxMsgs([]sdk.Msg{&authzMsg}...),
	)

	anteHandler := ante.NewMinFeeDecorator(chain.GetAppCodec(), keepers.RewardsKeeper)
	_, err := anteHandler.AnteHandle(chain.GetContext(), tx, false, testutils.NoopAnteHandler)

	require.NoError(t, err)
}
