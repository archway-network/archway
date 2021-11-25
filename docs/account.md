# Account

## Get Tokens

**Note:** Querying account balance, send and receive tokens work only if you are running a full node on your machine or you have joined a network.

On a testnet, getting tokens is usually done via a faucet.

### Example:
```bash
curl \
-X POST "https://faucet.constantine-1.archway.tech/" \
-H  "accept: application/json" \
-H  "Content-Type: application/json" \
-d "{    \"denom\": \"uconst\",    \"address\": \"archway1vt3fdsux63ucyndhk6gsx7rgn2495kwsmr9mxx\"}"
```

Please checkout the rest of the commands in [`gaiad` account management](https://github.com/cosmos/gaia/blob/main/docs/resources/gaiad.md#get-tokens).