package callback_test

import (
	"encoding/json"
	"testing"

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
	// The test contract is based on the default counter contract and behaves the following way:
	// When job_id = 1, it increments the count value
	// When job_id = 0, it decrements the count value
	// For any other job_id, it does nothing
	codeID := chain.UploadContract(contractAdminAcc, "../../contracts/callback-test/artifacts/callback_test.wasm", wasmdTypes.DefaultUploadAccess)
	initMsg := CallbackContractInstantiateMsg{Count: 100}
	contractAddr, _ := chain.InstantiateContract(contractAdminAcc, codeID, contractAdminAcc.Address.String(), "callback_test", nil, initMsg)

	// Reserving a callback for the very next height
	// This callback will decrement the count
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

	// Increment block height and run end blocker
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	require.Equal(t, ctx.BlockHeight(), reqMsg.CallbackHeight)
	_ = callbackabci.EndBlocker(ctx, keeper, chain.GetApp().Keepers.WASMKeeper)

	// Checking if the count value has been decremented
	count := getCount(t, chain, ctx, contractAddr)
	require.Equal(t, initMsg.Count-1, count)

	// Reserving a callback for next block
	// This callback will increment the count
	reqMsg.JobId = INCREMENT_JOBID
	reqMsg.CallbackHeight = ctx.BlockHeight() + 1
	_, err = msgServer.RequestCallback(sdk.WrapSDKContext(ctx), reqMsg)
	require.NoError(t, err)

	// Increment block height and run end blocker
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	_ = callbackabci.EndBlocker(ctx, keeper, chain.GetApp().Keepers.WASMKeeper)

	// Checking if the count value has been incremented. Should be same as og value now
	count = getCount(t, chain, ctx, contractAddr)
	require.Equal(t, initMsg.Count, count)

	// Reserving a callback for next block
	// This callback will do nothing to the count value
	reqMsg.JobId = DONOTHING_JOBID
	reqMsg.CallbackHeight = ctx.BlockHeight() + 1
	_, err = msgServer.RequestCallback(sdk.WrapSDKContext(ctx), reqMsg)
	require.NoError(t, err)

	// Increment block height and run end blocker
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	_ = callbackabci.EndBlocker(ctx, keeper, chain.GetApp().Keepers.WASMKeeper)

	// Checking if the count value has changed. Should be same as before
	count = getCount(t, chain, ctx, contractAddr)
	require.Equal(t, initMsg.Count, count)

	// To test when callback exceeds gas limit. Setting module params callbackGasLimit to 1.
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

	// Reserving a callback for next block
	// This callback should fail as it consumes more gas than allowed
	reqMsg.JobId = INCREMENT_JOBID
	reqMsg.CallbackHeight = ctx.BlockHeight() + 1
	_, err = msgServer.RequestCallback(sdk.WrapSDKContext(ctx), reqMsg)
	require.NoError(t, err)

	// Increment block height and run end blocker
	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 1)
	_ = callbackabci.EndBlocker(ctx, keeper, chain.GetApp().Keepers.WASMKeeper)

	// Checking if the count value has incremented. Should not have incremented as the callback failed due to out of gas error
	count = getCount(t, chain, ctx, contractAddr)
	require.Equal(t, initMsg.Count, count)
}

// getCount is a helper function to get the contract's count value
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
