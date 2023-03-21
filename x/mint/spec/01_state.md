<!--
order: 1
-->

# State

Section describes all stored by the module objects and their storage keys.

Refer to the [mint.proto](../../../proto/archway/mint/v1/mint.proto) for objects fields description.

## Params

- Params: `Paramsspace("mint") -> legacy_amino(params)`

[Params](../../../proto/archway/mint/v1/mint.proto#L11) is a module-wide configuration structure.

## LastBlockInfo

[LastBlockInfo](../../../proto/archway/mint/v1/mint.proto#L46) is a singleton object used to store the timestamp of the last inflation minting and the inflation percentage at that point.

Storage keys:

- LastBlockInfo: `0x00 -> ProtocolBuffer(LastBlockInfo)`

## MintDistribution

At every block the inflation distributed to each inflation recipient is stored. The value is stored in sdk.Coin

Storage keys:

- MintDistribution: `0x01 | blockHeight | inflationRecipient -> ProtocolBuffer(sdk.Coin)`
