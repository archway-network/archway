# Events

During BeginBlock using the information tracked rewards are distributed and the following event is emitted:


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
