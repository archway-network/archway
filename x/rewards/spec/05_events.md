<!--
order: 5
-->

# Events

Section describes the module events.

The module emits the following proto events:

| Source type | Source name              | Protobuf reference                                           |
| ----------- | ------------------------ | ------------------------------------------------------------ |
| Message     | `MsgSetContractMetadata` | [ContractMetadataSetEvent](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/proto/archway/rewards/v1beta1/events.proto#L11) |
| Message     | `MsgWithdrawRewards`     | [RewardsWithdrawEvent](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/proto/archway/rewards/v1beta1/events.proto#L40) |
| Module      | `BeginBlocker`           | [ContractRewardCalculationEvent](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/proto/archway/rewards/v1beta1/events.proto#L21) |
| Keeper      | `MintBankKeeper`         | [MinConsensusFeeSetEvent](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/proto/archway/rewards/v1beta1/events.proto#L50) |

