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

A custom message is appended to the messages by the VM wrapperthat encapsulates the actual CosmWasm VM, this message is interpreted by the gastracker message handler to store smart contract operation info to either add new contract meta data or track gas usage for other calls. 

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
