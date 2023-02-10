<!--
order: 5
-->

# Events

Section describes the module events.

The module emits the following proto-events:

| Source type | Source name              | Protobuf reference                                                                                                                                                       |
| ----------- | ------------------------ |--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Message     | `MsgSetContractMetadata` | [ContractMetadataSetEvent](../../../proto/archway/rewards/v1beta1/events.proto#L11)                                                                                      |
| Message     | `MsgSetFlatFee`          | [ContractFlatFeeSetEvent](../../../proto/archway/rewards/v1beta1/events.proto#L57)                                                                                       |
| Message     | `MsgWithdrawRewards`     | [RewardsWithdrawEvent](../../../proto/archway/rewards/v1beta1/events.proto#L40)                                                                                          |
| Module      | `BeginBlocker`           | [ContractRewardCalculationEvent](../../../proto/archway/rewards/v1beta1/events.proto#L21)                                                                                |
| Keeper      | `MintBankKeeper`         | [MinConsensusFeeSetEvent](../../../proto/archway/rewards/v1beta1/events.proto#L50)                                                                                       |

