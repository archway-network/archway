package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/cwerrors/types"
)

// TestSudoErrorValidate tests the json encoding of the sudo callback which is sent to the contract
func TestSudoErrorMsgString(t *testing.T) {
	contractAddr := e2eTesting.GenContractAddresses(1)[0]
	testCases := []struct {
		testCase    string
		msg         types.SudoError
		expectedMsg string
	}{
		{
			"ok",
			types.SudoError{
				ModuleName:      "callback",
				ContractAddress: contractAddr.String(),
				ErrorCode:       1,
				InputPayload:    "hello",
				ErrorMessage:    "world",
			},
			`{"module_name":"callback","error_code":1,"contract_address":"cosmos1w0w8sasnut0jx0vvsnvlc8nayq0q2ej8xgrpwgel05tn6wy4r57q8wwdxx","input_payload":"hello","error_message":"world"}`,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.testCase, func(t *testing.T) {
			res := tc.msg.Bytes()
			require.EqualValues(t, tc.expectedMsg, string(res))
		})
	}
}
