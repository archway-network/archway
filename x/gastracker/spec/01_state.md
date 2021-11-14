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

Whenever a smart contract is instantiated archway stores this address will be the target address for rewards coming for the smart contract

```proto3
message ContractInstanceMetadata {
  string reward_address = 1;
  bool gas_rebate_to_user = 2;
}
```

## Reward Entry
- Reward Entry `reward_entry |  Address -> Protobuffer(address)`

Stores all reward transactions performed for X address, its also used to determine leftover rewards.


```proto3
message LeftOverRewardEntry {
  repeated cosmos.base.v1beta1.DecCoin contract_rewards = 1;
}
```

