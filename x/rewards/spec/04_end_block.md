<!--
order: 4
-->

# End-Block

Section describes the module state change on the ABCI end block call.

## Rewards calculation

dApp rewards are calculated as follows:

1. Estimate gas usage by contracts

   * Query all the `x/tracking` module tracking data for the current block (contracts' CosmWasm operations and block transactions gas usage).
   * Query all the `x/rewards` module tracking data for the current block (block inflationary rewards and tx fee rebate rewards).
   * Query a contract metadata.
   * Aggregate all contract operations gas usage into a single value: total gas used by a contract within a specific transaction, total gas used by a contract within a block.

2. Estimate contract rewards

   * Transactions fee rebate rewards for a contract (sum of all block transaction fee rewards contract had operations in):
     
     $$\displaylines{
     TxRewardsShare_i = \frac{ContractTxGasUsed_i}{TxGasUsed} \\
     TxRewards_i = TxFees * TxRewardsShare_i \\
     ContractRewards = \sum_{i=1}^n TxRewards_i
     }$$

   * Block inflation rewards for a contract:
     
     $$\displaylines{
     InflationShare = \frac{ContractGasUsed}{BlockGasLimit} \\
     ContractRewards = BlockRewards * InflationShare
     }$$

3. Create reward records

   * Create a new `RewardsRecord` for a contract if:
     * A contract metadata is set;
     * The `rewards_address` metadata field is set;
   * Multiple `RewardsRecords` could be created for a single rewards address if that address is used by multiple contract metadata.

4. Cleanup

   * Remove `x/tracking` and `x/rewards` tracking entries for the `(currentHeight - 10)` block height;
   * Transfer all the undistributed rewards to the `Treasury` account:

     $$\displaylines{
     TreasuryTokens_i = BlockRewardsTotal_i - BlockRewardsDistributed_i
     }$$
     
     where:
     * *BlockRewardsTotal* - total rewards tracked for the block (inflationary rewards + transaction fee rewards);
     * *BlockRewardsDistributed* - rewards distributed to contracts' `rewards_address`;
