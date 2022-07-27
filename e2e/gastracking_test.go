package e2e

import (
	voterTypes "github.com/CosmWasm/cosmwasm-go/example/voter/src/types"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	trackingTypes "github.com/archway-network/archway/x/tracking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"time"
)

// TestGasTracking_TxGasConsumption tries to check gas consumption by contract execution, but
// since we can't (or I don't know how) check the real gas consumption by WASM only, it is kind of useless.
func (s *E2ETestSuite) TestGasTracking_TxGasConsumption() {
	chain := s.chainA

	acc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, acc)

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

		gasInfo, _, _ := chain.SendMsgs(acc, true, []sdk.Msg{&msg})
		txGasUsed = gasInfo.GasUsed
	}

	// Get gas tracking data
	var trackedGas uint64
	{
		ctx := chain.GetContext()

		txInfos := chain.GetApp().TrackingKeeper.GetState().TxInfoState(ctx).GetTxInfosByBlock(ctx.BlockHeight() - 1)
		s.Require().Len(txInfos, 1)

		contractOps := chain.GetApp().TrackingKeeper.GetState().ContractOpInfoState(ctx).GetContractOpInfoByTxID(txInfos[0].Id)
		s.Require().Len(contractOps, 1)
		s.Equal(trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION, contractOps[0].OperationType)

		trackedGas = txInfos[0].TotalGas
	}

	s.Assert().Less(trackedGas, txGasUsed)
}
