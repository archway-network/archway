# End Block

Section describes the module state changes on the ABCI end block call

## SudoError invoke

All the errors encountered in the current block are fetched from the transient store. For each error, a contract is hit at the sudo entrypoint. The execution happens with gas limit of `150_000` to prevent abusive operations and limit the usage to error handling.

In case, the execution fails, the error is stored in state such that the contract can query it.

## Prune expiring subscription

All the contract subscriptions which end in the current block are pruned from state

## Prune old errors

All errors which are marked for deletion for the current block height are pruned from the state.
