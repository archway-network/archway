package e2e

import (
	"time"

	"github.com/stretchr/testify/require"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	voterCustomTypes "github.com/archway-network/voter/src/pkg/archway/custom"
	voterTypes "github.com/archway-network/voter/src/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// TestRewardsWithdrawProfitAndFees ensures that Tx fees spent for withdrawing rewards are lower than withdraw Tx fee paid.
// Test uses the "new voting" Execute message as a sample for a Tx with low rewards (the Voter contract only adds a new obj to its state + JSON marshaling).
// This sample reward is used to create a bunch of rewards records (creating it directly saves a lot of time comparing to actually sending msgs).
// Test then withdraws records in batches (by limit and by IDs) using gas and Tx fee estimations.
func (s *E2ETestSuite) TestRewardsWithdrawProfitAndFees() {
	const (
		recordsLen     = 50000 // records with a sample rewards amt to be generated
		batchIncStep   = 300   // withdraw batch size increment
		batchStartSize = 100   // withdraw batch size start value
	)

	// Create a custom chain with "close to mainnet" params
	chain := e2eTesting.NewTestChain(s.T(), 1,
		// Set 1B total supply (10^9 * 10^6)
		e2eTesting.WithGenAccounts(1),
		e2eTesting.WithGenDefaultCoinBalance("1000000000000000"),
		// Set bonded ratio to 30%
		e2eTesting.WithBondAmount("300000000000000"),
		// Override the default Tx fee
		e2eTesting.WithDefaultFeeAmount("10000000"),
		// Set block gas limit (Archway mainnet param)
		e2eTesting.WithBlockGasLimit(100_000_000),
		// x/rewards distribution params
		e2eTesting.WithTxFeeRebatesRewardsRatio(sdk.NewDecWithPrec(5, 1)),
		e2eTesting.WithInflationRewardsRatio(sdk.NewDecWithPrec(2, 1)),
		// Set constant inflation rate
		e2eTesting.WithMintParams(
			sdk.NewDecWithPrec(10, 2), // 10%
			sdk.NewDecWithPrec(10, 2), // 10%
			uint64(60*60*8766/1),      // 1 seconds block time
		),
	)
	trackingKeeper, rewardsKeeper := chain.GetApp().TrackingKeeper, chain.GetApp().RewardsKeeper
	chain.NextBlock(0)

	// Upload a new contract and set its address as the rewardsAddress
	senderAcc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, senderAcc)

	chain.SetContractMetadata(senderAcc, contractAddr, rewardsTypes.ContractMetadata{
		ContractAddress: contractAddr.String(),
		OwnerAddress:    senderAcc.Address.String(),
		RewardsAddress:  senderAcc.Address.String(),
	})

	// Send sdk.Msg helper which estimates gas, adjusts it and sets the Tx fee
	sendMsg := func(msg sdk.Msg) (gasEstimated, gasUsed uint64, txFees sdk.Coins) {
		// Simulate msg
		gasEstInfo, _, _, _ := chain.SendMsgs(senderAcc, true, []sdk.Msg{msg},
			e2eTesting.WithSimulation(),
		)
		gasEstimated = gasEstInfo.GasUsed
		gasAdjusted := uint64(float64(gasEstimated) * 1.1)

		// Estimate Tx fees
		gasPrice, ok := rewardsKeeper.GetMinConsensusFee(chain.GetContext())
		s.Require().True(ok)

		txFees = sdk.NewCoins(
			sdk.NewCoin(
				gasPrice.Denom,
				gasPrice.Amount.MulInt64(int64(gasAdjusted)).RoundInt(),
			),
		)

		// Deliver msg
		gasUsedInfo, _, _, _ := chain.SendMsgs(senderAcc, true, []sdk.Msg{msg},
			e2eTesting.WithTxGasLimit(gasAdjusted),
			e2eTesting.WithMsgFees(txFees...),
		)
		gasUsed = gasUsedInfo.GasUsed

		return
	}

	// Create a new voting
	var recordRewards sdk.Coins
	{
		req := voterTypes.MsgExecute{
			NewVoting: &voterTypes.NewVotingRequest{
				Name:        "Test",
				VoteOptions: []string{"A", "B", "C"},
				Duration:    uint64(60 * time.Second),
			},
		}
		reqBz, err := req.MarshalJSON()
		s.Require().NoError(err)

		msg := wasmdTypes.MsgExecuteContract{
			Sender:   senderAcc.Address.String(),
			Contract: contractAddr.String(),
			Msg:      reqBz,
			Funds: sdk.NewCoins(sdk.Coin{
				Denom:  sdk.DefaultBondDenom,
				Amount: sdk.NewIntFromUint64(DefNewVotingCostAmt),
			}),
		}

		gasEstimated, gasUsed, txFees := sendMsg(&msg)
		s.T().Logf("New voting: msg: gasEst=%d, gasUsed=%d, txFees=%s", gasEstimated, gasUsed, e2eTesting.HumanizeCoins(6, txFees...))
	}

	// Get a sample rewards amount and tracking data
	{
		ctx := chain.GetContext()

		gasUnitPrice, found := rewardsKeeper.GetMinConsensusFee(ctx)
		s.Require().True(found)

		records, _, err := rewardsKeeper.GetRewardsRecords(ctx, senderAcc.Address, nil)
		s.Require().NoError(err)
		s.Require().Len(records, 1)
		record := records[0]
		s.Require().EqualValues(1, record.Id)

		trackingBlock := trackingKeeper.GetBlockTrackingInfo(ctx, record.CalculatedHeight)
		s.Require().Len(trackingBlock.Txs, 1)
		trackingTx := trackingBlock.Txs[0]
		s.Require().Len(trackingTx.ContractOperations, 1)
		trackingOp := trackingTx.ContractOperations[0]
		s.Require().Equal(trackingOp.ContractAddress, contractAddr.String())

		rewardsBlock, found := rewardsKeeper.GetState().BlockRewardsState(ctx).GetBlockRewards(record.CalculatedHeight)
		s.Require().True(found)

		rewardsTxs := rewardsKeeper.GetState().TxRewardsState(ctx).GetTxRewardsByBlock(record.CalculatedHeight)
		s.Require().Len(rewardsTxs, 1)
		rewardsTx := rewardsTxs[0]
		s.Require().EqualValues(trackingTx.Info.Id, rewardsTx.TxId)

		s.T().Logf("New voting: tracking: VM / SDK gas:  %d / %d", trackingOp.VmGas, trackingOp.SdkGas)

		s.T().Logf("Gas unit price: %s", gasUnitPrice)
		s.T().Logf("Block inflationary rewards / gas limit: %s / %d", e2eTesting.HumanizeCoins(6, rewardsBlock.InflationRewards), rewardsBlock.MaxGas)
		s.T().Logf("New voting: fee rewards: %s", e2eTesting.HumanizeCoins(6, rewardsTx.FeeRewards...))

		s.T().Logf("New voting: rewards: %s", e2eTesting.HumanizeCoins(6, record.Rewards...))

		recordRewards = records[0].Rewards
	}

	// Create a bunch of mock reward records
	{
		ctx := chain.GetContext()
		recordsState := rewardsKeeper.GetState().RewardsRecord(ctx)

		// Create records
		coinsToMint := sdk.NewCoins()
		for i := 1; i < recordsLen; i++ {
			record := recordsState.CreateRewardsRecord(senderAcc.Address, recordRewards, ctx.BlockHeight(), ctx.BlockTime())
			s.Require().EqualValues(i+1, record.Id)
			coinsToMint = coinsToMint.Add(recordRewards...)
		}

		// Mint rewards coins
		s.Require().NoError(chain.GetApp().MintKeeper.MintCoins(ctx, coinsToMint))
		s.Require().NoError(chain.GetApp().BankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, rewardsTypes.ContractRewardCollector, coinsToMint))

		// Invariants check (just in case)
		chain.NextBlock(0)
	}

	//
	batchStartRecordID, batchSize := 1, batchStartSize
	for {
		handleBatch := func(mode string, buildMsg func(startID, endID int) sdk.Msg) bool {
			batchEndRecordID := batchStartRecordID + batchSize
			if batchEndRecordID >= recordsLen {
				return false
			}

			// Send msg
			msg := buildMsg(batchStartRecordID, batchEndRecordID)
			gasEstimated, gasUsed, txFees := sendMsg(msg)

			// Calculate rewards received
			rewards := sdk.NewCoins()
			for i := 0; i < batchSize; i++ {
				rewards = rewards.Add(recordRewards...)
			}

			// Results
			s.Assert().True(rewards.IsAllGTE(txFees))
			s.T().Logf("%4d: %5s: gasEst=%9d, gasUsed=%9d, txFees=%s, \trewards=%s",
				batchSize,
				mode,
				gasEstimated, gasUsed,
				e2eTesting.HumanizeCoins(6, txFees...),
				e2eTesting.HumanizeCoins(6, rewards...),
			)

			// Next batch params
			batchStartRecordID = batchEndRecordID

			return true
		}

		msgByLimitBuilder := func(startID, endID int) sdk.Msg {
			return rewardsTypes.NewMsgWithdrawRewardsByLimit(senderAcc.Address, uint64(endID-startID))
		}

		msgByIDsBuilder := func(startID, endID int) sdk.Msg {
			batchIDs := make([]uint64, 0, endID-startID)
			for id := startID; id < endID; id++ {
				batchIDs = append(batchIDs, uint64(id))
			}
			return rewardsTypes.NewMsgWithdrawRewardsByIDs(senderAcc.Address, batchIDs)
		}

		if !handleBatch("Limit", msgByLimitBuilder) || !handleBatch("IDs", msgByIDsBuilder) {
			break
		}
		batchSize += batchIncStep
	}
}

// TestRewardsParamMaxWithdrawRecordsLimit check the x/rewards's MaxWithdrawRecords param limit (rough estimation).
// Limit is defined by the block gas limit (100M).
func (s *E2ETestSuite) TestRewardsParamMaxWithdrawRecordsLimit() {
	s.T().Skip("Skipped to save CI time, should be used manually to estimate a new limit")

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
	s.T().Skip("Skipped to save CI time, should be used manually to estimate a new limit")

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

// TestTXFailsAfterAnteHandler tests when a TX succeeds at ante handler level, but fails at msg exec level
// which means both tracking and rewards ante run, but then no concrete rewards or tracking happen.
func (s *E2ETestSuite) TestTXFailsAfterAnteHandler() {
	// Create a custom chain with "close to mainnet" params
	chain := e2eTesting.NewTestChain(s.T(), 1,
		// Set 1B total supply (10^9 * 10^6)
		e2eTesting.WithGenAccounts(1),
		e2eTesting.WithGenDefaultCoinBalance("1000000000000000"),
		// Set bonded ratio to 30%
		e2eTesting.WithBondAmount("300000000000000"),
		// Override the default Tx fee
		e2eTesting.WithDefaultFeeAmount("10000000"),
		// Set block gas limit (Archway mainnet param)
		e2eTesting.WithBlockGasLimit(100_000_000),
		// x/rewards distribution params
		e2eTesting.WithTxFeeRebatesRewardsRatio(sdk.NewDecWithPrec(5, 1)),
		e2eTesting.WithInflationRewardsRatio(sdk.NewDecWithPrec(2, 1)),
		// Set constant inflation rate
		e2eTesting.WithMintParams(
			sdk.NewDecWithPrec(10, 2), // 10%
			sdk.NewDecWithPrec(10, 2), // 10%
			uint64(60*60*8766/1),      // 1 seconds block time
		),
	)
	rewardsKeeper := chain.GetApp().RewardsKeeper

	// Upload a new contract and set its address as the rewardsAddress
	senderAcc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, senderAcc)

	chain.SetContractMetadata(senderAcc, contractAddr, rewardsTypes.ContractMetadata{
		ContractAddress: contractAddr.String(),
		OwnerAddress:    senderAcc.Address.String(),
		RewardsAddress:  contractAddr.String(),
	})

	err := chain.GetApp().RewardsKeeper.SetFlatFee(chain.GetContext(), senderAcc.Address, rewardsTypes.FlatFee{
		ContractAddress: contractAddr.String(),
		FlatFee:         sdk.NewInt64Coin("stake", 1000),
	})
	require.NoError(s.T(), err)

	sendMsg := func(msg sdk.Msg, passes bool) (gasEstimated, gasUsed uint64, txFees sdk.Coins) {
		// Simulate msg
		_, _, _, _ = chain.SendMsgs(senderAcc, passes, []sdk.Msg{msg})
		gasEstimated = 0
		gasAdjusted := uint64(float64(gasEstimated) * 1.1)

		// Estimate Tx fees
		gasPrice, ok := rewardsKeeper.GetMinConsensusFee(chain.GetContext())
		s.Require().True(ok)

		txFees = sdk.NewCoins(
			sdk.NewCoin(
				gasPrice.Denom,
				gasPrice.Amount.MulInt64(int64(gasAdjusted)).RoundInt(),
			),
		)

		// Deliver msg
		gasUsedInfo, _, _, _ := chain.SendMsgs(senderAcc, passes, []sdk.Msg{msg},
			e2eTesting.WithTxGasLimit(gasAdjusted),
			e2eTesting.WithMsgFees(txFees...),
		)
		gasUsed = gasUsedInfo.GasUsed

		return
	}
	rk := chain.GetApp().RewardsKeeper

	// send a message that passes the ante handler but not the wasm execution step
	sendMsg(&wasmdTypes.MsgExecuteContract{
		Sender:   senderAcc.Address.String(),
		Contract: contractAddr.String(),
		Msg:      []byte(`{"fail": {}}`),
		Funds:    nil,
	}, false)

	chain.NextBlock(1 * time.Second)

	// no rewards because the TX failed.
	rewards := rk.GetState().RewardsRecord(chain.GetContext()).GetRewardsRecordByRewardsAddress(contractAddr)
	require.Empty(s.T(), rewards)
}

// TestSubMsgRevert tests when a contract calls another contract but the sub message reverts,
// and the execution of the caller contract still proceeds because the sub message is sent with
// a reply on error flag.
func (s *E2ETestSuite) TestSubMsgRevert() {
	// Create a custom chain with "close to mainnet" params
	chain := e2eTesting.NewTestChain(s.T(), 1,
		// Set 1B total supply (10^9 * 10^6)
		e2eTesting.WithGenAccounts(2),
		e2eTesting.WithGenDefaultCoinBalance("1000000000000000"),
		// Set bonded ratio to 30%
		e2eTesting.WithBondAmount("300000000000000"),
		// Override the default Tx fee
		e2eTesting.WithDefaultFeeAmount("10000000"),
		// Set block gas limit (Archway mainnet param)
		e2eTesting.WithBlockGasLimit(100_000_000),
		// x/rewards distribution params
		e2eTesting.WithTxFeeRebatesRewardsRatio(sdk.NewDecWithPrec(5, 1)),
		e2eTesting.WithInflationRewardsRatio(sdk.NewDecWithPrec(2, 1)),
		// Set constant inflation rate
		e2eTesting.WithMintParams(
			sdk.NewDecWithPrec(10, 2), // 10%
			sdk.NewDecWithPrec(10, 2), // 10%
			uint64(60*60*8766/1),      // 1 seconds block time
		),
	)
	rewardsKeeper := chain.GetApp().RewardsKeeper

	// Upload a new contract and set its address as the rewardsAddress
	senderAcc := chain.GetAccount(0)
	calledAcc := chain.GetAccount(1)
	contractAddr := s.VoterUploadAndInstantiate(chain, senderAcc)
	calledContractAddr := s.VoterUploadAndInstantiate(chain, calledAcc)

	chain.SetContractMetadata(senderAcc, contractAddr, rewardsTypes.ContractMetadata{
		ContractAddress: contractAddr.String(),
		OwnerAddress:    senderAcc.Address.String(),
		RewardsAddress:  contractAddr.String(),
	})

	chain.SetContractMetadata(calledAcc, calledContractAddr, rewardsTypes.ContractMetadata{
		ContractAddress: calledContractAddr.String(),
		OwnerAddress:    calledAcc.Address.String(),
		RewardsAddress:  calledContractAddr.String(),
	})

	sendMsg := func(msg sdk.Msg, passes bool) (gasEstimated, gasUsed uint64, txFees sdk.Coins) {
		// Simulate msg
		gasEstInfo, _, _, _ := chain.SendMsgs(senderAcc, passes, []sdk.Msg{msg},
			e2eTesting.WithSimulation(),
		)
		gasEstimated = gasEstInfo.GasUsed
		gasAdjusted := uint64(float64(gasEstimated) * 1.1)

		// Estimate Tx fees
		gasPrice, ok := rewardsKeeper.GetMinConsensusFee(chain.GetContext())
		s.Require().True(ok)

		txFees = sdk.NewCoins(
			sdk.NewCoin(
				gasPrice.Denom,
				gasPrice.Amount.MulInt64(int64(gasAdjusted)).RoundInt(),
			),
		)

		// Deliver msg
		gasUsedInfo, _, _, _ := chain.SendMsgs(senderAcc, passes, []sdk.Msg{msg},
			e2eTesting.WithTxGasLimit(gasAdjusted),
			e2eTesting.WithMsgFees(txFees...),
		)
		gasUsed = gasUsedInfo.GasUsed

		return
	}
	rk := chain.GetApp().RewardsKeeper

	// send a message that passes the ante handler but not the wasm execution step
	sendMsg(&wasmdTypes.MsgExecuteContract{
		Sender:   senderAcc.Address.String(),
		Contract: contractAddr.String(),
		Msg:      []byte(`{"reply_on_error": "` + calledContractAddr.String() + `"}`),
		Funds:    nil,
	}, true)

	chain.NextBlock(1 * time.Second)

	// has rewards because of reply on error
	rewards := rk.GetState().RewardsRecord(chain.GetContext()).GetRewardsRecordByRewardsAddress(contractAddr)
	require.NotEmpty(s.T(), rewards)
	// does not have rewards because it failed
	rewards = rk.GetState().RewardsRecord(chain.GetContext()).GetRewardsRecordByRewardsAddress(calledContractAddr)
	require.Empty(s.T(), rewards)
}
