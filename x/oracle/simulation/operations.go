package simulation

// DONTCOVER

import (
	"math/rand"
	"strings"

	"cosmossdk.io/math"
	"github.com/CosmWasm/wasmd/app/params"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"

	"github.com/archway-network/archway/x/common/asset"
	"github.com/archway-network/archway/x/common/denoms"

	helpers "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/archway-network/archway/x/oracle/keeper"
	"github.com/archway-network/archway/x/oracle/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgAggregateExchangeRatePrevote = "op_weight_msg_exchange_rate_aggregate_prevote"
	OpWeightMsgAggregateExchangeRateVote    = "op_weight_msg_exchange_rate_aggregate_vote"
	OpWeightMsgDelegateFeedConsent          = "op_weight_msg_exchange_feed_consent"

	salt = "1234"
)

var (
	whitelist                     = []asset.Pair{asset.Registry.Pair(denoms.BTC, denoms.NUSD), asset.Registry.Pair(denoms.ETH, denoms.NUSD), asset.Registry.Pair(denoms.NIBI, denoms.NUSD)}
	voteHashMap map[string]string = make(map[string]string)
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgAggregateExchangeRatePrevote int
		weightMsgAggregateExchangeRateVote    int
		weightMsgDelegateFeedConsent          int
	)
	appParams.GetOrGenerate(OpWeightMsgAggregateExchangeRatePrevote, &weightMsgAggregateExchangeRatePrevote, nil,
		func(_ *rand.Rand) {
			weightMsgAggregateExchangeRatePrevote = params.DefaultWeightMsgSend * 2
		},
	)

	appParams.GetOrGenerate(OpWeightMsgAggregateExchangeRateVote, &weightMsgAggregateExchangeRateVote, nil,
		func(_ *rand.Rand) {
			weightMsgAggregateExchangeRateVote = params.DefaultWeightMsgSend * 2
		},
	)

	appParams.GetOrGenerate(OpWeightMsgDelegateFeedConsent, &weightMsgDelegateFeedConsent, nil,
		func(_ *rand.Rand) {
			weightMsgDelegateFeedConsent = params.DefaultWeightMsgDelegate // TODO: temp fix
		},
	)

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgAggregateExchangeRatePrevote,
			SimulateMsgAggregateExchangeRatePrevote(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgAggregateExchangeRateVote,
			SimulateMsgAggregateExchangeRateVote(ak, bk, k),
		),
		simulation.NewWeightedOperation(
			weightMsgDelegateFeedConsent,
			SimulateMsgDelegateFeedConsent(ak, bk, k),
		),
	}
}

// SimulateMsgAggregateExchangeRatePrevote generates a MsgAggregateExchangeRatePrevote with random values.
// nolint: funlen
func SimulateMsgAggregateExchangeRatePrevote(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)

		// ensure the validator exists
		val, err := k.StakingKeeper.Validator(ctx, address)
		if err != nil || !val.IsBonded() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAggregateExchangeRatePrevote, "unable to find validator"), nil, err
		}

		exchangeRatesStr := ""
		for _, pair := range whitelist {
			price := math.LegacyNewDecWithPrec(int64(simtypes.RandIntBetween(r, 1, 10000)), int64(1))
			exchangeRatesStr += price.String() + pair.String() + ","
		}

		exchangeRatesStr = strings.TrimRight(exchangeRatesStr, ",")
		voteHash := types.GetAggregateVoteHash(salt, exchangeRatesStr, address)

		feederAddrBytes, err := k.FeederDelegations.Get(ctx, address)
		var feederAddr sdk.AccAddress
		if err == nil {
			feederAddr = sdk.AccAddress(sdk.ValAddress(feederAddrBytes))
		} else {
			feederAddr = sdk.AccAddress(address)
		}
		feederSimAccount, _ := simtypes.FindAccount(accs, feederAddr)

		feederAccount := ak.GetAccount(ctx, feederAddr)
		spendable := bk.SpendableCoins(ctx, feederAccount.GetAddress())

		fees, err := simtypes.RandomFees(r, ctx, spendable)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAggregateExchangeRatePrevote, "unable to generate fees"), nil, err
		}

		msg := types.NewMsgAggregateExchangeRatePrevote(voteHash, feederAddr, address)

		txGen := testutil.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{feederAccount.GetAccountNumber()},
			[]uint64{feederAccount.GetSequence()},
			feederSimAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		voteHashMap[address.String()] = exchangeRatesStr

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgAggregateExchangeRateVote generates a MsgAggregateExchangeRateVote with random values.
// nolint: funlen
func SimulateMsgAggregateExchangeRateVote(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		address := sdk.ValAddress(simAccount.Address)

		// ensure the validator exists
		val, err := k.StakingKeeper.Validator(ctx, address)
		if err != nil || !val.IsBonded() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAggregateExchangeRateVote, "unable to find validator"), nil, err
		}

		// ensure vote hash exists
		exchangeRatesStr, ok := voteHashMap[address.String()]
		if !ok {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAggregateExchangeRateVote, "vote hash not exists"), nil, nil
		}

		// get prevote
		prevote, err := k.Prevotes.Get(ctx, address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAggregateExchangeRateVote, "prevote not found"), nil, nil
		}

		params, _ := k.Params.Get(ctx)
		if (uint64(ctx.BlockHeight())/params.VotePeriod)-(prevote.SubmitBlock/params.VotePeriod) != 1 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAggregateExchangeRateVote, "reveal period of submitted vote do not match with registered prevote"), nil, nil
		}

		feederAddrBytes, err := k.FeederDelegations.Get(ctx, address)
		var feederAddr sdk.AccAddress
		if err == nil {
			feederAddr = sdk.AccAddress(sdk.ValAddress(feederAddrBytes))
		} else {
			feederAddr = sdk.AccAddress(address)
		}
		feederSimAccount, _ := simtypes.FindAccount(accs, feederAddr)
		feederAccount := ak.GetAccount(ctx, feederAddr)
		spendableCoins := bk.SpendableCoins(ctx, feederAddr)

		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAggregateExchangeRateVote, "unable to generate fees"), nil, err
		}

		msg := types.NewMsgAggregateExchangeRateVote(salt, exchangeRatesStr, feederAddr, address)

		txGen := testutil.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{feederAccount.GetAccountNumber()},
			[]uint64{feederAccount.GetSequence()},
			feederSimAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}

// SimulateMsgDelegateFeedConsent generates a MsgDelegateFeedConsent with random values.
// nolint: funlen
func SimulateMsgDelegateFeedConsent(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		delegateAccount, _ := simtypes.RandomAcc(r, accs)
		valAddress := sdk.ValAddress(simAccount.Address)
		delegateValAddress := sdk.ValAddress(delegateAccount.Address)
		account := ak.GetAccount(ctx, simAccount.Address)

		// ensure the validator exists
		_, err := k.StakingKeeper.Validator(ctx, valAddress)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegateFeedConsent, "unable to find validator"), nil, err
		}

		// ensure the target address is not a validator
		_, err = k.StakingKeeper.Validator(ctx, delegateValAddress)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDelegateFeedConsent, "unable to delegate to validator"), nil, err
		}

		spendableCoins := bk.SpendableCoins(ctx, account.GetAddress())
		fees, err := simtypes.RandomFees(r, ctx, spendableCoins)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgAggregateExchangeRateVote, "unable to generate fees"), nil, err
		}

		msg := types.NewMsgDelegateFeedConsent(valAddress, delegateAccount.Address)

		txGen := testutil.MakeTestEncodingConfig().TxConfig
		tx, err := helpers.GenSignedMockTx(
			r,
			txGen,
			[]sdk.Msg{msg},
			fees,
			helpers.DefaultGenTxGas,
			chainID,
			[]uint64{account.GetAccountNumber()},
			[]uint64{account.GetSequence()},
			simAccount.PrivKey,
		)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to generate mock tx"), nil, err
		}

		_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "unable to deliver tx"), nil, err
		}

		return simtypes.NewOperationMsg(msg, true, ""), nil, nil
	}
}
