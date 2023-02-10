<!--
order: 3
-->

# AnteHandlers

Section describes the module ante handlers.

## MinFeeDecorator

The [MinFeeDecorator](../ante/min_cons_fee.go#L19) checks if a transaction fees are greater or equal to a calculated value.
The handler declines the transaction if the provided fees do not match the condition:

$$
TxFees < (TxGasLimit * MinConsensusFee) + \sum_{msg=1, type_{msg} = MsgExecuteContract}^{len(msgs)} flatfee(ContractAddress_{msg})
$$

where:

* $TxFees$ - transaction fees provided by a user;
* $TxGasLimit$ - transaction gas limit provided by a user;
* $MinConsensusFee$ - minimum gas unit price estimated by the module;
* $ContractAddress_{msg}$ - contract address of the msg which needs to be executed;
* $flatfee(x)$ - function which fetches the flat fee for the given input;

Every msg in the transaction is parsed to check if it is a `wasmTypes.MsgExecuteContract` or a `authz.MsgExec` msg. Contract address is identified for matching msgs and `flat_fee` (if set) is fetched for the given contract addresses.


## DeductFeeDecorator

The [DeductFeeDecorator](../ante/fee_deduction.go#L29) handler splits a transaction fees between the **FeeCollector** (`x/auth`) and the **Rewards** (`x/rewards`) modules using the *TxFeeRebateRatio* module parameter.
Handler also creates a new [TxRewards](01_state.md#TxRewards) tracking entry.

