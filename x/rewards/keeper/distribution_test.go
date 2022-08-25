package keeper_test

import (
	"testing"

	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintTypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// TestRewardsKeeper_Distribution tests rewards distribution for a single block with different edge cases.
// This is not an E2E test, we emulate x/tracking and x/rewards Ante handler calls to create tracking entries directly.
// Ante handlers are tested independently.
func TestRewardsKeeper_Distribution(t *testing.T) {
	type (
		contractInput struct {
			metadataExists bool           // if true, metadata is set
			contractAddr   sdk.AccAddress // any random address to merge operations [sdk.AccAddr]
			rewardsAddr    string         // might be empty to skip distribution (should be a real chain address) [sdk.AccAddr]
			operations     []uint64       // list of gas consumptions per operation (opType is set randomly)
		}

		transactionInput struct {
			feeCoins  string          // fee coins for this transaction (might be empty to skip distribution) [sdk.Coins]
			contracts []contractInput // list of contracts and their operations
		}

		contractOutput struct {
			rewardsAddr sdk.AccAddress // must be set since we are checking its balance [sdk.AccAddr]
			recordsNum  int            // expected number of rewards records created (0 if none)
			rewards     string         // expected rewards (might be empty if no rewards are expected) [sdk.Coins]
		}

		testCase struct {
			name string
			// inputs
			blockInflationCoin string             // block inflation coin (might be empty to skip distribution) [sdk.Coin]
			blockGasLimit      int64              // consensus parameter (might be 0 to skip inflation distribution)
			txs                []transactionInput // block transactions input
			// expected outputs
			contractsOutput  []contractOutput // list of contracts and their expected rewards (might not include some contracts if they don't have metadata set)
			treasuryExpected string           // rewards leftovers expected
		}
	)

	// Generate empty addresses
	accAddrs, _ := e2eTesting.GenAccounts(10)
	contractAddrs := e2eTesting.GenContractAddresses(10)
	testCases := []testCase{
		{
			name: "No-op",
		},
		{
			name:               "1 tx, 1 contract, 1 op",
			blockInflationCoin: "1000stake",
			blockGasLimit:      1000,
			txs: []transactionInput{
				{
					feeCoins: "500stake",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								100,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					// Tx rewards:  1.0 (100 / 100 tx gas)     = 500stake
					// Inf rewards: 0.1 (100 / 1000 block gas) = 100stake
					rewards:    "600stake",
					recordsNum: 1, // from 1 contract
				},
			},
			// Leftovers:
			//   - Tx:  all distributed
			//   - Inf: 1000stake - 100stake = 900stake
			treasuryExpected: "900stake",
		},
		{
			name:               "1 tx, 1 contract, 3 ops",
			blockInflationCoin: "1000stake",
			blockGasLimit:      1000,
			txs: []transactionInput{
				{
					feeCoins: "500stake",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								100,
								50,
								25,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					// Tx rewards:  1.0   (175 / 175 tx gas)     = 500stake
					// Inf rewards: 0.175 (175 / 1000 block gas) = 175stake
					rewards:    "675stake",
					recordsNum: 1, // from 1 contract
				},
			},
			// Leftovers:
			//   - Tx:  all distributed
			//   - Inf: 1000stake - 175stake = 825stake
			treasuryExpected: "825stake",
		},
		{
			name:               "1 tx, 2 contracts with 2 ops for each",
			blockInflationCoin: "1000stake",
			blockGasLimit:      1000,
			txs: []transactionInput{
				{
					feeCoins: "500stake",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								100,
								50,
							},
						},
						{
							metadataExists: true,
							contractAddr:   contractAddrs[1],
							rewardsAddr:    accAddrs[1].String(),
							operations: []uint64{
								200,
								100,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					// Tx rewards:  ~0.3 (150 / 450 tx gas)     = 166stake
					// Inf rewards: 0.15 (150 / 1000 block gas) = 150stake
					rewards:    "316stake",
					recordsNum: 1, // from 1 contract
				},
				{
					rewardsAddr: accAddrs[1],
					// Tx rewards:  ~0.6 (300 / 450 tx gas)     = 333stake
					// Inf rewards: 0.3  (300 / 1000 block gas) = 300stake
					rewards:    "633stake",
					recordsNum: 1, // from 1 contract
				},
			},
			// Leftovers:
			//   - Tx:  500stake - 166stake - 333stake  = 1stake
			//   - Inf: 1000stake - 150stake - 300stake = 550stake
			treasuryExpected: "551stake",
		},
		{
			name:               "2 txs with contract ops intersection (rewards from both txs)",
			blockInflationCoin: "1000stake",
			blockGasLimit:      1500,
			txs: []transactionInput{
				{
					feeCoins: "500stake",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								200,
								250,
							},
						},
						{
							metadataExists: true,
							contractAddr:   contractAddrs[1],
							rewardsAddr:    accAddrs[1].String(),
							operations: []uint64{
								100,
								200,
								300,
							},
						},
					},
				},
				{
					feeCoins: "600stake",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								10,
							},
						},
						{
							metadataExists: true,
							contractAddr:   contractAddrs[1],
							rewardsAddr:    accAddrs[1].String(),
							operations: []uint64{
								20,
								30,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					// Tx 1 rewards: ~0.43 (450 / 1050 tx gas)    = 214stake
					// Tx 2 rewards: ~0.17 (10 / 60 tx gas)       = 100stake
					// Inf rewards:  ~0.30 (460 / 1500 block gas) = 306stake
					rewards:    "620stake",
					recordsNum: 1, // from 1 contract
				},
				{
					rewardsAddr: accAddrs[1],
					// Tx 1 rewards:  ~0.57 (600 / 1050 tx gas)    = 285stake
					// Tx 2 rewards:  ~0.83 (50 / 60 tx gas)       = 499stake
					// Inf rewards:   ~0.43 (650 / 1500 block gas) = 433stake
					rewards:    "1217stake",
					recordsNum: 1, // from 1 contract
				},
			},
			// Leftovers:
			//   - Tx 1: 500stake - 214stake - 285stake  = 1stake
			//   - Tx 2: 600stake - 100stake - 499stake  = 1stake
			//   - Inf:  1000stake - 306stake - 433stake = 261stake
			treasuryExpected: "263stake",
		},
		{
			name:               "1 tx with 2 contracts (one without metadata)",
			blockInflationCoin: "1000stake",
			blockGasLimit:      1000,
			txs: []transactionInput{
				{
					feeCoins: "500stake",
					contracts: []contractInput{
						{
							metadataExists: false,
							contractAddr:   contractAddrs[0],
							operations: []uint64{
								100,
							},
						},
						{
							metadataExists: true,
							contractAddr:   contractAddrs[1],
							rewardsAddr:    accAddrs[1].String(),
							operations: []uint64{
								100,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					rewards:     "",
				},
				{
					rewardsAddr: accAddrs[1],
					// Tx rewards:  0.5 (100 / 200 tx gas)     = 250stake
					// Inf rewards: 0.1 (100 / 1000 block gas) = 100stake
					rewards:    "350stake",
					recordsNum: 1, // from 1 contract
				},
			},
			// Leftovers:
			//   - Tx:  500stake - 250stake  = 250stake
			//   - Inf: 1000stake - 100stake = 900stake
			treasuryExpected: "1150stake",
		},
		{
			name:               "1 tx with 2 contracts (one without rewardsAddress)",
			blockInflationCoin: "1000stake",
			blockGasLimit:      1000,
			txs: []transactionInput{
				{
					feeCoins: "500stake",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							operations: []uint64{
								100,
							},
						},
						{
							metadataExists: true,
							contractAddr:   contractAddrs[1],
							rewardsAddr:    accAddrs[1].String(),
							operations: []uint64{
								100,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					rewards:     "",
				},
				{
					rewardsAddr: accAddrs[1],
					// Tx rewards:  0.5 (100 / 200 tx gas)     = 250stake
					// Inf rewards: 0.1 (100 / 1000 block gas) = 100stake
					rewards:    "350stake",
					recordsNum: 1, // from 1 contract
				},
			},
			// Leftovers:
			//   - Tx:  500stake - 250stake  = 250stake
			//   - Inf: 1000stake - 100stake = 900stake
			treasuryExpected: "1150stake",
		},
		{
			name:               "1 tx, 1 contract, 1 op (no tx fees)",
			blockInflationCoin: "1000stake",
			blockGasLimit:      1000,
			txs: []transactionInput{
				{
					feeCoins: "",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								100,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					// Inf rewards: 0.1 (100 / 1000 block gas) = 100stake
					rewards:    "100stake",
					recordsNum: 1, // from 1 contract
				},
			},
			// Leftovers:
			//   - Tx:  none
			//   - Inf: 1000stake - 100stake = 900stake
			treasuryExpected: "900stake",
		},
		{
			name:               "1 tx, 1 contract, 1 op (no inflation)",
			blockInflationCoin: "",
			blockGasLimit:      1000,
			txs: []transactionInput{
				{
					feeCoins: "500stake",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								100,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					// Tx rewards: 1.0 (100 / 100 tx gas) = 500stake
					rewards:    "500stake",
					recordsNum: 1, // from 1 contract
				},
			},
			// Leftovers:
			//   - Tx:  none
			//   - Inf: none
			treasuryExpected: "",
		},
		{
			name:               "1 tx, 1 contract, 1 op (no block gas limit)",
			blockInflationCoin: "1000stake",
			blockGasLimit:      -1,
			txs: []transactionInput{
				{
					feeCoins: "500stake",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								100,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					// Tx rewards: 1.0 (100 / 100 tx gas) = 500stake
					rewards:    "500stake",
					recordsNum: 1, // from 1 contract
				},
			},
			// Leftovers:
			//   - Tx:  none
			//   - Inf: 1000stake (not distributed at all)
			treasuryExpected: "1000stake",
		},
		{
			name:               "1 tx, 1 contract, 1 op (no tx fee, no inflation)",
			blockInflationCoin: "",
			blockGasLimit:      1000,
			txs: []transactionInput{
				{
					feeCoins: "",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								100,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					rewards:     "",
				},
			},
			// Leftovers:
			//   - Tx:  none
			//   - Inf: none
			treasuryExpected: "",
		},
		{
			name:               "1 tx, 2 contracts with the same rewardsAddress (multiple records created)",
			blockInflationCoin: "1000stake",
			blockGasLimit:      1000,
			txs: []transactionInput{
				{
					feeCoins: "900stake",
					contracts: []contractInput{
						{
							metadataExists: true,
							contractAddr:   contractAddrs[0],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								100,
							},
						},
						{
							metadataExists: true,
							contractAddr:   contractAddrs[1],
							rewardsAddr:    accAddrs[0].String(),
							operations: []uint64{
								200,
							},
						},
					},
				},
			},
			contractsOutput: []contractOutput{
				{
					rewardsAddr: accAddrs[0],
					// Tx rewards 1st contract:  ~0.33 (100 / 300 tx gas)     = 299stake
					// Inf rewards 1st contract:  0.1  (100 / 1000 block gas) = 100stake
					// Tx rewards 2nd contract:  ~0.66 (200 / 300 tx gas)     = 600stake
					// Inf rewards 2nd contract:  0.2  (200 / 1000 block gas) = 200stake
					rewards:    "1199stake",
					recordsNum: 2, // from 2 contracts
				},
			},
			// Leftovers:
			//   - Tx:  900stake - 299stake - 600stake  = 1stake
			//   - Inf: 1000stake - 100stake - 200stake = 700stake
			treasuryExpected: "701stake",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create chain with block gas limit
			chain := e2eTesting.NewTestChain(t, 1,
				e2eTesting.WithBlockGasLimit(tc.blockGasLimit),
			)
			acc := chain.GetAccount(0)

			// Set mock ContractViewer (to pass contract admin check for metadata setup)
			contractViewer := testutils.NewMockContractViewer()
			chain.GetApp().RewardsKeeper.SetContractInfoViewer(contractViewer)

			tKeeper, rKeeper := chain.GetApp().TrackingKeeper, chain.GetApp().RewardsKeeper
			ctx := chain.GetContext()

			// Setup
			{
				// Create transactions gas tracking and rewards tracking data for the current block
				for _, tx := range tc.txs {
					// Emulate x/tracking AnteHandler call
					tKeeper.TrackNewTx(ctx)

					// Contracts setup
					for _, contract := range tx.contracts {
						// Ingest gas tracking for each contract
						var gasConsumptionRecords []wasmdTypes.ContractGasRecord
						for _, op := range contract.operations {
							gasConsumptionRecord := wasmdTypes.ContractGasRecord{
								OperationId:     testutils.GetRandomContractOperationType(),
								ContractAddress: contract.contractAddr.String(),
								OriginalGas: wasmdTypes.GasConsumptionInfo{
									SDKGas: op,
									VMGas:  0, // to simplify testCase inputs, we don't use VMGas
								},
							}
							gasConsumptionRecords = append(gasConsumptionRecords, gasConsumptionRecord)
						}
						require.NoError(t, tKeeper.IngestGasRecord(ctx, gasConsumptionRecords))

						// Set metadata for the contract
						if contract.metadataExists {
							contractViewer.AddContractAdmin(contract.contractAddr.String(), acc.Address.String())

							metadata := rewardsTypes.ContractMetadata{
								OwnerAddress:   acc.Address.String(),
								RewardsAddress: contract.rewardsAddr,
							}

							require.NoError(t, rKeeper.SetContractMetadata(ctx, acc.Address, contract.contractAddr, metadata))
						}
					}

					// Track fee rewards
					if tx.feeCoins != "" {
						feeRewards, err := sdk.ParseCoinsNormalized(tx.feeCoins)
						require.NoError(t, err)

						// Emulate x/rewards AnteHandler call
						rKeeper.TrackFeeRebatesRewards(ctx, feeRewards)
						// Mint and transfer
						require.NoError(t, chain.GetApp().BankKeeper.MintCoins(ctx, mintTypes.ModuleName, feeRewards))
						require.NoError(t, chain.GetApp().BankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, rewardsTypes.ContractRewardCollector, feeRewards))
					}
				}

				// Burn inflation rewards for the current block caused by the x/mint (we override this value below)
				{
					curBlockRewards, found := rKeeper.GetState().BlockRewardsState(ctx).GetBlockRewards(ctx.BlockHeight())
					require.True(t, found)
					rewardsToBurn := sdk.Coins{curBlockRewards.InflationRewards}

					require.NoError(t, chain.GetApp().BankKeeper.SendCoinsFromModuleToModule(ctx, rewardsTypes.ContractRewardCollector, rewardsTypes.TreasuryCollector, rewardsToBurn))
					require.NoError(t, chain.GetApp().BankKeeper.BurnCoins(ctx, rewardsTypes.TreasuryCollector, rewardsToBurn))
				}

				// Track inflation rewards
				if tc.blockInflationCoin != "" {
					inflationReward, err := sdk.ParseCoinNormalized(tc.blockInflationCoin)
					require.NoError(t, err)
					inflationRewards := sdk.NewCoins(inflationReward)

					// Emulate x/rewards MintKeeper call
					rKeeper.TrackInflationRewards(ctx, inflationReward)
					// Mint and transfer
					require.NoError(t, chain.GetApp().BankKeeper.MintCoins(ctx, mintTypes.ModuleName, inflationRewards))
					require.NoError(t, chain.GetApp().BankKeeper.SendCoinsFromModuleToModule(ctx, mintTypes.ModuleName, rewardsTypes.ContractRewardCollector, inflationRewards))
				} else {
					// We have to remove it since it was created by the x/mint
					rKeeper.GetState().BlockRewardsState(ctx).DeleteBlockRewards(ctx.BlockHeight())
				}
			}

			// Call EndBlocker to finalize x/tracking entries and distribute rewards via x/rewards
			// Clean up treasury balance to assert the module balance later
			{
				// Burn all the treasury collected for previous blocks
				{
					treasuryBalanceInitial := chain.GetModuleBalance(rewardsTypes.TreasuryCollector)
					require.NoError(t, chain.GetApp().BankKeeper.BurnCoins(chain.GetContext(), rewardsTypes.TreasuryCollector, treasuryBalanceInitial))
				}

				chain.NextBlock(0)
			}

			// Assert expectations
			for i, outExpected := range tc.contractsOutput {
				totalRewardsExpected, err := sdk.ParseCoinsNormalized(outExpected.rewards)
				require.NoError(t, err)

				// Check the number of records created
				recordsCreated := chain.GetApp().RewardsKeeper.GetState().RewardsRecord(chain.GetContext()).GetRewardsRecordByRewardsAddress(outExpected.rewardsAddr)
				require.Len(t, recordsCreated, outExpected.recordsNum)

				// Basic check of records and merge total rewards
				totalRewardsReceived := sdk.NewCoins()
				for _, record := range recordsCreated {
					require.NotEmpty(t, record.Id, "output [%d]", i)
					require.NotEmpty(t, record.CalculatedHeight, "output [%d]", i)
					require.NotEmpty(t, record.CalculatedTime, "output [%d]", i)
					require.Equal(t, outExpected.rewardsAddr.String(), record.RewardsAddress, "output [%d]", i)

					totalRewardsReceived = totalRewardsReceived.Add(record.Rewards...)
				}

				assert.Equal(t, totalRewardsExpected.String(), totalRewardsReceived.String(), "output [%d]", i)
			}

			// Assert rewards leftovers
			{
				treasuryPoolExpected, err := sdk.ParseCoinsNormalized(tc.treasuryExpected)
				require.NoError(t, err)
				treasuryPoolReceived := chain.GetApp().RewardsKeeper.TreasuryPool(chain.GetContext())
				assert.Equal(t, treasuryPoolExpected.String(), treasuryPoolReceived.String(), "treasury pool")
			}

			// Skip a block to ensure invariants are not violated
			chain.NextBlock(0)
		})
	}
}
