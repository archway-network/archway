package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

func TestRewardsGenesisStateValidate(t *testing.T) {
	type testCase struct {
		name         string
		genesisState rewardsTypes.GenesisState
		errExpected  bool
	}

	accAddrs, _ := e2eTesting.GenAccounts(1)
	accAddr := accAddrs[0]

	contractAddr := e2eTesting.GenContractAddresses(1)[0]

	testCases := []testCase{
		{
			name: "OK: empty",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
			},
		},
		{
			name: "OK",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				ContractsMetadata: []rewardsTypes.ContractMetadata{
					{ContractAddress: contractAddr.String(), OwnerAddress: accAddr.String()},
				},
				BlockRewards: []rewardsTypes.BlockRewards{
					{Height: 1},
				},
				TxRewards: []rewardsTypes.TxRewards{
					{TxId: 1, Height: 1},
				},
				MinConsensusFee: sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.OneInt()),
			},
		},
		{
			name: "Fail: invalid Params",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.Params{
					InflationRewardsRatio: sdk.NewDecWithPrec(15, 0),
					TxFeeRebateRatio:      sdk.NewDecWithPrec(5, 2),
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid ContractsMetadata: ownerAddress not set",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				ContractsMetadata: []rewardsTypes.ContractMetadata{
					{ContractAddress: contractAddr.String()},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid ContractsMetadata: duplicates",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				ContractsMetadata: []rewardsTypes.ContractMetadata{
					{ContractAddress: contractAddr.String(), OwnerAddress: accAddr.String()},
					{ContractAddress: contractAddr.String(), OwnerAddress: accAddr.String()},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid BlockRewards",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				BlockRewards: []rewardsTypes.BlockRewards{
					{Height: -1},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid BlockRewards: duplicates",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				BlockRewards: []rewardsTypes.BlockRewards{
					{Height: 1},
					{Height: 1},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid TxRewards",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				TxRewards: []rewardsTypes.TxRewards{
					{TxId: 0},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid TxRewards: duplicates",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				BlockRewards: []rewardsTypes.BlockRewards{
					{Height: 1},
				},
				TxRewards: []rewardsTypes.TxRewards{
					{TxId: 1},
					{TxId: 1},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid TxRewards: non-existing block rewards",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				TxRewards: []rewardsTypes.TxRewards{
					{TxId: 1, Height: 1},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid MinConsensusFee",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				MinConsensusFee: sdk.DecCoin{
					Denom:  sdk.DefaultBondDenom,
					Amount: sdk.NewDec(-1),
				},
			},
			errExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.genesisState.Validate()
			if tc.errExpected {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
