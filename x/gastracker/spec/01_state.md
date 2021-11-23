# State 

## Current Block Tracking
- CurrentBlockTrackingKey `currnt_blk |  BlockGasTracking -> Protobuffer(BlockGasTracking)`

BlockGasTracking is represented by a custom protobuffer message 
```proto3
message BlockGasTracking { repeated TransactionTracking tx_tracking_infos = 1; }

message TransactionTracking {
  uint64 max_gas_allowed = 1;
  repeated cosmos.base.v1beta1.DecCoin max_contract_rewards = 3;
  repeated ContractGasTracking contract_tracking_infos = 4;
}
```

When a new block starts the protocol retrieves previously stored Block iterates over all Transactions tracked on the block and disburse rewards to smart contracts.

## Contract Instance Metadata
- Contract Instance Metadata `c_inst_md |  Address -> Protobuffer(address)`

Whenever a smart contract is instantiated archway stores this address, which will be the deposit address for rewards accumulated by that smart contract.

```proto3
message ContractInstanceMetadata {
  string reward_address = 1;
  bool gas_rebate_to_user = 2;
  // Flag to indicate whether to charge premium or not
  bool collect_premium = 3;
  // Percentage of gas consumed to be charged.
  uint64 premium_percentage_charged = 4;
}

## Reward Entry
- Reward Entry `reward_entry |  Address -> Protobuffer(address)`

Stores left over reward for particular address.


```proto3
message LeftOverRewardEntry {
  repeated cosmos.base.v1beta1.DecCoin contract_rewards = 1;
}
```

