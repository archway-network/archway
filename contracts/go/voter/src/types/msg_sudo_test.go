package types

import (
	"testing"

	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	"github.com/stretchr/testify/assert"
)

func TestChangeCostRequestValidate(t *testing.T) {
	type testCase struct {
		name string
		msg  ChangeCostRequest
		//
		errExpected bool
	}

	testCases := []testCase{
		{
			name: "OK",
			msg: ChangeCostRequest{
				NewCost: stdTypes.NewCoinFromUint64(100, "uatom"),
			},
		},
		{
			name: "Fail: NewCost: invalid denom",
			msg: ChangeCostRequest{
				NewCost: stdTypes.NewCoinFromUint64(100, "1uatom"),
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.errExpected {
				assert.Error(t, tc.msg.Validate())
				return
			}
			assert.NoError(t, tc.msg.Validate())
		})
	}
}
