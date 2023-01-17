package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestContractMetadataValidate(t *testing.T) {
	type testCase struct {
		name                string
		meta                rewardsTypes.ContractMetadata
		isGenesisValidation bool
		errExpected         bool
	}

	accAddrs, _ := e2eTesting.GenAccounts(1)
	accAddr := accAddrs[0]

	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	testCases := []testCase{
		{
			name: "OK: empty",
			meta: rewardsTypes.ContractMetadata{
				ContractAddress: contractAddr.String(),
			},
		},
		{
			name: "OK: with OwnerAddress",
			meta: rewardsTypes.ContractMetadata{
				ContractAddress: contractAddr.String(),
				OwnerAddress:    accAddr.String(),
			},
		},
		{
			name: "OK: with RewardsAddress",
			meta: rewardsTypes.ContractMetadata{
				ContractAddress: contractAddr.String(),
				RewardsAddress:  accAddr.String(),
			},
		},
		{
			name: "Fail: invalid ContractAddress",
			meta: rewardsTypes.ContractMetadata{
				ContractAddress: "invalid",
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid RewardsAddress",
			meta: rewardsTypes.ContractMetadata{
				ContractAddress: contractAddr.String(),
				RewardsAddress:  "invalid",
			},
			errExpected: true,
		},
		{
			name: "Fail: empty OwnerAddress with genesis validation",
			meta: rewardsTypes.ContractMetadata{
				ContractAddress: contractAddr.String(),
			},
			isGenesisValidation: true,
			errExpected:         true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.meta.Validate(tc.isGenesisValidation)
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
