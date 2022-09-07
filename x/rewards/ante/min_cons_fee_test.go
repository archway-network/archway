package ante_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards/ante"
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
			anteHandler := ante.NewMinFeeDecorator(chain.GetApp().RewardsKeeper)
			_, err = anteHandler.AnteHandle(ctx, tx, false, testutils.NoopAnteHandler)
			if tc.errExpected != nil {
				assert.ErrorIs(t, err, tc.errExpected)
				return
			}
			require.NoError(t, err)
		})
	}
}
