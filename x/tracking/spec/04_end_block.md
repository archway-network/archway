<!--
order: 3
-->

# End-Block

Section describes the module state change on the ABCI end block call.

## Finalize Tx Tracking
Tx Infos are compiled and set with their total gas consumed

  1. Retrieve tx info state 
  2. Retrieve Contract Operation state
  3. for each tx info in the block, 
    - get contractOp.VmGas
    - contractOp.SdkGas
    - set TxInfo.TotalGas

