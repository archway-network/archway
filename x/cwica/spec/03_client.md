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

#### interchain account

To fetch the interchain-account addresses, use the interchain accounts module.

Usage: 

`archwayd query interchain-accounts controller interchain-account <smart_contract_address> <ibc_connection_id>`

Example input:

`archwayd query interchain-accounts controller interchain-account archway1zlc00gjw4ecan3tkk5g0lfd78gyfldh4hvkv2g8z5qnwlkz9vqmsdfvs7q connection-1`

Example output:

```yaml
address: "cosmos1layxcsmyye0dc0har9sdfzwckaz8sjwlfsj8zs"
```