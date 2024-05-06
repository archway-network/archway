package ante_test

import (
	"testing"

	math "cosmossdk.io/math"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/rewards/ante"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestRewardsFeeDeductionAnteHandler(t *testing.T) {
	type testCase struct {
		name string
		// Inputs
		feeRebateRatio string    // fee rebate rewards ratio (could be 0 to skip the deduction) [sdk.Dec]
		feeCoins       sdk.Coins // transaction fees (might be invalid)
		txMsgs         []sdk.Msg // transaction messages
		// Output expected
		errExpected                     bool
		rewardRecordExpected            bool   // reward record expected to be created
		feeCollectorBalanceDiffExpected string // expected FeeCollector module balance diff [sdk.Coins]
		rewardsBalanceDiffExpected      string // expected x/rewards module balance diff [sdk.Coins]
	}

	mockWasmExecuteMsg := &wasmdTypes.MsgExecuteContract{
		Contract: e2eTesting.GenContractAddresses(1)[0].String(),
	}

	newStakeCoin := func(amt uint64) sdk.Coin {
		return sdk.NewCoin(sdk.DefaultBondDenom, math.NewIntFromUint64(amt))
	}
	newArchCoin := func(amt uint64) sdk.Coin {
		return sdk.NewCoin("uarch", math.NewIntFromUint64(amt))
	}
	newInvalidCoin := func() sdk.Coin {
		return sdk.Coin{Denom: "", Amount: math.OneInt()}
	}

	testCases := []testCase{
		{
			name:           "OK: 1000stake fees with 0.5 ratio",
			feeRebateRatio: "0.5",
			feeCoins:       sdk.Coins{newStakeCoin(1000)},
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				mockWasmExecuteMsg,
			},
			rewardRecordExpected:            true,
			feeCollectorBalanceDiffExpected: "0stake", // fees are burnt
			rewardsBalanceDiffExpected:      "500stake",
		},
		{
			name:           "OK: 1000stake,500uarch fees with 0.1 ratio",
			feeRebateRatio: "0.1",
			feeCoins:       sdk.Coins{newStakeCoin(1000), newArchCoin(500)},
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				mockWasmExecuteMsg,
			},
			rewardRecordExpected:            true,
			feeCollectorBalanceDiffExpected: "0stake,0uarch", // fees are burnt
			rewardsBalanceDiffExpected:      "100stake,50uarch",
		},
		{
			name:           "OK: 1000stake fees with 0.5 ratio (no WASM msgs, rewards are skipped)",
			feeRebateRatio: "0.5",
			feeCoins:       sdk.Coins{newStakeCoin(1000)},
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
			},
			rewardRecordExpected:            false,
			feeCollectorBalanceDiffExpected: "1000stake",
			rewardsBalanceDiffExpected:      "",
		},
		{
			name:           "OK: 1000stake fees with 0 ratio (rewards are skipped)",
			feeRebateRatio: "0",
			feeCoins:       sdk.Coins{newStakeCoin(1000)},
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				mockWasmExecuteMsg,
			},
			rewardRecordExpected:            false,
			feeCollectorBalanceDiffExpected: "1000stake",
			rewardsBalanceDiffExpected:      "",
		},
		{
			name:           "Fail: invalid fees",
			feeRebateRatio: "0.5",
			feeCoins:       sdk.Coins{newInvalidCoin()},
			txMsgs: []sdk.Msg{
				testutils.NewMockMsg(),
				mockWasmExecuteMsg,
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create chain
			feeRewardsRatio, err := math.LegacyNewDecFromStr(tc.feeRebateRatio)
			require.NoError(t, err)

			chain := e2eTesting.NewTestChain(t, 1,
				e2eTesting.WithTxFeeRebatesRewardsRatio(feeRewardsRatio),
			)
			acc := chain.GetAccount(0)
			ctx := chain.GetContext()
			keepers := chain.GetApp().Keepers

			// Mint coins for account
			if err := tc.feeCoins.Validate(); err == nil {
				require.NoError(t, keepers.BankKeeper.MintCoins(ctx, mintTypes.ModuleName, tc.feeCoins))
				require.NoError(t, keepers.BankKeeper.SendCoinsFromModuleToAccount(ctx, mintTypes.ModuleName, acc.Address, tc.feeCoins))
			}

			// Fetch initial balances
			feeCollectorBalanceBefore := chain.GetModuleBalance(authTypes.FeeCollectorName)
			rewardsBalanceBefore := chain.GetModuleBalance(rewardsTypes.ContractRewardCollector)

			// Build transaction
			tx := testutils.NewMockFeeTx(
				testutils.WithMockFeeTxFees(tc.feeCoins),
				testutils.WithMockFeeTxPayer(acc.Address),
				testutils.WithMockFeeTxMsgs(tc.txMsgs...),
			)

			// Call the deduction Ante handler manually
			anteHandler := ante.NewDeductFeeDecorator(chain.GetAppCodec(), keepers.AccountKeeper, keepers.BankKeeper, keepers.FeeGrantKeeper, keepers.RewardsKeeper, keepers.CWFeesKeeper)
			_, err = anteHandler.AnteHandle(ctx, tx, false, testutils.NoopAnteHandler)
			if tc.errExpected {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Check final balances
			feeCollectorBalanceAfter := chain.GetModuleBalance(authTypes.FeeCollectorName)
			rewardsBalanceAfter := chain.GetModuleBalance(rewardsTypes.ContractRewardCollector)

			feeCollectorBalanceDiffReceived := feeCollectorBalanceAfter.Sub(feeCollectorBalanceBefore...) // positive
			rewardsBalanceDiffReceived := rewardsBalanceAfter.Sub(rewardsBalanceBefore...)                // positive

			feeCollectorBalanceDiffExpected, err := sdk.ParseCoinsNormalized(tc.feeCollectorBalanceDiffExpected)
			require.NoError(t, err)
			rewardsBalanceDiffExpected, err := sdk.ParseCoinsNormalized(tc.rewardsBalanceDiffExpected)
			require.NoError(t, err)

			assert.Equal(t, feeCollectorBalanceDiffExpected.String(), feeCollectorBalanceDiffReceived.String())
			assert.Equal(t, rewardsBalanceDiffExpected.String(), rewardsBalanceDiffReceived.String())

			// Check rewards record
			if tc.rewardRecordExpected {
				txID := keepers.TrackingKeeper.GetCurrentTxID(ctx)
				rewardsRecordsReceived, err := keepers.RewardsKeeper.TxRewards.Get(ctx, txID)
				require.NoError(t, err)

				assert.Equal(t, txID, rewardsRecordsReceived.TxId)
				assert.Equal(t, ctx.BlockHeight(), rewardsRecordsReceived.Height)
				assert.ElementsMatch(t, rewardsBalanceDiffExpected, rewardsRecordsReceived.FeeRewards)
			}
		})
	}
}
