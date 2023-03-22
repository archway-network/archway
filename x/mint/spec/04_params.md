<!--
order: 4
-->

# Parameters

Section describes the module parameters.

Parameters available:

| Key                   | Type                    | Default value                                | Allowed values        | Description                                                                                                |
| --------------------- | ----------------------- | -------------------------------------------- | --------------------- | ---------------------------------------------------------------------------------------------------------- |
| MinInflation          | `sdk.Dec`               | "0.00"                                       | [ 0.0 : 1.0 )         | The network's minimum inflation                                                                            |
| MaxInflation          | `sdk.Dec`               | "1.00"                                       | [ 0.0 : 1.0 )         | The network's maximum inflation                                                                            |
| MinBonded             | `sdk.Dec`               | "0.00"                                       | [ 0.0 : 1.0 )         | The minimum wanted bond ratio (staked supply/total supply)                                                 |
| MaxBonded             | `sdk.Dec`               | "1.00"                                       | [ 0.0 : 1.0 )         | The maximum wanted bond ratio (staked supply/total supply)                                                 |
| InflationChange       | `sdk.Dec`               | "1.00"                                       | [ 0.0 : 1.0 )         | How much the inflation should change if the bond ratio is not between the defined bands of min/max_bonded. |
| MaxBlockDuration      | `time.Duration`         | "60"                                         | Any positive duration | The maximum duration of a block                                                                            |
| InflationRecipients   | `[]InflationRecipient`  | `{recipient: "fee_collector", ratio:"1.00"}` | Any module account    | The list of inflation recipients                                                                           |