# Client

Section describes interaction with the module by the user

## CLI

### Query

The `query` commands alllows a user to query the module state

Use the `-h`/`--help` flag to get a help description of a command.

`archwayd q cwica -h`

> You can add the `-o json` for the JSON output format

#### params

Get the current module parameters

Usage:

`archwayd q cwica params [flags]`

Example output:

```yaml
msg_submit_tx_max_messages: "5"
```

#### interchain-account

Gets the interchain account associated with given owner-address, connection-id and interchain-account-id

Usage:

`archwayd q cwica interchain-account [owner-address] [connection-id] [interchain-account-id]`

Example:

`archwayd q cwica interchain-account archway1t0fchjcgpj9zr07guy6u42ph3p4e5ypzz4uhlv connection-1 test2`

Example output:

```yaml
interchain_account_address: "cosmos1c4k24jzduc365kywrsvf5ujz4ya6mwymy8vq4q"
```