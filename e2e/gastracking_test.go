package e2e

import (
	"encoding/json"
	"strconv"
	"time"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	voterTypes "github.com/archway-network/voter/src/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
	trackingTypes "github.com/archway-network/archway/x/tracking/types"
)

// TestGasTrackingAndRewardsDistribution tests the whole x/tracking + x/rewards chain:
//   - sets contract metadata and check an emitted event;
//   - sends WASM Execute event;
//   - checks x/tracking records created;
//   - checks x/rewards records created;
//   - checks x/rewards events emitted;
//   - checks rewards address receives distributed rewards;
func (s *E2ETestSuite) TestGasTrackingAndRewardsDistribution() {
	txFeeRebateRewardsRatio := sdk.NewDecWithPrec(5, 1)
	inflationRewardsRatio := sdk.NewDecWithPrec(5, 1)
	blockGasLimit := int64(10_000_000)

	// Setup (create new chain here with custom params)
	chain := e2eTesting.NewTestChain(s.T(), 1,
		e2eTesting.WithTxFeeRebatesRewardsRatio(txFeeRebateRewardsRatio),
		e2eTesting.WithInflationRewardsRatio(inflationRewardsRatio),
		e2eTesting.WithBlockGasLimit(blockGasLimit),
		// Artificially increase the minted inflation coin to get some rewards for the contract (otherwise contractOp gas / blockGasLimit ratio will be 0)
		e2eTesting.WithMintParams(
			sdk.NewDecWithPrec(8, 1),
			sdk.NewDecWithPrec(8, 1),
			1000000,
		),
		// Set default Tx fee for non-manual transaction like Upload / Instantiate
		e2eTesting.WithDefaultFeeAmount("10000"),
	)
	trackingKeeper, rewardsKeeper := chain.GetApp().TrackingKeeper, chain.GetApp().RewardsKeeper

	senderAcc := chain.GetAccount(0)
	contractAddr := s.VoterUploadAndInstantiate(chain, senderAcc)
	accAddrs, accPrivKeys := e2eTesting.GenAccounts(1) // an empty account

	// Inputs
	txFees := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10000)))
	rewardsAcc := e2eTesting.Account{
		Address: accAddrs[0],
		PrivKey: accPrivKeys[0],
	}

	// Collected values
	var abciEvents []abci.Event        // all ABCI events from Tx execution, BeginBlocker and EndBlockers
	var txID uint64                    // tracked Tx ID
	var txGasUsed, txGasTracked uint64 // tx gas tracking

	// Expected values (set below)
	contractMetadataExpected := rewardsTypes.ContractMetadata{
		ContractAddress: contractAddr.String(),
		OwnerAddress:    senderAcc.Address.String(),
		RewardsAddress:  rewardsAcc.Address.String(),
	}
	var contractTxRewardsExpected sdk.Coins       // contract tx fee rebate rewards expected
	var contractInflationRewardsExpected sdk.Coin // contract inflation rewards expected
	var contractTotalRewardsExpected sdk.Coins    // contract tx + inflation rewards expected
	var blockInflationRewardsExpected sdk.Coin    // block rewards expected

	// Set metadata and fetch ABCI events
	{
		msg := rewardsTypes.NewMsgSetContractMetadata(senderAcc.Address, contractAddr, &senderAcc.Address, &rewardsAcc.Address)
		_, _, events, _ := chain.SendMsgs(senderAcc, true, []sdk.Msg{msg})

		abciEvents = append(abciEvents, events...)
	}

	// Send some coins to the rewardsAcc to pay withdraw Tx fees
	rewardsAccInitialBalance := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1000000)))
	{
		s.Require().NoError(
			chain.GetApp().BankKeeper.SendCoins(
				chain.GetContext(),
				senderAcc.Address,
				rewardsAcc.Address,
				rewardsAccInitialBalance,
			),
		)
	}

	// Check x/rewards metadata set event
	s.Run("Check metadata set event", func() {
		eventContractAddr := e2eTesting.GetStringEventAttribute(abciEvents,
			"archway.rewards.v1beta1.ContractMetadataSetEvent",
			"contract_address",
		)
		eventMetadataBz := e2eTesting.GetStringEventAttribute(abciEvents,
			"archway.rewards.v1beta1.ContractMetadataSetEvent",
			"metadata",
		)

		var metadataReceived rewardsTypes.ContractMetadata
		s.Require().NoError(json.Unmarshal([]byte(eventMetadataBz), &metadataReceived))

		s.Assert().Equal(contractAddr.String(), eventContractAddr)
		s.Assert().Equal(contractMetadataExpected, metadataReceived)
	})

	// Estimate block rewards (inflation portion that should be distributed over contracts)
	// This should be done before the actual Tx to get Minter values for the Tx's block
	{
		ctx := chain.GetContext()

		mintKeeper := chain.GetApp().MintKeeper
		mintParams := mintKeeper.GetParams(ctx)

		mintedCoin := chain.GetApp().MintKeeper.GetMinter(ctx).BlockProvision(mintParams)
		inflationRewards, _ := pkg.SplitCoins(sdk.NewCoins(mintedCoin), inflationRewardsRatio)
		s.Require().Len(inflationRewards, 1)
		blockInflationRewardsExpected = inflationRewards[0]
	}

	// Send contract Execute Tx with fees, fetch ABCI events and Tx gas used
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
			Funds: sdk.NewCoins(sdk.Coin{
				Denom:  sdk.DefaultBondDenom,
				Amount: sdk.NewIntFromUint64(DefNewVotingCostAmt),
			}),
		}
		gasInfo, _, events, _ := chain.SendMsgs(senderAcc, true, []sdk.Msg{&msg},
			e2eTesting.WithMsgFees(txFees...),
		)

		txGasUsed = gasInfo.GasUsed
		abciEvents = append(abciEvents, events...)
	}

	// Check x/tracking Tx and ContractOps records
	s.Run("Check gas tracked records", func() {
		ctx := chain.GetContext()
		txInfoState := trackingKeeper.GetState().TxInfoState(ctx)
		contractOpState := trackingKeeper.GetState().ContractOpInfoState(ctx)

		// TxInfo
		txInfos := txInfoState.GetTxInfosByBlock(ctx.BlockHeight() - 1)
		s.Require().Len(txInfos, 1)
		s.Assert().NotEmpty(txInfos[0].Id)
		s.Assert().EqualValues(ctx.BlockHeight()-1, txInfos[0].Height)
		s.Assert().NotEmpty(txInfos[0].TotalGas)

		txID = txInfos[0].Id
		txGasTracked = txInfos[0].TotalGas

		// Contract operations
		contractOps := contractOpState.GetContractOpInfoByTxID(txInfos[0].Id)
		s.Require().Len(contractOps, 1)
		s.Assert().NotEmpty(contractOps[0].Id)
		s.Assert().Equal(txInfos[0].Id, contractOps[0].TxId)
		s.Assert().Equal(contractAddr.String(), contractOps[0].ContractAddress)
		s.Assert().Equal(trackingTypes.ContractOperation_CONTRACT_OPERATION_EXECUTION, contractOps[0].OperationType)
		s.Assert().NotEmpty(contractOps[0].VmGas)
		s.Assert().NotEmpty(contractOps[0].SdkGas)

		contractGasTracked := contractOps[0].VmGas + contractOps[0].SdkGas

		// Assert gas consumptions
		s.Assert().Equal(txGasTracked, contractGasTracked)
		s.Assert().LessOrEqual(txGasTracked, txGasUsed)
	})

	// Estimate contract rewards
	// This should be done after the actual Tx to get gas tracking data
	{
		// Contract fee rewards
		txFeeRewards, _ := pkg.SplitCoins(txFees, txFeeRebateRewardsRatio)
		contractTxRewardsExpected = txFeeRewards

		// Contract inflation rewards
		contractToBlockGasRatio := sdk.NewDec(int64(txGasTracked)).Quo(sdk.NewDec(blockGasLimit))
		contractInflationRewardsExpected = sdk.Coin{
			Denom:  blockInflationRewardsExpected.Denom,
			Amount: blockInflationRewardsExpected.Amount.ToDec().Mul(contractToBlockGasRatio).TruncateInt(),
		}

		// Total
		contractTotalRewardsExpected = contractTxRewardsExpected.Add(contractInflationRewardsExpected)
	}

	// Check x/rewards Tx record
	s.Run("Check tx fee rebate rewards records", func() {
		ctx := chain.GetContext()
		txRewardsState := rewardsKeeper.GetState().TxRewardsState(ctx)

		txRewards := txRewardsState.GetTxRewardsByBlock(ctx.BlockHeight() - 1)
		s.Require().Len(txRewards, 1)
		s.Assert().Equal(txID, txRewards[0].TxId)
		s.Assert().Equal(ctx.BlockHeight()-1, txRewards[0].Height)
		s.Assert().Equal(contractTxRewardsExpected.String(), sdk.NewCoins(txRewards[0].FeeRewards...).String())
	})

	// Check x/rewards Block record
	s.Run("Check block rewards record", func() {
		ctx := chain.GetContext()
		blockRewardsState := rewardsKeeper.GetState().BlockRewardsState(ctx)

		blockRewards, found := blockRewardsState.GetBlockRewards(ctx.BlockHeight() - 1)
		s.Require().True(found)
		s.Assert().Equal(ctx.BlockHeight()-1, blockRewards.Height)
		s.Assert().Equal(blockInflationRewardsExpected.String(), blockRewards.InflationRewards.String())
		s.Assert().EqualValues(blockGasLimit, blockRewards.MaxGas)
	})

	// Check x/rewards calculation event
	s.Run("Check calculation event", func() {
		eventContractAddr := e2eTesting.GetStringEventAttribute(abciEvents,
			"archway.rewards.v1beta1.ContractRewardCalculationEvent",
			"contract_address",
		)
		eventGasConsumedBz := e2eTesting.GetStringEventAttribute(abciEvents,
			"archway.rewards.v1beta1.ContractRewardCalculationEvent",
			"gas_consumed",
		)
		eventInflationRewardsBz := e2eTesting.GetStringEventAttribute(abciEvents,
			"archway.rewards.v1beta1.ContractRewardCalculationEvent",
			"inflation_rewards",
		)
		eventFeeRebateRewardsBz := e2eTesting.GetStringEventAttribute(abciEvents,
			"archway.rewards.v1beta1.ContractRewardCalculationEvent",
			"fee_rebate_rewards",
		)
		eventMetadataBz := e2eTesting.GetStringEventAttribute(abciEvents,
			"archway.rewards.v1beta1.ContractRewardCalculationEvent",
			"metadata",
		)

		gasConsumedReceived, err := strconv.ParseUint(eventGasConsumedBz, 10, 64)
		s.Require().NoError(err)

		var inflationRewardsReceived sdk.Coin
		s.Require().NoError(json.Unmarshal([]byte(eventInflationRewardsBz), &inflationRewardsReceived))

		var feeRebateRewardsReceived sdk.Coins
		s.Require().NoError(json.Unmarshal([]byte(eventFeeRebateRewardsBz), &feeRebateRewardsReceived))

		var metadataReceived rewardsTypes.ContractMetadata
		s.Require().NoError(json.Unmarshal([]byte(eventMetadataBz), &metadataReceived))

		s.Assert().Equal(contractAddr.String(), eventContractAddr)
		s.Assert().Equal(txGasTracked, gasConsumedReceived)
		s.Assert().Equal(contractInflationRewardsExpected.String(), inflationRewardsReceived.String())
		s.Assert().Equal(contractTxRewardsExpected.String(), feeRebateRewardsReceived.String())
		s.Assert().Equal(contractMetadataExpected, metadataReceived)
	})

	// Withdraw rewards and check x/rewards withdraw event (spend all account coins as fees)
	s.Run("Withdraw rewards and check distribution event", func() {
		msg := rewardsTypes.NewMsgWithdrawRewardsByLimit(rewardsAcc.Address, 1000)
		_, _, msgEvents, _ := chain.SendMsgs(rewardsAcc, true, []sdk.Msg{msg}, e2eTesting.WithMsgFees(rewardsAccInitialBalance...))

		eventRewardsAddr := e2eTesting.GetStringEventAttribute(msgEvents,
			"archway.rewards.v1beta1.RewardsWithdrawEvent",
			"reward_address",
		)
		eventRewardsBz := e2eTesting.GetStringEventAttribute(msgEvents,
			"archway.rewards.v1beta1.RewardsWithdrawEvent",
			"rewards",
		)

		var rewardsReceived sdk.Coins
		s.Require().NoError(json.Unmarshal([]byte(eventRewardsBz), &rewardsReceived))

		s.Assert().Equal(rewardsAcc.Address.String(), eventRewardsAddr)
		s.Assert().Equal(contractTotalRewardsExpected.String(), rewardsReceived.String())
	})

	// Check rewards address balance
	s.Run("Check funds transferred", func() {
		accCoins := chain.GetBalance(rewardsAcc.Address)
		s.Assert().Equal(contractTotalRewardsExpected.String(), accCoins.String())
	})
}
