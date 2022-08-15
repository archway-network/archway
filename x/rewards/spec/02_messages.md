<!--
order: 2
-->

# Messages

Section describes the processing of the module messages.

## MsgSetContractMetadata

A contract metadata is created / updated using the [MsgSetContractMetadata](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/proto/archway/rewards/v1beta1/tx.proto#L22) message.

On success:

- Metadata's `owner_address` / `rewards_address` is set / updated;

This message is expected to fail if:

* A corresponding contract is not found (not *Instantiated*);
* Metadata does not exist: the message sender is not the contract admin (CowmWasm *Instantiate* option);
* Metadata exists: the message sender is not the `owner_address` (metadata field);

Metadata can also be updated by a contract (refer to WASM bindings section).

## MsgWithdrawRewards

Contract(s) rewards are withdrawn using the [MsgWithdrawRewards](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/proto/archway/rewards/v1beta1/tx.proto#L36) message.

On success:

* Rewards address receives rewards tokens;

Returns:

* The message [response](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/proto/archway/rewards/v1beta1/tx.proto#L42) contains the total amount of rewards tokens transferred (empty if this rewards address has no rewards yet);

This *withdraw* operation can also be triggered by a contract (refer to WASM bindings section).

