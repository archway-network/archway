# Staking

## Set up a Validator

Please refer to the [Validator Setup](https://docs.archway.io/docs/validator/overview) documentations for a more complete guide on how to set up a validator-candidate.

## Delegate to a Validator

On the Archway network, you can delegate some tokens to a validator. These delegators can receive part of the validator's fee revenue. Read more about the [Cosmos Token Model](https://github.com/cosmos/cosmos/raw/master/Cosmos_Token_Model.pdf).

### Query Validators

You can query the list of all validators of a specific chain:

```bash
archwayd query staking validators
```

If you want to get the information of a single validator you can check it with:

```bash
archwayd query staking validator <account_archway_val>
```

## Bond Tokens

On the Archway mainnet, we delegate `uARCH`, where `1ARCH = 1000000uARCH`. Here's how you can bond tokens to a testnet validator (_i.e._ delegate):

```bash
archwayd tx staking delegate \
  --amount=10000000uARCH \
  --validator=<validator> \
  --from=<key_name> \
  --chain-id=<chain_id>
```

`<validator>` is the operator address of the validator to which you intend to delegate. If you are running a local testnet, you can find this with:

```bash
archwayd keys show [name] --bech val
```

where `[name]` is the name of the key you specified when you initialized `archwayd`.

While tokens are bonded, they are pooled with all the other bonded tokens in the network. Validators and delegators obtain a percentage of shares that equal their stake in this pool.

### Query Delegations

Once submitted a delegation to a validator, you can see it's information by using the following command:

```bash
archwayd query staking delegation <delegator_addr> <validator_addr>
```

Or if you want to check all your current delegations with disctinct validators:

```bash
archwayd query staking delegations <delegator_addr>
```

## Unbond Tokens

If for any reason the validator misbehaves, or you just want to unbond a certain amount of tokens, use this following command.

```bash
archwayd tx staking unbond \
  <validator_addr> \
  10ARCH \
  --from=<key_name> \
  --chain-id=<chain_id>
```

The unbonding will be automatically completed when the unbonding period has passed.

### Query Unbonding-Delegations

Once you begin an unbonding-delegation, you can see it's information by using the following command:

```bash
archwayd query staking unbonding-delegation <delegator_addr> <validator_addr>
```

Or if you want to check all your current unbonding-delegations with disctinct validators:

```bash
archwayd query staking unbonding-delegations <account_cosmos>
```

Additionally, as you can get all the unbonding-delegations from a particular validator:

```bash
archwayd query staking unbonding-delegations-from <account_archway_val>
```

## Redelegate Tokens

A redelegation is a type delegation that allows you to bond illiquid tokens from one validator to another:

```bash
archwayd tx staking redelegate \
  <src-validator-operator-addr> \
  <dst-validator-operator-addr> \
  10ARCH \
  --from=<key_name> \
  --chain-id=<chain_id>
```

Here you can also redelegate a specific `shares-amount` or a `shares-fraction` with the corresponding flags.

The redelegation will be automatically completed when the unbonding period has passed.

### Query Redelegations

Once you begin an redelegation, you can see it's information by using the following command:

```bash
archwayd query staking redelegation <delegator_addr> <src_val_addr> <dst_val_addr>
```

Or if you want to check all your current unbonding-delegations with distinct validators:

```bash
archwayd query staking redelegations <account_cosmos>
```

Additionally, as you can get all the outgoing redelegations from a particular validator:

```bash
  archwayd query staking redelegations-from <account_archway_val>
```

## Query Parameters

Parameters define high level settings for staking. You can get the current values by using:

```bash
archwayd query staking params
```

With the above command you will get the values for:

- Unbonding time
- Maximum numbers of validators
- Coin denomination for staking

All these values will be subject to updates though a `governance` process by `ParameterChange` proposals.

## Query Pool

A staking `Pool` defines the dynamic parameters of the current state. You can query them with the following command:

```bash
archwayd query staking pool
```

With the `pool` command you will get the values for:

- Not-bonded and bonded tokens
- Token supply
- Current annual inflation and the block in which the last inflation was processed
- Last recorded bonded shares

### Query Delegations To Validator

You can also query all of the delegations to a particular validator:

```bash
  archwayd query delegations-to <account_archway_val>
```
