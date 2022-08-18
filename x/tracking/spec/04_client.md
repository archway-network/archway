<!--
order: 4
-->

# Client

Section describes interaction with the module by a user.

## CLI

### Query

The `query` commands allows a user to query the module state.

Use the `-h` / `--help` flag to get a help description of a command.

```bash
archwayd q tracking -h
```

> You can add the `-o json` for the JSON output format.

### block-gas-tracking

Get the current gas tracking data.

```bash
archwayd q tracking block-gas-tracking [flags]
```

Example output:

```yaml
txs: 
  - info:
      id: 1
      height: 2
      TotalGas: 1000
    contract_operations:
      - id: 1
        tx_id: 1
        contract_address: archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u
        operation_type: 1
        vm_gas: 900
        sdk_gas: 100
  - info:
      id: 2
      height: 2
      TotalGas: 10000
    contract_operations:
      - id: 1
        tx_id: 2
        contract_address: archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u
        operation_type: 2
        vm_gas: 8000
        sdk_gas: 1700
      - id: 2
        tx_id: 2
        contract_address: archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u
        operation_type: 3
        vm_gas: 300
        sdk_gas: 0
```
