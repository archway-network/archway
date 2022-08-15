<!--
order: 7
-->

# Client

Section describes interaction with the module by a user.

## CLI

### Query

The `query` commands allows a user to query the module state.

Use the `-h` / `--help` flag to get a help description of a command.

```bash
archwayd q rewards -h
```

> You can add the `-o json` for the JSON output format.

#### params

Get the current module parameters.

Usage:

```bash
archwayd q rewards params [flags]
```

Example output:

```yaml
inflation_rewards_ratio: "0.200000000000000000"
tx_fee_rebate_ratio: "0.500000000000000000"
```

#### estimate-fees

Estimate the minimum transaction fees based on transaction gas limit.

Usage:

```bash
archwayd q rewards estimate-fees [transaction-gas-limit] [flags]
```

Example:

```bash
archwayd q rewards estimate-fees 100000
```

Example output:

```yaml
estimated_fee:
  amount: "1268"
  denom: uarch
gas_unit_price:
  amount: "0.012675360000000000"
  denom: uarch
```

#### contract-metadata

Get an existing contract metadata. Query fails if a contract is not *Instantiated* or its metadata is not set.

Usage:

```bash
archwayd q rewards contract-metadata [contract-address] [flags]
```

Example output:

```yaml
contract_address: archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u
owner_address: archway12reqvcenxgv5s7z96pkytzajtl4lf2epyfman2
rewards_address: archway12reqvcenxgv5s7z96pkytzajtl4lf2epyfman2
```

#### rewards

Get the current credited dApp rewards for an account. Those rewards are "ready" for the *withdraw* operation.

Usage:

```bash
archwayd q rewards rewards [rewards-address] [flags]
```

Example output:

```yaml
rewards:
- amount: "6460"
  denom: uarch
```

#### block-rewards-tracking

Get the current rewards tracking state (tracked inflation and tx fee rebate rewards).

> Use the `--height` flag to specify the block height.

Usage:

```bash
archwayd q rewards block-rewards-tracking [flags]
```

Example:

```bash
archwayd q rewards block-rewards-tracking --height 3189
```

Example output:

```yaml
block:
  inflation_rewards:
    height: "3189"
    inflation_rewards:
      amount: "633768"
      denom: uarch
    max_gas: "100000000"
  tx_rewards:
  - fee_rewards:
    - amount: "6337"
      denom: uarch
    height: "3189"
    tx_id: "9"
```

#### pool

Get the current rewards pool balance (undistributed yet tokens).

Usage:

```bash
archwayd q rewards pool [flags]
```

Example output:

```yaml
funds:
- amount: "2038832654"
  denom: uarch
```

### Transactions

The `tx` commands allows a user to interact with the module.

Use the `-h` / `--help` flag to get a help description of a command.

```bash
archwayd tx rewards -h
```

#### set-contract-metadata

Create / update a contract metadata state. Operation is authorized to:

* Creating metadata: contract admin (set via CosmWasm *Instantiate* operation);
* Updating metadata: metadata's `owner_address`;

Usage:

```bash
archwayd tx rewards set-contract-metadata [contract-address] [flags]
```

Command specific flags:

* `--owner-address` - update the contract owner address;
* `--rewards-address` - update the contract rewards receiver address;

Example (delegate rewards ownership to the contract):

```bash
archwayd tx rewards set-contract-metadata archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u \
  --owner-address archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u \
  --rewards-address archway14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9sy85n2u \
  --from myAccountKey
  --fees 1500uarch
```

#### withdraw-rewards

Withdraw the current credited dApp rewards to a sender account.

Usage:

```bash
archwayd tx rewards withdraw-rewards [flags]
```

Example:

```bash
archwayd tx rewards withdraw-rewards \
  --from myAccountKey
  --fees 1500uarch
```

#### 
