package keeper_test

import (
	mintabci "github.com/archway-network/archway/x/mint"
	"github.com/archway-network/archway/x/mint/keeper"
	"github.com/archway-network/archway/x/mint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dustin/go-humanize"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
	"time"
)

func ptime(ctx sdk.Context) *time.Time {
	t := ctx.BlockTime()
	return &t
}

func TestSimulateConstantInflation(t *testing.T) {
	denom := "denom"
	initialSupplyAmt := sdk.NewInt(1_000_000_000)
	blockTime := humanize.Day
	supply := sdk.NewCoins(sdk.NewCoin(denom, initialSupplyAmt))
	balances := map[string]sdk.Coins{}
	inflationRecp := "test"
	inflation := sdk.MustNewDecFromStr("0.10")

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
				return sdk.MustNewDecFromStr("0.65")
			},
			BondDenomFn: func(ctx sdk.Context) string {
				return denom
			},
		}),
	)
	require.NoError(t, k.SetLastBlockInfo(ctx, types.LastBlockInfo{
		Inflation: inflation,
		Time:      ptime(ctx),
	}))

	k.SetParams(ctx, types.Params{
		MinInflation:     sdk.MustNewDecFromStr("0.05"),
		MaxInflation:     sdk.MustNewDecFromStr("0.2"),
		MinBonded:        sdk.MustNewDecFromStr("0.5"),
		MaxBonded:        sdk.MustNewDecFromStr("0.9"),
		InflationChange:  sdk.MustNewDecFromStr("0.001"),
		MaxBlockDuration: humanize.Day + 1*time.Hour,
		InflationRecipients: []*types.InflationRecipient{
			{
				Recipient: inflationRecp,
				Ratio:     sdk.MustNewDecFromStr("1.0"),
			},
		},
	})
	h := int64(1)
	blocksPerYear := int64(keeper.Year / blockTime)
	for i := int64(0); i < blocksPerYear; i++ {
		ctx = ctx.WithBlockTime(ctx.BlockTime().Add(blockTime))
		ctx = ctx.WithBlockHeight(h)
		mintabci.BeginBlocker(ctx, k)
		h++
	}

	// A = P(1 + r/n)^nt <= calculates the compound interest
	// A = supply * (1 + inflation/blocksPerYear)^(blocksPerYear*1)
	expectedSupply := float64(initialSupplyAmt.Int64()) * math.Pow(1+inflation.MustFloat64()/float64(blocksPerYear), float64(blocksPerYear))
	expectedInflationaryRewards := expectedSupply - float64(initialSupplyAmt.Int64())
	t.Logf("wantRecipientInflation: %f", expectedInflationaryRewards)
	t.Logf("wantSupply: %f", expectedSupply)
	t.Logf("gotRecipientInflation: %s", balances[inflationRecp].String())
	t.Logf("gotSupply: %s", supply)
}
