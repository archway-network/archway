<!--
order: 5
-->

# Client

Section describes interaction with the module by a user.

## CLI

### Query

The `query` commands allows a user to query the module state.

Use the `-h` / `--help` flag to get a help description of a command.

```bash
archwayd q mint -h
```

> You can add the `-o json` for the JSON output format.

#### params

Get the current module parameters.

Usage:

```bash
archwayd q mint params [flags]
```

Example output:

```yaml
min_inflation: "0.200000000000000000"
max_inflation: "0.500000000000000000"
min_bonded: "0.500000000000000000"
max_bonded: "0.500000000000000000"
inflation_change: "0.500000000000000000"
max_block_duration: 60s
inflation_recipients: 
- ratio: "0.800000000000000000"
  recipient: fee_collector
- ratio: "0.200000000000000000"
  recipient: rewards
```

#### inflation

Fetch the last block inflation

Usage:

```bash
archwayd q mint inflation [flags]
```

Example:

```bash
archwayd q rewards inflation
```

Example output:

```yaml
inflation: "0.200000000000000000"
```
