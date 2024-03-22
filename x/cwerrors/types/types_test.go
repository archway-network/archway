package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/cwerrors/types"
)

func TestSudoErrorValidate(t *testing.T) {
	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	type testCase struct {
		name        string
		sudoError   types.SudoError
		errExpected bool
	}

	testCases := []testCase{
		{
			name:        "Fail: Empty values",
			sudoError:   types.SudoError{},
			errExpected: true,
		},
		{
			name: "Fail: Invalid contract address",
			sudoError: types.SudoError{
				ContractAddress: "👻",
				ModuleName:      "test",
				ErrorCode:       1,
				InputPayload:    "test",
				ErrorMessage:    "test",
			},
			errExpected: true,
		},
		{
			name: "Fail: Invalid module name",
			sudoError: types.SudoError{
				ContractAddress: contractAddr.String(),
				ModuleName:      "",
				ErrorCode:       1,
				InputPayload:    "test",
				ErrorMessage:    "test",
			},
			errExpected: true,
		},
		{
			name: "OK: Valid callback",
			sudoError: types.SudoError{
				ContractAddress: contractAddr.String(),
				ModuleName:      "test",
				ErrorCode:       1,
				InputPayload:    "test",
				ErrorMessage:    "test",
			},
			errExpected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.sudoError.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
