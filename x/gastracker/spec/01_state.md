# State 
The `gastracker` module keeps state of: 
1. **[Contract Instance Metadata](https://github.com/archway-network/archway/blob/c22d753a241672c6974a49063c2f95c5b68bef41/proto/gastracker/types.proto#L78-L81)**
2. **[RewardEntry](https://github.com/archway-network/archway/blob/c22d753a241672c6974a49063c2f95c5b68bef41/proto/gastracker/types.proto#L84-L86)**
3. **[BlockGasTracking](https://github.com/archway-network/archway/blob/500b1e9602714a74c7ba7a1399489605db226b91/proto/gastracker/types.proto#L45)**

We use the following indexes to manage the state:
- Contract Instance Metadata `c_inst_md |  Address -> Protobuffer(address)`
- RewardEntry `reward_entry |  Address -> Protobuffer(address)`

