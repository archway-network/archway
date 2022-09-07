package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateDenom(t *testing.T) {
	type testCase struct {
		denom string
		//
		errExpected bool
	}

	testCases := []testCase{
		{
			denom: "uatom",
		},
		{
			denom: "ibc/BE1BB42D4BE3C30D50B68D7C41DB4DFCE9678E8EF8C539F6E6A9345048894FCC",
		},
		{
			denom:       "a",
			errExpected: true,
		},
		{
			denom:       "0uatom",
			errExpected: true,
		},
		{
			denom:       "uatom#",
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.denom, func(t *testing.T) {
			if tc.errExpected {
				assert.Error(t, ValidateDenom(tc.denom))
				return
			}
			assert.NoError(t, ValidateDenom(tc.denom))
		})
	}
}
