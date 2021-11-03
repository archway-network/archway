# Fee Distribution

## Query Distribution Parameters

To check the current distribution parameters, run:

```bash
archwayd query distribution params
```

## Query distribution Community Pool

To query all coins in the community pool which is under Governance control:

```bash
archwayd query distribution community-pool
```

## Query outstanding rewards

To check the current outstanding (un-withdrawn) rewards, run:

```bash
archwayd query distribution outstanding-rewards
```

## Query Validator Commission

To check the current outstanding commission for a validator, run:

```bash
archwayd query distribution commission <validator_address>
```

## Query Validator Slashes

To check historical slashes for a validator, run:

```bash
archwayd query distribution slashes <validator_address> <start_height> <end_height>
```

## Query Delegator Rewards

To check current rewards for a delegation (were they to be withdrawn), run:

```bash
archwayd query distribution rewards <delegator_address> <validator_address>
```

## Query All Delegator Rewards

To check all current rewards for a delegation (were they to be withdrawn), run:

```bash
archwayd query distribution rewards <delegator_address>
```
