package e2e

import (
	"encoding/json"
	"fmt"
	"time"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	voterTypes "github.com/archway-network/voter/src/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// TestTxFees ensures that a transaction fees paid are less than rewards received.
// Test configures a chain based on the Archway mainnet parameters.
func (s *E2ETestSuite) TestTxFees() {
	const (
		txGasLimit        = 200_000
		txFeeAmtIncrement = 1000
	)

	coinsToStr := func(coins ...sdk.Coin) string {
		return fmt.Sprintf("%12s", e2eTesting.HumanizeCoins(6, coins...))
	}

	minConfFeeToStr := func(coin sdk.DecCoin) string {
		if coin.IsZero() {
			return "-"
		}
		return fmt.Sprintf("%8s", e2eTesting.HumanizeDecCoins(0, coin))
	}

	// Create a custom chain with fixed inflation (10%) and 10M block gas limit
	chain := e2eTesting.NewTestChain(s.T(), 1,
		// Set 1B total supply (10^9 * 10^6) (Archway mainnet param)
		e2eTesting.WithGenAccounts(1),
		e2eTesting.WithGenDefaultCoinBalance("1000000000000000"),
		// Set bonded ratio to 30%
		e2eTesting.WithBondAmount("300000000000000"),
		// Override the default Tx fee
		e2eTesting.WithDefaultFeeAmount("10000000"),
		// Set block gas limit (Archway mainnet param)
		e2eTesting.WithBlockGasLimit(100_000_000),
		// x/rewards distribution params
		e2eTesting.WithTxFeeRebatesRewardsRatio(sdk.NewDecWithPrec(5, 1)), // 50 % (Archway mainnet param)
		e2eTesting.WithInflationRewardsRatio(sdk.NewDecWithPrec(2, 1)),    // 20 % (Archway mainnet param)
		// Set constant inflation rate
		e2eTesting.WithMintParams(
			sdk.NewDecWithPrec(10, 2), // 10% (Archway mainnet param)
			sdk.NewDecWithPrec(10, 2), // 10% (Archway mainnet param)
			uint64(60*60*8766/1),      // 1 seconds block time (Archway mainnet param)
		),
	)

	// Check total supply
	{
		ctx := chain.GetContext()

		totalSupplyMaxAmtExpected, ok := sdk.NewIntFromString("1000000050000000") // small gap for minted coins for 2 blocks
		s.Require().True(ok)

		totalSupplyReceived := chain.GetApp().BankKeeper.GetSupply(ctx, sdk.DefaultBondDenom)
		s.Require().True(totalSupplyReceived.Amount.LTE(totalSupplyMaxAmtExpected), "total supply", totalSupplyReceived.String())
	}

	senderAcc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, senderAcc)

	accAddrs, accPrivKeys := e2eTesting.GenAccounts(1) // an empty account
	rewardsAcc := e2eTesting.Account{
		Address: accAddrs[0],
		PrivKey: accPrivKeys[0],
	}

	// Send some coins to the rewardsAcc to pay withdraw Tx fees
	{
		s.Require().NoError(
			chain.GetApp().BankKeeper.SendCoins(
				chain.GetContext(),
				senderAcc.Address,
				rewardsAcc.Address,
				sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000))), // 100.0
			),
		)
	}

	// Set metadata
	chain.SetContractMetadata(senderAcc, contractAddr, rewardsTypes.ContractMetadata{
		OwnerAddress:   senderAcc.Address.String(),
		RewardsAddress: rewardsAcc.Address.String(),
	})

	// Get minted inflation amount
	{
		ctx := chain.GetContext()

		mintParams := chain.GetApp().MintKeeper.GetParams(ctx)
		mintedCoin := chain.GetApp().MintKeeper.GetMinter(ctx).BlockProvision(mintParams)
		s.T().Logf("x/mint minted amount per block: %s", coinsToStr(mintedCoin))
	}

	var abciEvents []abci.Event
	txFee := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewInt(0)} // this one gonna increase
	rewardsAccPrevBalance := chain.GetBalance(rewardsAcc.Address)

	// Generate transactions and check fees (some txs might fail with InsufficientFee error)
	for i := 0; i < 100; i++ {
		// Increase next TxFees
		txFee.Amount = txFee.Amount.AddRaw(txFeeAmtIncrement)

		// Get min consensus fee for the current block and check the update event
		minConsensusFee := sdk.DecCoin{Amount: sdk.ZeroDec()}
		{
			ctx := chain.GetContext()

			if fee, found := chain.GetApp().RewardsKeeper.GetMinConsensusFee(ctx); found {
				minConsensusFee = fee

				// Check the event from the previous BeginBlocker
				if len(abciEvents) > 0 {
					eventFeeBz := e2eTesting.GetStringEventAttribute(abciEvents,
						"archway.rewards.v1beta1.MinConsensusFeeSetEvent",
						"fee",
					)

					var eventFee sdk.DecCoin
					s.Require().NoError(json.Unmarshal([]byte(eventFeeBz), &eventFee))

					s.Require().Equal(minConsensusFee.String(), eventFee.String())
				}
			}
		}

		// Send Tx
		var txGasUsed uint64
		{
			req := voterTypes.MsgExecute{
				NewVoting: &voterTypes.NewVotingRequest{
					Name:        "Test",
					VoteOptions: []string{"Yes", "No"},
					Duration:    uint64(time.Minute),
				},
			}
			reqBz, err := req.MarshalJSON()
			s.Require().NoError(err)

			msg := wasmdTypes.MsgExecuteContract{
				Sender:   senderAcc.Address.String(),
				Contract: contractAddr.String(),
				Msg:      reqBz,
				Funds:    sdk.NewCoins(sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.NewIntFromUint64(DefNewVotingCostAmt)}),
			}

			gasUsed, res, err := chain.SendMsgsRaw(senderAcc, []sdk.Msg{&msg},
				e2eTesting.WithMsgFees(txFee),
				e2eTesting.WithTxGasLimit(txGasLimit),
			)
			if err != nil {
				s.Require().ErrorIs(err, sdkErrors.ErrInsufficientFee)

				s.T().Logf("TxID %03d: %s fees (%s minConsFee): insufficient fees: %v",
					i,
					coinsToStr(txFee), minConfFeeToStr(minConsensusFee),
					err,
				)

				// Skip the block to avoid "out of gas" for this one
				abciEvents = chain.NextBlock(0)
				continue
			}

			abciEvents, txGasUsed = res.Events, gasUsed.GasUsed
		}

		// Start a new block to get rewards and tracking for the previous one
		abciEvents = append(abciEvents, chain.NextBlock(0)...)

		// Get gas tracked for this Tx
		var txGasTracked uint64
		{
			ctx := chain.GetContext()
			txInfosState := chain.GetApp().TrackingKeeper.GetState().TxInfoState(ctx)

			txInfos := txInfosState.GetTxInfosByBlock(ctx.BlockHeight() - 1)
			s.Require().GreaterOrEqual(len(txInfos), 1) // at least one Tx in the previous block (+1 for the withdrawal operation)
			txGasTracked = txInfos[0].TotalGas
		}

		// Get block rewards for prev. block
		var blockRewards sdk.Coin
		{
			ctx := chain.GetContext()
			blockRewardsState := chain.GetApp().RewardsKeeper.GetState().BlockRewardsState(ctx)

			blockRewardsInfo, found := blockRewardsState.GetBlockRewards(ctx.BlockHeight() - 1)
			s.Require().True(found)

			blockRewards = blockRewardsInfo.InflationRewards
		}

		// Get rewards for this Tx
		var inflationRewards sdk.Coin
		var feeRebateRewards sdk.Coins
		{
			eventInflationRewardsBz := e2eTesting.GetStringEventAttribute(abciEvents,
				"archway.rewards.v1beta1.ContractRewardCalculationEvent",
				"inflation_rewards",
			)
			eventFeeRebateRewardsBz := e2eTesting.GetStringEventAttribute(abciEvents,
				"archway.rewards.v1beta1.ContractRewardCalculationEvent",
				"fee_rebate_rewards",
			)

			s.Require().NoError(json.Unmarshal([]byte(eventInflationRewardsBz), &inflationRewards))
			s.Require().NoError(json.Unmarshal([]byte(eventFeeRebateRewardsBz), &feeRebateRewards))
		}

		// Withdraw rewards
		withdrawTxFees := sdk.Coin{Denom: sdk.DefaultBondDenom, Amount: sdk.ZeroInt()}
		{
			const withdrawGas = 100_000

			minConsFee, found := chain.GetApp().RewardsKeeper.GetMinConsensusFee(chain.GetContext())
			s.Require().True(found)

			withdrawTxFees.Amount = minConsFee.Amount.MulInt64(withdrawGas).RoundInt()

			msg := rewardsTypes.NewMsgWithdrawRewardsByLimit(rewardsAcc.Address, 1000)
			_, _, err := chain.SendMsgsRaw(rewardsAcc, []sdk.Msg{msg},
				e2eTesting.WithMsgFees(withdrawTxFees),
				e2eTesting.WithTxGasLimit(withdrawGas),
			)
			s.Require().NoError(err)
		}

		// Get rewards address balance diff (adjusting prev balance with fees paid)
		var rewardsAddrBalanceDiff sdk.Coins
		{
			rewardsAccPrevBalance = rewardsAccPrevBalance.Sub(sdk.Coins{withdrawTxFees})

			curBalance := chain.GetBalance(rewardsAcc.Address)
			rewardsAddrBalanceDiff = curBalance.Sub(rewardsAccPrevBalance)
			rewardsAccPrevBalance = curBalance

			s.Require().Equal(rewardsAddrBalanceDiff.String(), feeRebateRewards.Add(inflationRewards).String())
		}

		// Output
		s.T().Logf("TxID %03d (gas %06d / %06d / %06d): %s fees (%s inflRewards, %s minConsFee) -> %s rewards taken (%s + %s)",
			i, txGasTracked, txGasUsed, txGasLimit,
			coinsToStr(txFee),
			coinsToStr(blockRewards), minConfFeeToStr(minConsensusFee),
			coinsToStr(rewardsAddrBalanceDiff...), coinsToStr(feeRebateRewards...), coinsToStr(inflationRewards),
		)

		// Check rewards are lower than the fee paid
		{
			s.Assert().True(rewardsAddrBalanceDiff.IsAllLT(sdk.Coins{txFee}))
		}
	}
}
