package mint_test

import (
	"time"

	"github.com/archway-network/archway/x/mint/types"

	mintabci "github.com/archway-network/archway/x/mint"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const REWARDSMODULE string = "rewards"

func (s *KeeperTestSuite) TestBeginBlocker() {
	currentTime := time.Now()
	fiveSecAgo := currentTime.Add(-time.Second * 5)
	currentInflation := sdk.MustNewDecFromStr("0.33")
	ctx, k := s.chain.GetContext().WithBlockTime(fiveSecAgo), s.chain.GetApp().MintKeeper
	params := getTestParams()
	k.SetParams(ctx, params)

	k.SetLastBlockInfo(ctx, types.LastBlockInfo{
		Inflation: currentInflation,
		Time:      &fiveSecAgo,
	})

	s.Run("OK: last mint was just now. should not mint any tokens", func() {
		mintabci.BeginBlocker(ctx, k)

		_, found := k.GetInflationForRecipient(ctx, authtypes.FeeCollectorName)
		s.Require().False(found)
		_, found = k.GetInflationForRecipient(ctx, REWARDSMODULE)
		s.Require().False(found)
	})

	s.Run("OK: last mint was a 5 seconds ago. should mint some tokens and update lbi", func() {
		ctx = ctx.WithBlockTime(currentTime)

		mintabci.BeginBlocker(ctx, k)

		lbi, found := k.GetLastBlockInfo(ctx)
		s.Require().True(found)
		s.Require().EqualValues(currentTime.UTC(), lbi.Time.UTC())

		feeCollected, found := k.GetInflationForRecipient(ctx, authtypes.FeeCollectorName)
		s.Require().True(found)
		s.Require().True(feeCollected.Amount.GT(sdk.ZeroInt()))

		rewardsCollected, found := k.GetInflationForRecipient(ctx, REWARDSMODULE)
		s.Require().True(found)
		s.Require().True(rewardsCollected.Amount.GT(sdk.ZeroInt()))

		s.Require().True(feeCollected.IsGTE(rewardsCollected)) // feeCollected should be greater than rewards
	})
}

func getTestParams() types.Params {
	params := types.NewParams(
		sdk.MustNewDecFromStr("0.1"), sdk.OneDec(), // inflation
		sdk.ZeroDec(), sdk.OneDec(), // bonded
		sdk.MustNewDecFromStr("0.1"), // inflation change
		time.Minute,
		[]*types.InflationRecipient{{
			Recipient: authtypes.FeeCollectorName,
			Ratio:     sdk.MustNewDecFromStr("0.9"), // 90%
		}, {
			Recipient: REWARDSMODULE,
			Ratio:     sdk.MustNewDecFromStr("0.1"), // 10%
		}})
	return params
}
