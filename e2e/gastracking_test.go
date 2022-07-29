package e2e

import (
	voterTypes "github.com/CosmWasm/cosmwasm-go/example/voter/src/types"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
	trackingTypes "github.com/archway-network/archway/x/tracking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
	"time"
)

// TestGasTracking_TxGasConsumption tries to check gas consumption by contract execution, but
// since we can't (or I don't know how) check the real gas consumption by WASM only, it is kind of useless.
// TODO: modify the Voter (or create an other contract) with a pure "Execute" call without bank involvement.
func (s *E2ETestSuite) TestGasTracking_TxGasConsumption() {
	chain := s.chainA

	acc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc)

	// Set metadata to get rewards
	{
		chain.SetContractMetadata(acc, contractAddr, rewardsTypes.ContractMetadata{
			OwnerAddress:   acc.Address.String(),
			RewardsAddress: acc.Address.String(),
		})
	}

	// Send Tx manually to get Tx results
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
			Sender:   acc.Address.String(),
			Contract: contractAddr.String(),
			Msg:      reqBz,
			Funds: sdk.NewCoins(sdk.Coin{
				Denom:  sdk.DefaultBondDenom,
				Amount: sdk.NewIntFromUint64(DefNewVotingCostAmt),
			}),
		}

		gasInfo, _, events, _ := chain.SendMsgs(acc, true, []sdk.Msg{&msg})
		txGasUsed = gasInfo.GasUsed

		// TODO: add proper event checks
		calcEventContractAddr, err := strconv.Unquote(
			e2eTesting.GetStringEventAttribute(events,
				"archway.rewards.v1beta1.ContractRewardCalculationEvent",
				"contract_address",
			),
		)
		s.Require().NoError(err)
		s.Assert().Equal(contractAddr.String(), calcEventContractAddr)
	}

	// Get gas tracking data
	var trackedGas uint64
	{
		ctx := chain.GetContext()

		txInfos := chain.GetApp().TrackingKeeper.GetState().TxInfoState(ctx).GetTxInfosByBlock(ctx.BlockHeight() - 1)
		s.Require().Len(txInfos, 1)

		contractOps := chain.GetApp().TrackingKeeper.GetState().ContractOpInfoState(ctx).GetContractOpInfoByTxID(txInfos[0].Id)
		s.Require().Len(contractOps, 1)
		s.Assert().Equal(trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION, contractOps[0].OperationType)
		s.Assert().Equal(contractOps[0].VmGas+contractOps[0].SdkGas, txInfos[0].TotalGas)

		trackedGas = txInfos[0].TotalGas
	}

	s.Assert().NotEmpty(trackedGas)
	s.Assert().Less(trackedGas, txGasUsed)
	s.T().Log("Gas consumption:", trackedGas, ">", txGasUsed)
}
