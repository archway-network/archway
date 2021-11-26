# Events

In gastracker's BeginBlocker, we use the `BlockGasTracking` object stored to disburse gas rewards, inflation rewards and smart contract premium to the reward address specified by the contract. Following event is emitted on reward payout:


## RewardDistributionEvent

```json

"type": "gastracker.RewardDistributionEvent",
"attributes": [
  {
    "key": "rewardAddress",
    "value":  "{{sdk.AccAddress of the contract receiving rewards}}"
  },
  {
    "key": "contractRewards",
    "value":  "{{sdk.Coins being received}}"
  },
  {
    "key": "leftoverRewards",
    "value":  "{{sdk.Coins left over from total amount of gas accumulated}}"
  }
]
```
