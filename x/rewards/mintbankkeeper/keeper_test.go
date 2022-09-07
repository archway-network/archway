package mintbankkeeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distrTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg"
	"github.com/archway-network/archway/x/rewards/mintbankkeeper"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestMintBankKeeper(t *testing.T) {
	type testCase struct {
		name string
		// Inputs
		inflationRewardsRatio string // x/rewards inflation rewards ratio
		blockMaxGas           int64  // block max gas (consensus param)
		srcModule             string // source module name
		dstModule             string // destination module name
		transferCoins         string // coins to send [sdk.Coins]
		// Expected outputs
		errExpected                bool
		rewardRecordExpected       bool   // reward record expected to be created
		dstBalanceDiffExpected     string // expected destination module balance diff [sdk.Coins]
		rewardsBalanceDiffExpected string // expected x/rewards module balance diff [sdk.Coins]
	}

	testCases := []testCase{
		{
			name: "OK: 1000stake: Mint -> FeeCollector with 0.6 ratio to Rewards",
			//
			inflationRewardsRatio: "0.6",
			blockMaxGas:           1000,
			srcModule:             mintTypes.ModuleName,
			dstModule:             authTypes.FeeCollectorName,
			transferCoins:         "1000stake",
			//
			rewardRecordExpected:       true,
			dstBalanceDiffExpected:     "400stake",
			rewardsBalanceDiffExpected: "600stake",
		},
		{
			name: "OK: 45stake: Mint -> FeeCollector with 0.5 ratio to Rewards with Int truncated",
			//
			inflationRewardsRatio: "0.5",
			blockMaxGas:           1000,
			srcModule:             mintTypes.ModuleName,
			dstModule:             authTypes.FeeCollectorName,
			transferCoins:         "45stake",
			//
			rewardRecordExpected:       true,
			dstBalanceDiffExpected:     "23stake",
			rewardsBalanceDiffExpected: "22stake",
		},
		{
			name: "OK: 100stake: Mint -> FeeCollector with 0.99 ratio to Rewards",
			//
			inflationRewardsRatio: "0.99",
			blockMaxGas:           1000,
			srcModule:             mintTypes.ModuleName,
			dstModule:             authTypes.FeeCollectorName,
			transferCoins:         "100stake",
			//
			rewardRecordExpected:       true,
			dstBalanceDiffExpected:     "1stake",
			rewardsBalanceDiffExpected: "99stake",
		},
		{
			name: "OK: 100stake: Mint -> FeeCollector with 0.0 ratio to Rewards (no rewards)",
			//
			inflationRewardsRatio: "0",
			blockMaxGas:           1000,
			srcModule:             mintTypes.ModuleName,
			dstModule:             authTypes.FeeCollectorName,
			transferCoins:         "100stake",
			//
			rewardRecordExpected:       false,
			dstBalanceDiffExpected:     "100stake",
			rewardsBalanceDiffExpected: "",
		},
		{
			name: "OK: 100stake: Mint -> FeeCollector with 0.01 ratio to Rewards (no block gas limit)",
			//
			inflationRewardsRatio: "0.01",
			blockMaxGas:           -1,
			srcModule:             mintTypes.ModuleName,
			dstModule:             authTypes.FeeCollectorName,
			transferCoins:         "100stake",
			//
			rewardRecordExpected:       true,
			dstBalanceDiffExpected:     "99stake",
			rewardsBalanceDiffExpected: "1stake",
		},
		{
			name: "OK: 100stake: Mint -> Distr (no x/rewards involved)",
			//
			inflationRewardsRatio: "0.5",
			blockMaxGas:           -1,
			srcModule:             mintTypes.ModuleName,
			dstModule:             distrTypes.ModuleName,
			transferCoins:         "100stake",
			//
			rewardRecordExpected:       false,
			dstBalanceDiffExpected:     "100stake",
			rewardsBalanceDiffExpected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create chain
			inflationRewardsRatio, err := sdk.NewDecFromStr(tc.inflationRewardsRatio)
			require.NoError(t, err)

			chain := e2eTesting.NewTestChain(t, 1,
				e2eTesting.WithInflationRewardsRatio(inflationRewardsRatio),
				e2eTesting.WithBlockGasLimit(tc.blockMaxGas),
			)
			ctx := chain.GetContext()

			// Fetch initial balances
			srcBalanceBefore := chain.GetModuleBalance(tc.srcModule)
			dstBalanceBefore := chain.GetModuleBalance(tc.dstModule)
			rewardsBalanceBefore := chain.GetModuleBalance(rewardsTypes.ContractRewardCollector)

			// Mint funds for the source module
			transferCoins, err := sdk.ParseCoinsNormalized(tc.transferCoins)
			require.NoError(t, err)

			require.NoError(t, chain.GetApp().MintKeeper.MintCoins(ctx, transferCoins))
			require.NoError(t, chain.GetApp().BankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, tc.srcModule, transferCoins))

			// Remove rewards records which is created automagically
			chain.GetApp().RewardsKeeper.GetState().DeleteBlockRewardsCascade(ctx, ctx.BlockHeight())

			// Transfer via keeper
			k := mintbankkeeper.NewKeeper(chain.GetApp().BankKeeper, chain.GetApp().RewardsKeeper)
			err = k.SendCoinsFromModuleToModule(ctx, tc.srcModule, tc.dstModule, transferCoins)
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Check final balances
			srcBalanceAfter := chain.GetModuleBalance(tc.srcModule)
			dstBalanceAfter := chain.GetModuleBalance(tc.dstModule)
			rewardsBalanceAfter := chain.GetModuleBalance(rewardsTypes.ContractRewardCollector)

			srcBalanceDiffReceived := srcBalanceBefore.Sub(srcBalanceAfter)             // negative
			dstBalanceDiffReceived := dstBalanceAfter.Sub(dstBalanceBefore)             // positive
			rewardsBalanceDiffReceived := rewardsBalanceAfter.Sub(rewardsBalanceBefore) // positive

			dstBalanceDiffExpected, err := sdk.ParseCoinsNormalized(tc.dstBalanceDiffExpected)
			require.NoError(t, err)
			rewardsDiffExpected, err := sdk.ParseCoinsNormalized(tc.rewardsBalanceDiffExpected)
			require.NoError(t, err)

			assert.True(t, srcBalanceDiffReceived.IsZero())
			assert.Equal(t, dstBalanceDiffExpected.String(), dstBalanceDiffReceived.String())
			assert.Equal(t, rewardsDiffExpected.String(), rewardsBalanceDiffReceived.String())

			// Check rewards record
			rewardsRecordReceived, found := chain.GetApp().RewardsKeeper.GetState().BlockRewardsState(ctx).GetBlockRewards(ctx.BlockHeight())
			if !tc.rewardRecordExpected {
				require.False(t, found)
				return
			}
			require.True(t, found)

			maxGasExpected := uint64(0)
			if tc.blockMaxGas > 0 {
				maxGasExpected = uint64(tc.blockMaxGas)
			}

			assert.Equal(t, ctx.BlockHeight(), rewardsRecordReceived.Height)
			assert.Equal(t, rewardsDiffExpected.String(), rewardsRecordReceived.InflationRewards.String())
			assert.Equal(t, maxGasExpected, rewardsRecordReceived.MaxGas)

			// Check minimum consensus fee record
			minConsFeeReceived, minConfFeeFound := chain.GetApp().RewardsKeeper.GetState().MinConsensusFee(ctx).GetFee()
			if maxGasExpected == 0 || rewardsDiffExpected.IsZero() {
				assert.False(t, minConfFeeFound)
			} else {
				require.True(t, minConfFeeFound)

				minConsFeeExpected := sdk.DecCoin{
					Denom: sdk.DefaultBondDenom,
					Amount: rewardsDiffExpected[0].Amount.ToDec().Quo(
						pkg.NewDecFromUint64(maxGasExpected).Mul(
							chain.GetApp().RewardsKeeper.TxFeeRebateRatio(ctx).Sub(sdk.OneDec()),
						),
					).Neg(),
				}
				assert.Equal(t, minConsFeeExpected.String(), minConsFeeReceived.String())
			}
		})
	}
}
