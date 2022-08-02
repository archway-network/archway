package e2e

import (
	"encoding/json"
	"fmt"
	"time"

	voterTypes "github.com/CosmWasm/cosmwasm-go/example/voter/src/types"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func (s *E2ETestSuite) TestTxFees() {
	// Create a custom chain with fixed inflation (10%) and 10M block gas limit
	chain := e2eTesting.NewTestChain(s.T(), 1,
		e2eTesting.WithBlockGasLimit(10_000_000),
		e2eTesting.WithMintParams(
			sdk.NewDecWithPrec(1, 1), // 10%
			sdk.NewDecWithPrec(1, 1), // 10%
			uint64(60*60*8766/5),     // standard calculation
		),
	)

	senderAcc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, senderAcc)
	accAddrs, _ := e2eTesting.GenAccounts(1) // an empty account
	rewardsAddr := accAddrs[0]

	// Set metadata
	chain.SetContractMetadata(senderAcc, contractAddr, rewardsTypes.ContractMetadata{
		OwnerAddress:   senderAcc.Address.String(),
		RewardsAddress: rewardsAddr.String(),
	})

	txFee := sdk.NewCoin(sdk.DefaultBondDenom, sdk.ZeroInt())
	rewardsAccPrevBalance := chain.GetBalance(rewardsAddr)
	for i := 0; i < 10; i++ {
		// Send Tx
		var abciEvents []abci.Event
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
			gasUsed, _, events, _ := chain.SendMsgs(senderAcc, true, []sdk.Msg{&msg},
				e2eTesting.WithMsgFees(txFee),
			)

			abciEvents, txGasUsed = events, gasUsed.GasUsed
		}

		// Get gas tracked for this Tx
		var txGasTracked uint64
		{
			ctx := chain.GetContext()
			txInfosState := chain.GetApp().TrackingKeeper.GetState().TxInfoState(ctx)

			txInfos := txInfosState.GetTxInfosByBlock(ctx.BlockHeight() - 1)
			s.Require().Len(txInfos, 1)
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

		// Get rewards address balance diff
		var rewardsAddrBalanceDiff sdk.Coins
		{
			curBalance := chain.GetBalance(rewardsAddr)
			rewardsAddrBalanceDiff = curBalance.Sub(rewardsAccPrevBalance)
			rewardsAccPrevBalance = curBalance

			s.Require().Equal(rewardsAddrBalanceDiff.String(), feeRebateRewards.Add(inflationRewards).String())
		}

		// Output
		fmt.Printf("TxID %d (gas %d / %d): \t%s fees (%s infl rewards) -> \t%s rewards taken (%s + %s)\n",
			i, txGasTracked, txGasUsed,
			txFee.String(), blockRewards.String(), rewardsAddrBalanceDiff.String(),
			feeRebateRewards.String(), inflationRewards.String(),
		)

		// Increase next TxFees
		txFee.Amount = txFee.Amount.AddRaw(1)
	}
}
