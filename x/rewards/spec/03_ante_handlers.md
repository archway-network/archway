<!--
order: 3
-->

# AnteHandlers

Section describes the module ante handlers.

## MinFeeDecorator

The [MinFeeDecorator](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/x/rewards/ante/min_cons_fee.go#L19) checks if a transaction fees are greater or equal to a calculated value.
The handler declines the transaction if the provided fees do not match the condition:
$$
TxFees < TxGasLimit * MinConsensusFee
$$
where:

* *TxFees* - transaction fees provided by a user;
* *TxGasLimit* - transaction gas limit provided by a user;
* *MinConsensusFee* - minimum gas unit price estimated by the module;

## DeductFeeDecorator

The [DeductFeeDecorator](https://github.com/archway-network/archway/blob/e130d74bd456be037b4e60dea7dada5d7a8760b5/x/rewards/ante/fee_deduction.go#L29) handler splits a transaction fees between the **FeeCollector** (`x/auth`) and the **Rewards** (`x/rewards`) modules using the *TxFeeRebateRatio* module parameter.
Handler also creates a new [TxRewards](01_state.md#TxRewards) tracking entry.

