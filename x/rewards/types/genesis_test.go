package types_test

import (
	"testing"
	"time"

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

	accAddrs, _ := e2eTesting.GenAccounts(2)
	contractAddrs := e2eTesting.GenContractAddresses(2)

	mockTime := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)

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
					{ContractAddress: contractAddrs[0].String(), OwnerAddress: accAddrs[0].String()},
				},
				BlockRewards: []rewardsTypes.BlockRewards{
					{Height: 1},
				},
				TxRewards: []rewardsTypes.TxRewards{
					{TxId: 1, Height: 1},
				},
				MinConsensusFee:     sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.OneInt()),
				RewardsRecordLastId: 1,
				RewardsRecords: []rewardsTypes.RewardsRecord{
					{
						Id:               1,
						RewardsAddress:   accAddrs[0].String(),
						Rewards:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())),
						CalculatedHeight: 1,
						CalculatedTime:   mockTime,
					},
				},
			},
		},
		{
			name: "OK: Flat Fees",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				ContractsMetadata: []rewardsTypes.ContractMetadata{
					{ContractAddress: contractAddrs[0].String(), OwnerAddress: accAddrs[0].String()},
					{ContractAddress: contractAddrs[1].String(), OwnerAddress: accAddrs[1].String()},
				},
				BlockRewards: []rewardsTypes.BlockRewards{
					{Height: 1},
				},
				TxRewards: []rewardsTypes.TxRewards{
					{TxId: 1, Height: 1},
				},
				MinConsensusFee:     sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.OneInt()),
				RewardsRecordLastId: 1,
				RewardsRecords: []rewardsTypes.RewardsRecord{
					{
						Id:               1,
						RewardsAddress:   accAddrs[0].String(),
						Rewards:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())),
						CalculatedHeight: 1,
						CalculatedTime:   mockTime,
					},
				},
				FlatFees: []rewardsTypes.FlatFee{
					{
						ContractAddress: contractAddrs[0].String(),
						FlatFee:         sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt()),
					},
					{
						ContractAddress: contractAddrs[1].String(),
						FlatFee:         sdk.NewCoin("uarch", sdk.NewInt(10)),
					},
				},
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
					{ContractAddress: contractAddrs[0].String()},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid ContractsMetadata: duplicates",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				ContractsMetadata: []rewardsTypes.ContractMetadata{
					{ContractAddress: contractAddrs[0].String(), OwnerAddress: accAddrs[0].String()},
					{ContractAddress: contractAddrs[0].String(), OwnerAddress: accAddrs[0].String()},
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
					{TxId: 1, Height: 1},
					{TxId: 1, Height: 1},
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
		{
			name: "Fail: invalid RewardsRecords",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				RewardsRecords: []rewardsTypes.RewardsRecord{
					{},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid RewardsRecords: duplicates",
			genesisState: rewardsTypes.GenesisState{
				Params:              rewardsTypes.DefaultParams(),
				RewardsRecordLastId: 1,
				RewardsRecords: []rewardsTypes.RewardsRecord{
					{
						Id:               1,
						RewardsAddress:   accAddrs[0].String(),
						Rewards:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())),
						CalculatedHeight: 1,
						CalculatedTime:   mockTime,
					},
					{
						Id:               1,
						RewardsAddress:   accAddrs[0].String(),
						Rewards:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())),
						CalculatedHeight: 1,
						CalculatedTime:   mockTime,
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid RewardsRecord lastID",
			genesisState: rewardsTypes.GenesisState{
				Params:              rewardsTypes.DefaultParams(),
				RewardsRecordLastId: 0,
				RewardsRecords: []rewardsTypes.RewardsRecord{
					{
						Id:               1,
						RewardsAddress:   accAddrs[0].String(),
						Rewards:          sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())),
						CalculatedHeight: 1,
						CalculatedTime:   mockTime,
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid FlatFees: metadata not found for corresponding contract",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				ContractsMetadata: []rewardsTypes.ContractMetadata{
					{ContractAddress: contractAddrs[0].String(), OwnerAddress: accAddrs[0].String()},
				},
				FlatFees: []rewardsTypes.FlatFee{
					{
						ContractAddress: contractAddrs[0].String(),
						FlatFee:         sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt()),
					},
					{
						ContractAddress: contractAddrs[1].String(),
						FlatFee:         sdk.NewCoin("uarch", sdk.NewInt(10)),
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid FlatFees: invalid coin",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				ContractsMetadata: []rewardsTypes.ContractMetadata{
					{ContractAddress: contractAddrs[0].String(), OwnerAddress: accAddrs[0].String()},
				},
				FlatFees: []rewardsTypes.FlatFee{
					{
						ContractAddress: contractAddrs[0].String(),
						FlatFee: sdk.Coin{
							Amount: sdk.NewInt(-1),
							Denom:  sdk.DefaultBondDenom,
						},
					},
				},
			},
			errExpected: true,
		},
		{
			name: "Fail: invalid FlatFees: duplicates",
			genesisState: rewardsTypes.GenesisState{
				Params: rewardsTypes.DefaultParams(),
				ContractsMetadata: []rewardsTypes.ContractMetadata{
					{ContractAddress: contractAddrs[0].String(), OwnerAddress: accAddrs[0].String()},
				},
				FlatFees: []rewardsTypes.FlatFee{
					{
						ContractAddress: contractAddrs[0].String(),
						FlatFee:         sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(1)),
					},
					{
						ContractAddress: contractAddrs[0].String(),
						FlatFee:         sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(2)),
					},
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
