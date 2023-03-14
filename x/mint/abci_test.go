package mint_test

import (
	"time"

	"github.com/archway-network/archway/x/mint/types"

	mintabci "github.com/archway-network/archway/x/mint"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const REWARDS_MODULE string = "rewards"
const WASM_MODULE string = "wasm"

func (s *KeeperTestSuite) TestBeginBlocker() {
	currentTime := time.Now()
	fiveSecAgo := currentTime.Add(-time.Second * 5)
	currentInflation := sdk.MustNewDecFromStr("0.33")
	ctx, k := s.chain.GetContext().WithBlockTime(fiveSecAgo), s.chain.GetApp().MintKeeper

	k.SetLastBlockInfo(ctx, types.LastBlockInfo{
		Inflation: currentInflation,
		Time:      &fiveSecAgo,
	})

	s.Run("OK: last mint was a 5 seconds ago. should mint some tokens and update lbi", func() {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(currentTime)

		mintabci.BeginBlocker(ctx, k)

		lbi, found := k.GetLastBlockInfo(ctx)
		s.Require().True(found)
		s.Require().EqualValues(currentTime.UTC(), lbi.Time.UTC())

		feeCollected, found := k.GetInflationForRecipient(ctx, authtypes.FeeCollectorName)
		s.Require().True(found)
		s.Require().True(feeCollected.Amount.GT(sdk.ZeroInt()))

		rewardsCollected, found := s.chain.GetApp().RewardsKeeper.GetInflationForRewards(ctx)
		s.Require().True(found)
		s.Require().True(rewardsCollected.Amount.GT(sdk.ZeroInt()))

		s.Require().True(feeCollected.IsGTE(rewardsCollected)) // feeCollected should be greater than rewards cuz we set up inflation distribution that way
	})

	s.Run("OK: last mint was just now. do not mint any in this new block", func() {
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)

		mintabci.BeginBlocker(ctx, k)

		lbi, found := k.GetLastBlockInfo(ctx)
		s.Require().True(found)
		s.Require().EqualValues(currentTime.UTC(), lbi.Time.UTC())

		_, found = k.GetInflationForRecipient(ctx, authtypes.FeeCollectorName)
		s.Require().False(found)

		_, found = s.chain.GetApp().RewardsKeeper.GetInflationForRewards(ctx)
		s.Require().False(found)
	})

	s.Run("OK: add a new recipient and check distribution now updates", func() {
		newTime := currentTime.Add(time.Second * 5) // we time travel five seconds into the future
		ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1).WithBlockTime(newTime)

		params := k.GetParams(ctx)
		params.InflationRecipients = []*types.InflationRecipient{{
			Recipient: authtypes.FeeCollectorName,
			Ratio:     sdk.MustNewDecFromStr("0.8"), // 80%
		}, {
			Recipient: REWARDS_MODULE,
			Ratio:     sdk.MustNewDecFromStr("0.1"), // 10%
		}, {
			Recipient: WASM_MODULE,
			Ratio:     sdk.MustNewDecFromStr("0.1"), // 10%
		}}
		k.SetParams(ctx, params)

		mintabci.BeginBlocker(ctx, k)

		lbi, found := k.GetLastBlockInfo(ctx)
		s.Require().True(found)
		s.Require().EqualValues(newTime.UTC(), lbi.Time.UTC())

		feeCollected, found := k.GetInflationForRecipient(ctx, authtypes.FeeCollectorName)
		s.Require().True(found)
		s.Require().True(feeCollected.Amount.GT(sdk.ZeroInt()))

		rewardsCollected, found := s.chain.GetApp().RewardsKeeper.GetInflationForRewards(ctx)
		s.Require().True(found)
		s.Require().True(rewardsCollected.Amount.GT(sdk.ZeroInt()))

		wasmCollected, found := k.GetInflationForRecipient(ctx, WASM_MODULE)
		s.Require().True(found)
		s.Require().True(wasmCollected.Amount.GT(sdk.ZeroInt()))

		s.Require().True(rewardsCollected.IsEqual(wasmCollected)) // both x/rewards and x/wasm get same amounts cuz same ratio
	})
}
