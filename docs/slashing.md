# Slashing

## Unjailing

To unjail your jailed validator

```bash
archwayd tx slashing unjail --from <validator-operator-addr>
```

## Signing Info

To retrieve a validator's signing info:

```bash
archwayd query slashing signing-info <validator-pubkey>
```

## Query Parameters

You can get the current slashing parameters via:

```bash
archwayd query slashing params
```
