package callback_test

import (
	"encoding/json"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	callbackabci "github.com/archway-network/archway/x/callback"
	callbackKeeper "github.com/archway-network/archway/x/callback/keeper"
	"github.com/archway-network/archway/x/callback/types"
)

const (
	DECREMENT_JOBID = 0
	INCREMENT_JOBID = 1
	DONOTHING_JOBID = 2
)

func TestEndBlocker(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	ctx, keeper := chain.GetContext(), chain.GetApp().Keepers.CallbackKeeper
	msgServer := callbackKeeper.NewMsgServer(keeper)
	contractAdminAcc := chain.GetAccount(0)

	// Upload and instantiate contract
	codeID := chain.UploadContract(contractAdminAcc, "../../contracts/callback-test/artifacts/callback_test.wasm", wasmdTypes.DefaultUploadAccess)
	initMsg := CallbackContractInstantiateMsg{Count: 100}
	contractAddr, _ := chain.InstantiateContract(contractAdminAcc, codeID, contractAdminAcc.Address.String(), "callback_test", nil, initMsg)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	currentBlockHeight := ctx.BlockHeight()
	callbackHeight := currentBlockHeight + 1
	futureResFee, blockResFee, txFee, err := keeper.EstimateCallbackFees(ctx, callbackHeight)
	require.NoError(t, err)
	feesToPay := futureResFee.Add(blockResFee).Add(txFee)

	reqMsg := &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           DECREMENT_JOBID,
		CallbackHeight:  callbackHeight,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            feesToPay,
	}

	_, err = msgServer.RequestCallback(sdk.WrapSDKContext(ctx), reqMsg)
	require.NoError(t, err)
	callback, err := keeper.GetCallback(ctx, reqMsg.CallbackHeight, reqMsg.ContractAddress, reqMsg.JobId)
	require.NoError(t, err)
	require.Equal(t, txFee.Amount, callback.FeeSplit.TransactionFees.Amount)
	require.Equal(t, blockResFee.Amount, callback.FeeSplit.BlockReservationFees.Amount)
	require.Equal(t, futureResFee.Amount, callback.FeeSplit.FutureReservationFees.Amount)
	require.Equal(t, math.ZeroInt(), callback.FeeSplit.SurplusFees.Amount)

	count := getCount(t, chain, ctx, contractAddr)
	require.Equal(t, initMsg.Count, count)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	require.Equal(t, ctx.BlockHeight(), callback.CallbackHeight)

	c, err := keeper.GetAllCallbacks(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(c))

	cbs, err := keeper.GetCallbacksByHeight(ctx, ctx.BlockHeight())
	require.NoError(t, err)
	require.Equal(t, 1, len(cbs))

	_ = callbackabci.EndBlocker(ctx, keeper, chain.GetApp().Keepers.WASMKeeper)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	count = getCount(t, chain, ctx, contractAddr)
	require.Equal(t, initMsg.Count-1, count)

	reqMsg.JobId = INCREMENT_JOBID
	reqMsg.CallbackHeight = ctx.BlockHeight() + 1
	_, err = msgServer.RequestCallback(sdk.WrapSDKContext(ctx), reqMsg)
	require.NoError(t, err)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	_ = callbackabci.EndBlocker(ctx, keeper, chain.GetApp().Keepers.WASMKeeper)
	count = getCount(t, chain, ctx, contractAddr)
	require.Equal(t, initMsg.Count, count)

	reqMsg.JobId = DONOTHING_JOBID
	reqMsg.CallbackHeight = ctx.BlockHeight() + 1
	_, err = msgServer.RequestCallback(sdk.WrapSDKContext(ctx), reqMsg)
	require.NoError(t, err)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	_ = callbackabci.EndBlocker(ctx, keeper, chain.GetApp().Keepers.WASMKeeper)
	count = getCount(t, chain, ctx, contractAddr)
	require.Equal(t, initMsg.Count, count)

	// When callback exceeds gas limit
	params, err := keeper.GetParams(ctx)
	require.NoError(t, err)
	err = keeper.SetParams(ctx, types.Params{
		CallbackGasLimit:               1,
		MaxBlockReservationLimit:       params.MaxBlockReservationLimit,
		MaxFutureReservationLimit:      params.MaxFutureReservationLimit,
		FutureReservationFeeMultiplier: params.FutureReservationFeeMultiplier,
		BlockReservationFeeMultiplier:  params.BlockReservationFeeMultiplier,
	})
	require.NoError(t, err)

	reqMsg.JobId = INCREMENT_JOBID
	reqMsg.CallbackHeight = ctx.BlockHeight() + 1
	_, err = msgServer.RequestCallback(sdk.WrapSDKContext(ctx), reqMsg)
	require.NoError(t, err)
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	_ = callbackabci.EndBlocker(ctx, keeper, chain.GetApp().Keepers.WASMKeeper)
	count = getCount(t, chain, ctx, contractAddr)
	require.Equal(t, initMsg.Count, count)
}

func getCount(t *testing.T, chain *e2eTesting.TestChain, ctx sdk.Context, contractAddr sdk.AccAddress) int32 {
	getCountQuery := "{\"get_count\":{}}"
	resp, err := chain.GetApp().Keepers.WASMKeeper.QuerySmart(ctx, contractAddr, []byte(getCountQuery))
	require.NoError(t, err)
	var getCountResp CallbackContractQueryMsg
	err = json.Unmarshal(resp, &getCountResp)
	require.NoError(t, err)
	return getCountResp.Count
}

type CallbackContractInstantiateMsg struct {
	Count int32 `json:"count"`
}

func (msg CallbackContractInstantiateMsg) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Count int32 `json:"count"`
	}{
		Count: msg.Count,
	})
}

type CallbackContractQueryMsg struct {
	Count int32 `json:"count"`
}
