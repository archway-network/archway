package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/rewards/keeper"
	"github.com/archway-network/archway/x/rewards/types"
)

func TestRewardsModuleAccountInvariant(t *testing.T) {
	type testCase struct {
		name string
		// Input
		poolCoins      sdk.Coins             // module account balance to check (might be nil)
		rewardsRecords []types.RewardsRecord // records to store (might be nil)
		// Expected output
		brokenExpected bool
	}

	accAddr, _ := e2eTesting.GenAccounts(2)
	mockTime := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	testCases := []testCase{
		{
			name:           "OK: empty pool, no records",
			poolCoins:      nil,
			rewardsRecords: nil,
		},
		{
			name: "OK: pool == records tokens",
			poolCoins: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
				sdk.NewCoin("uarch", sdk.NewInt(200)),
			),
			rewardsRecords: []types.RewardsRecord{
				{
					Id:             1,
					RewardsAddress: accAddr[0].String(),
					Rewards: sdk.NewCoins(
						sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(75)),
						sdk.NewCoin("uarch", sdk.NewInt(100)),
					),
					CalculatedHeight: 1,
					CalculatedTime:   mockTime,
				},
				{
					Id:             2,
					RewardsAddress: accAddr[0].String(),
					Rewards: sdk.NewCoins(
						sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(25)),
						sdk.NewCoin("uarch", sdk.NewInt(100)),
					),
					CalculatedHeight: 2,
					CalculatedTime:   mockTime.Add(5 * time.Second),
				},
			},
		},
		{
			name: "Fail: non-empty pool, no records",
			poolCoins: sdk.NewCoins(
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)),
			),
			rewardsRecords: nil,
			brokenExpected: true,
		},
		{
			name: "Fail: pool > records tokens",
			poolCoins: []sdk.Coin{
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100)),
				sdk.NewCoin("uarch", sdk.NewInt(200)),
			},
			rewardsRecords: []types.RewardsRecord{
				{
					Id:             1,
					RewardsAddress: accAddr[0].String(),
					Rewards: sdk.NewCoins(
						sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(50)),
					),
					CalculatedHeight: 1,
					CalculatedTime:   mockTime,
				},
			},
			brokenExpected: true,
		},
		{
			name: "Fail: pool < records tokens",
			poolCoins: []sdk.Coin{
				sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(50)),
				sdk.NewCoin("uarch", sdk.NewInt(50)),
			},
			rewardsRecords: []types.RewardsRecord{
				{
					Id:             1,
					RewardsAddress: accAddr[0].String(),
					Rewards: sdk.NewCoins(
						sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(55)),
					),
					CalculatedHeight: 1,
					CalculatedTime:   mockTime,
				},
				{
					Id:             1,
					RewardsAddress: accAddr[1].String(),
					Rewards: sdk.NewCoins(
						sdk.NewCoin("uarch", sdk.NewInt(55)),
					),
					CalculatedHeight: 2,
					CalculatedTime:   mockTime.Add(5 * time.Second),
				},
			},
			brokenExpected: true,
		},
		{
			name:      "Fail: empty pool, non-empty records tokens",
			poolCoins: nil,
			rewardsRecords: []types.RewardsRecord{
				{
					Id:             1,
					RewardsAddress: accAddr[0].String(),
					Rewards: sdk.NewCoins(
						sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)),
					),
					CalculatedHeight: 1,
					CalculatedTime:   mockTime,
				},
			},
			brokenExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create chain
			chain := e2eTesting.NewTestChain(t, 1)
			ctx := chain.GetContext()

			// Remove all pool coins (not empty due to inflation rewards for previous blocks)
			poolInitial := chain.GetApp().RewardsKeeper.UndistributedRewardsPool(ctx)
			require.NoError(t, chain.GetApp().BankKeeper.SendCoinsFromModuleToModule(ctx, types.ContractRewardCollector, mintTypes.ModuleName, poolInitial))

			// Mint coins for module account
			if tc.poolCoins != nil {
				require.NoError(t, chain.GetApp().BankKeeper.MintCoins(ctx, mintTypes.ModuleName, tc.poolCoins))
				require.NoError(t, chain.GetApp().BankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, types.ContractRewardCollector, tc.poolCoins))
			}

			// Store rewards records
			recordLastID := uint64(0)
			for _, record := range tc.rewardsRecords {
				if record.Id > recordLastID {
					recordLastID = record.Id
				}
			}

			chain.GetApp().RewardsKeeper.GetState().RewardsRecord(ctx).Import(recordLastID, tc.rewardsRecords)

			// Check invariant
			_, brokenReceived := keeper.ModuleAccountBalanceInvariant(chain.GetApp().RewardsKeeper)(ctx)
			assert.Equal(t, tc.brokenExpected, brokenReceived)
		})
	}
}
