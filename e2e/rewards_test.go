package e2e

import (
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
	voterCustomTypes "github.com/archway-network/voter/src/pkg/archway/custom"
	voterTypes "github.com/archway-network/voter/src/types"
)

// TestRewardsParamMaxWithdrawRecordsLimit check the x/rewards's MaxWithdrawRecords param limit (rough estimation).
// Limit is defined by the block gas limit (100M).
func (s *E2ETestSuite) TestRewardsParamMaxWithdrawRecordsLimit() {
	rewardsTypes.MaxWithdrawRecordsParamLimit = uint64(29500) // an actual value is (thisValue - 1), refer to the query below

	chain := e2eTesting.NewTestChain(s.T(), 1,
		e2eTesting.WithBlockGasLimit(100_000_000),
		e2eTesting.WithMaxWithdrawRecords(rewardsTypes.MaxWithdrawRecordsParamLimit),
	)
	bankKeeper, mintKeeper, rewardsKeeper := chain.GetApp().BankKeeper, chain.GetApp().MintKeeper, chain.GetApp().RewardsKeeper

	// Upload a new contract and set its address as the rewardsAddress
	senderAcc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, senderAcc)

	chain.SetContractMetadata(senderAcc, contractAddr, rewardsTypes.ContractMetadata{
		ContractAddress: contractAddr.String(),
		OwnerAddress:    senderAcc.Address.String(),
		RewardsAddress:  contractAddr.String(),
	})

	// Add mock rewards records for the contract and mint tokens to pass invariant checks
	recordIDs := make([]uint64, 0, rewardsTypes.MaxWithdrawRecordsParamLimit)
	{
		ctx := chain.GetContext()
		recordsState := rewardsKeeper.GetState().RewardsRecord(ctx)

		recordRewards := sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())
		for i := uint64(0); i < rewardsTypes.MaxWithdrawRecordsParamLimit; i++ {
			record := recordsState.CreateRewardsRecord(
				contractAddr,
				sdk.Coins{recordRewards},
				ctx.BlockHeight(),
				ctx.BlockTime(),
			)

			recordIDs = append(recordIDs, record.Id)
		}

		mintCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(rewardsTypes.MaxWithdrawRecordsParamLimit)))
		s.Require().NoError(mintKeeper.MintCoins(ctx, mintCoins))
		s.Require().NoError(bankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, rewardsTypes.ContractRewardCollector, mintCoins))
	}

	// Withdraw the all contract rewards
	{
		req := voterTypes.MsgExecute{
			CustomWithdrawRewards: &voterCustomTypes.WithdrawRewardsRequest{
				RecordIds: recordIDs,
			},
		}
		reqBz, err := req.MarshalJSON()
		s.Require().NoError(err)

		msg := wasmdTypes.MsgExecuteContract{
			Sender:   senderAcc.Address.String(),
			Contract: contractAddr.String(),
			Msg:      reqBz,
		}

		gasUsed, _, _, _ := chain.SendMsgs(senderAcc, true, []sdk.Msg{&msg}, e2eTesting.WithTxGasLimit(100_000_000))

		msgBz, err := msg.Marshal()
		s.Require().NoError(err)

		s.T().Log("Records:", len(recordIDs))
		s.T().Log("Msg size:", len(msgBz))
		s.T().Log("Gas used:", gasUsed.GasUsed)
	}

	// Invariants extra check
	chain.NextBlock(0)
}

// TestRewardsRecordsQueryLimit defines the x/rewards's RewardsRecords query limit (rough estimation).
// Limit is defined by the max CosmWasm VM.
func (s *E2ETestSuite) TestRewardsRecordsQueryLimit() {
	rewardsTypes.MaxRecordsQueryLimit = uint64(7716) // an actual value is (thisValue - 1), refer to the query below

	chain := e2eTesting.NewTestChain(s.T(), 1)
	bankKeeper, mintKeeper, rewardsKeeper := chain.GetApp().BankKeeper, chain.GetApp().MintKeeper, chain.GetApp().RewardsKeeper

	// Upload a new contract and set its address as the rewardsAddress
	senderAcc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, senderAcc)

	chain.SetContractMetadata(senderAcc, contractAddr, rewardsTypes.ContractMetadata{
		ContractAddress: contractAddr.String(),
		OwnerAddress:    senderAcc.Address.String(),
		RewardsAddress:  contractAddr.String(),
	})

	// Add mock rewards records for the contract and mint tokens to pass invariant checks
	var recordsExpected []rewardsTypes.RewardsRecord
	{
		ctx := chain.GetContext()
		recordsState := rewardsKeeper.GetState().RewardsRecord(ctx)

		records := make([]rewardsTypes.RewardsRecord, 0, rewardsTypes.MaxRecordsQueryLimit)
		recordRewards := sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())
		for i := uint64(0); i < rewardsTypes.MaxRecordsQueryLimit; i++ {
			record := recordsState.CreateRewardsRecord(
				contractAddr,
				sdk.Coins{recordRewards},
				ctx.BlockHeight(),
				ctx.BlockTime(),
			)

			records = append(records, record)
		}

		mintCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(rewardsTypes.MaxRecordsQueryLimit)))
		s.Require().NoError(mintKeeper.MintCoins(ctx, mintCoins))
		s.Require().NoError(bankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, rewardsTypes.ContractRewardCollector, mintCoins))

		recordsExpected = records
	}

	// Query the contract rewards and check the result
	{
		// We query one less for the "NextKey" response field to be filled up
		pageLimit := rewardsTypes.MaxRecordsQueryLimit - 1
		pageReq := query.PageRequest{
			Limit:      pageLimit,
			CountTotal: true,
		}
		recordsReceived, pageResp, respSize, _ := s.VoterGetRewardsRecords(chain, contractAddr, &pageReq, true)

		// Check page response is filled up
		s.Assert().Equal(rewardsTypes.MaxRecordsQueryLimit, pageResp.Total)
		s.Assert().NotEmpty(pageResp.NextKey)

		s.Assert().ElementsMatch(recordsExpected[:pageLimit], recordsReceived)

		s.T().Log("Response size:", respSize)
	}

	// Invariants extra check
	chain.NextBlock(0)
}
