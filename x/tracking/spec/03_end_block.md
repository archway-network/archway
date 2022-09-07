<!--
order: 3
-->

# End-Block

Section describes the module state change on the ABCI end block call.

## Finalize Tx Tracking

`TxInfo` objects are compiled and set with their total gas consumed:

  1. Retrieve created `TxInfo` objects for this block.
  2. Retrieve `ContractOperationInfo` objects linked to the `TxInfo` objects.
  3. For each `TxInfo` in the block: 
    - get `contractOp.VmGas`;
    - get `contractOp.SdkGas`;
    - set `TxInfo.TotalGas`;
