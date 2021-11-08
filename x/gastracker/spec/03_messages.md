# Messages

## Contract Operation Info

When the Gas Tracking Engine does one of the following:
- Instantiate
- Execute
- Migrate
- Sudo 
- Reply 
- IBCChannelConnect
- IBCChannelClose
- IBCPacketReceive
- IBCPacketAck
- IBCPacketTimeout

A custom Contract Operation Info is created by the Custom Engine, this allows us to track contracts gas consumptions

```proto
message ContractOperationInfo {
  uint64 gas_consumed = 1;
  ContractOperation operation = 2;
  // Only set in case of instantiate operation
  string reward_address = 3;
  // Only set in case of instantiate operation
  bool gas_rebate_to_end_user = 4;
}
```
