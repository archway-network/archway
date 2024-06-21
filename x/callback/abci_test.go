package callback_test

import (
	"encoding/json"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/callback/types"
	cwerrortypes "github.com/archway-network/archway/x/cwerrors/types"
)

const (
	DECREMENT_JOBID = 0
	INCREMENT_JOBID = 1
	ERROR_JOBID     = 2
	DONOTHING_JOBID = 3
)

func TestEndBlocker(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	errorsKeeper := chain.GetApp().Keepers.CWErrorsKeeper
	contractAdminAcc := chain.GetAccount(0)

	// Upload and instantiate contract
	// The test contract is based on the default counter contract and behaves the following way:
	// When job_id = 1, it increments the count value
	// When job_id = 0, it decrements the count value
	// When job_id = 2, it throws an error
	// For any other job_id, it does nothing
	codeID := chain.UploadContract(contractAdminAcc, "../../contracts/callback-test/artifacts/callback_test.wasm", wasmdTypes.DefaultUploadAccess)
	initMsg := CallbackContractInstantiateMsg{Count: 100}
	contractAddr, _ := chain.InstantiateContract(contractAdminAcc, codeID, contractAdminAcc.Address.String(), "callback_test", nil, initMsg)

	testCases := []struct {
		testCase      string
		jobId         uint64
		expectedCount int32
	}{
		{
			testCase:      "Decrement count",
			jobId:         DECREMENT_JOBID,
			expectedCount: initMsg.Count - 1,
		},
		{
			testCase:      "Increment count",
			jobId:         INCREMENT_JOBID,
			expectedCount: initMsg.Count,
		},
		{
			testCase:      "Do nothing",
			jobId:         DONOTHING_JOBID,
			expectedCount: initMsg.Count,
		},
		{
			testCase:      "Throw error", // The contract throws error but the EndBlocker should not.
			jobId:         ERROR_JOBID,
			expectedCount: initMsg.Count,
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Case: %s", tc.testCase), func(t *testing.T) {
			ctx := chain.GetContext()
			feesToPay, err := getCallbackRegistrationFees(chain)
			require.NoError(t, err)

			reqMsg := types.MsgRequestCallback{
				ContractAddress: contractAddr.String(),
				JobId:           tc.jobId,
				CallbackHeight:  ctx.BlockHeight() + 2,
				Sender:          contractAdminAcc.Address.String(),
				Fees:            feesToPay,
			}
			_, _, _, err = chain.SendMsgs(contractAdminAcc, true, []sdk.Msg{&reqMsg})
			require.NoError(t, err)
			//Increment block height
			chain.NextBlock(1)
			chain.NextBlock(1)

			// Checking if the count value is as expected
			count := getCount(t, chain, contractAddr)
			require.Equal(t, tc.expectedCount, count)
		})
	}

	// Ensure error is captured by the cwerrors module - the case is when job id = 2
	sudoErrs, err := errorsKeeper.GetErrorsByContractAddress(chain.GetContext(), contractAddr)
	require.NoError(t, err)
	require.Len(t, sudoErrs, 1)
	require.Equal(t, "SomeError: execute wasm contract failed", sudoErrs[0].ErrorMessage)
	require.Equal(t, types.ModuleName, sudoErrs[0].ModuleName)
	require.Equal(t, int32(types.ModuleErrors_ERR_CONTRACT_EXECUTION_FAILED), sudoErrs[0].ErrorCode)
}

func TestEndBlockerWithCallbackGasLimit(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1,
		e2eTesting.WithCallbackParams(1),
	)
	errorsKeeper := chain.GetApp().Keepers.CWErrorsKeeper
	contractAdminAcc := chain.GetAccount(0)

	// Upload and instantiate contract
	// The test contract is based on the default counter contract and behaves the following way:
	// When job_id = 1, it increments the count value
	codeID := chain.UploadContract(contractAdminAcc, "../../contracts/callback-test/artifacts/callback_test.wasm", wasmdTypes.DefaultUploadAccess)
	initMsg := CallbackContractInstantiateMsg{Count: 100}
	contractAddr, _ := chain.InstantiateContract(contractAdminAcc, codeID, contractAdminAcc.Address.String(), "callback_test", nil, initMsg)

	// Subscribing to errors module. The contract does not implement the Sudo::Error entrypoint.
	subReqMsg := &cwerrortypes.MsgSubscribeToError{
		Sender:          contractAdminAcc.Address.String(),
		ContractAddress: contractAddr.String(),
		Fee:             sdk.NewInt64Coin(sdk.DefaultBondDenom, 0),
	}
	_, _, _, err := chain.SendMsgs(contractAdminAcc, true, []sdk.Msg{subReqMsg})
	require.NoError(t, err)
	require.True(t, errorsKeeper.HasSubscription(chain.GetContext(), contractAddr))

	// This callback should fail as it consumes more gas than allowed
	feesToPay, err := getCallbackRegistrationFees(chain)
	require.NoError(t, err)
	reqMsg := &types.MsgRequestCallback{
		ContractAddress: contractAddr.String(),
		JobId:           INCREMENT_JOBID,
		CallbackHeight:  chain.GetContext().BlockHeight() + 2,
		Sender:          contractAdminAcc.Address.String(),
		Fees:            feesToPay,
	}
	_, _, _, err = chain.SendMsgs(contractAdminAcc, true, []sdk.Msg{reqMsg})
	require.NoError(t, err)

	// Increment block height
	chain.NextBlock(1)
	chain.NextBlock(1)

	// Checking if the count value has incremented. Should not have incremented as the callback failed due to out of gas error
	count := getCount(t, chain, contractAddr)
	require.Equal(t, initMsg.Count, count)

	sudoErrs, err := errorsKeeper.GetErrorsByContractAddress(chain.GetContext(), contractAddr)
	require.NoError(t, err)
	require.Len(t, sudoErrs, 1)
	require.Equal(t, cwerrortypes.ModuleName, sudoErrs[0].ModuleName) // because Sudo::Error entrypoint does not exist on the contract
	require.Equal(t, int32(cwerrortypes.ModuleErrors_ERR_CALLBACK_EXECUTION_FAILED), sudoErrs[0].ErrorCode)
}

func getCallbackRegistrationFees(chain *e2eTesting.TestChain) (sdk.Coin, error) {
	ctx := chain.GetContext()
	currentBlockHeight := ctx.BlockHeight()
	callbackHeight := currentBlockHeight + 2
	futureResFee, blockResFee, txFee, err := chain.GetApp().Keepers.CallbackKeeper.EstimateCallbackFees(ctx, callbackHeight)
	if err != nil {
		return sdk.Coin{}, err
	}
	feesToPay := futureResFee.Add(blockResFee).Add(txFee)
	return feesToPay.Add(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000)), nil
}

// getCount is a helper function to get the contract's count value
func getCount(t *testing.T, chain *e2eTesting.TestChain, contractAddr sdk.AccAddress) int32 {
	getCountQuery := "{\"get_count\":{}}"
	resp, err := chain.GetApp().Keepers.WASMKeeper.QuerySmart(chain.GetContext(), contractAddr, []byte(getCountQuery))
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
