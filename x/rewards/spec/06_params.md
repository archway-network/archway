<!--
order: 6
-->

# Parameters

Section describes the module parameters.

Parameters available:

| Key                   | Type      | Default value | Allowed values | Description                                                  |
| --------------------- | --------- | ------------- | -------------- | ------------------------------------------------------------ |
| TxFeeRebateRatio      | `sdk.Dec` | "0.50"        | [ 0.0 : 1.0 )  | Ratio to split transaction fee rewards between dApps and Validators / Delegators |
| InflationRewardsRatio | `sdk.Dec` | "0.20"        | [ 0.0 : 1.0 )  | Ratio to split minted inflation rewards between dApps and Validators / Delegators |
| MaxWithdrawRecords    | `uint64`  | 25000         | GT 0           | The maximum number of `RewardsRecord` entries to process by the *withdrawal* operation or to query via WASM bindings. |

