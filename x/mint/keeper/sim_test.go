package keeper_test

import (
	"math"
	"strconv"
	"testing"
	"time"

	mintabci "github.com/archway-network/archway/x/mint"
	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/require"
)

const (
	DENOM                      string = "test"
	INFLATION_RECIPIENT_MODULE string = "testModule"
)

func ptime(ctx sdk.Context) *time.Time {
	t := ctx.BlockTime()
	return &t
}

func TestSimulateConstantInflation(t *testing.T) {
	initialSupplyAmt := sdk.NewInt(1_000_000_000)
	inflation := sdk.MustNewDecFromStr("0.10") // 10%
	blockTime := humanize.Day                  // One block per day

	supply := sdk.NewCoins(sdk.NewCoin(DENOM, initialSupplyAmt))
	balances := map[string]sdk.Coins{}

	k, ctx := SetupTestMintKeeper(
		t,
		SetupTestMintKeeperWithBankKeeper(MockBankKeeper{
			MintCoinsFn: func(ctx sdk.Context, name string, amt sdk.Coins) error {
				supply = supply.Add(amt...)
				return nil
			},
			GetSupplyFn: func(ctx sdk.Context, denom string) sdk.Coin {
				return sdk.NewCoin(denom, supply.AmountOf(denom))
			},
			SendCoinsFromModuleToModuleFn: func(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error {
				if _, ok := balances[recipientModule]; !ok {
					balances[recipientModule] = sdk.NewCoins()
				}
				balances[recipientModule] = balances[recipientModule].Add(amt...)
				return nil
			},
		}),
		SetupTestMintKeeperWithStakingKeeper(MockStakingKeeper{
			BondedRatioFn: func(ctx sdk.Context) sdk.Dec {
				return sdk.MustNewDecFromStr("0.65") // bonded ratio : 65%
			},
			BondDenomFn: func(ctx sdk.Context) string {
				return DENOM
			},
		}),
	)
	require.NoError(t, k.SetLastBlockInfo(ctx, types.LastBlockInfo{
		Inflation: inflation,
		Time:      ptime(ctx),
	}))

	k.SetParams(ctx, types.Params{
		MinInflation:     sdk.MustNewDecFromStr("0.05"),  // 5%
		MaxInflation:     sdk.MustNewDecFromStr("0.2"),   // 20%
		MinBonded:        sdk.MustNewDecFromStr("0.5"),   // 50%
		MaxBonded:        sdk.MustNewDecFromStr("0.9"),   // 90%
		InflationChange:  sdk.MustNewDecFromStr("0.001"), // 0.1%
		MaxBlockDuration: humanize.Day + 1*time.Hour,     // 25hours
		InflationRecipients: []*types.InflationRecipient{
			{
				Recipient: INFLATION_RECIPIENT_MODULE,
				Ratio:     sdk.MustNewDecFromStr("1.0"), // 100%
			},
		},
	})

	h := int64(1) // starting block height
	blocksPerYear := int64(keeper.Year / blockTime)
	inflationDistributed := sdk.NewInt64Coin(DENOM, 0)

	for i := int64(0); i < blocksPerYear; i++ {
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(blockTime)).WithBlockHeight(h)

		mintabci.BeginBlocker(ctx, k) // inflation is minted and distributed

		inflationReceivded, found := k.GetInflationForRecipient(ctx, INFLATION_RECIPIENT_MODULE) // keeping track of how much inflation was distributed in the block
		require.True(t, found)
		inflationDistributed = inflationDistributed.Add(inflationReceivded)
		h++
	}

	// A = P(1 + r/n)^nt <= calculates the compound interest
	// A = supply * (1 + inflation/blocksPerYear)^(blocksPerYear*1) - Calculating inflation for one year
	expectedSupply := float64(initialSupplyAmt.Int64()) * math.Pow(1+inflation.MustFloat64()/float64(blocksPerYear), float64(blocksPerYear))
	currentSupply := k.GetBondedTokenSupply(ctx)
	currentSupplyAmount := currentSupply.Amount.ToDec().MustFloat64()
	require.Equal(t, truncateFloat(expectedSupply), truncateFloat(currentSupplyAmount))

	expectedInflationaryRewards := expectedSupply - initialSupplyAmt.ToDec().MustFloat64()
	currentInflationRewardAmount := inflationDistributed.Amount.ToDec().MustFloat64()
	require.Equal(t, truncateFloat(expectedInflationaryRewards), truncateFloat(currentInflationRewardAmount))
}

func truncateFloat(f float64) string {
	return strconv.FormatFloat(f, 'g', 6, 64) // Rounding to SIX decimal places
}
