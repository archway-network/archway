# State 
The `gastracker` module keeps state of: 
1. **[Contract Instance Metadata](https://github.com/archway-network/archway/blob/c22d753a241672c6974a49063c2f95c5b68bef41/proto/gastracker/types.proto#L78-L81)**
2. **[RewardEntry](https://github.com/archway-network/archway/blob/c22d753a241672c6974a49063c2f95c5b68bef41/proto/gastracker/types.proto#L84-L86)**

We use the following indexes to manage the state:
- Contract Instance Metadata `c_inst_md |  Address -> Protobuffer(address)`
- RewardEntry `reward_entry |  Address -> Protobuffer(address)`

