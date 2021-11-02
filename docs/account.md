# Account

## Get Tokens

**Note:** Querying account balance, send and receive tokens work only if you are running a full node on your machine or you are joined a network.

On a testnet, getting tokens is usually done via a faucet.

### Example:
```bash
curl \
-X POST "https://faucet.constantine-1.archway.tech/" \
-H  "accept: application/json" \
-H  "Content-Type: application/json" \
-d "{    \"denom\": \"uconst\",    \"address\": \"archway1vt3fdsux63ucyndhk6gsx7rgn2495kwsmr9mxx\"}"
```



## Query Account Balance

After receiving tokens to your address, you can view your account's balance by typing:

```bash
archwayd query account <account_archway>
```

**Warning:** When you query an account balance with zero tokens, you will get this error: `No account with address <account_archway> was found in the state.` This can also happen if you fund the account before your node has fully synced with the chain. These are both normal.


## Send Tokens

The following command could be used to send coins from one account to another:

```bash
archwayd tx bank send <sender_key_name_or_address> <recipient_address> 10ARCH \
  --chain-id=<chain_id>
```

**Warning:** The `amount` argument accepts the format `<value|coin_name>`.

**Tip:** You may want to cap the maximum gas that can be consumed by the transaction via the `--gas` flag.
If you pass `--gas=auto`, the gas supply will be automatically estimated before executing the transaction.
Gas estimate might be inaccurate as state changes could occur in between the end of the simulation and the actual execution of a transaction, thus an adjustment is applied on top of the original estimate in order to ensure the transaction is broadcasted successfully. The adjustment can be controlled via the `--gas-adjustment` flag, whose default value is 1.0.


Now, view the updated balances of the origin and destination accounts:

```bash
archwayd query account <account_archway>
archwayd query account <destination_archway>
```

You can also check your balance at a given block by using the `--block` flag:

```bash
archwayd query account <account_archway> --block=<block_height>
```

You can simulate a transaction without actually broadcasting it by appending the
`--dry-run` flag to the command line:

```bash
archwayd tx bank send <sender_key_name_or_address> <destination_archway_acc_addr> 10ARCH \
  --chain-id=<chain_id> \
  --dry-run
```

Furthermore, you can build a transaction and print its JSON format to STDOUT by
appending `--generate-only` to the list of the command line arguments:

```bash
archwayd tx bank send <sender_address> <recipient_address> 10faucetToken \
  --chain-id=<chain_id> \
  --generate-only > unsignedSendTx.json
```

```bash
archwayd tx sign \
  --chain-id=<chain_id> \
  --from=<key_name> \
  unsignedSendTx.json > signedSendTx.json
```

**Tip:** The `--generate-only` flag prevents `archwayd` from accessing the local keybase.
Thus when such flag is supplied `<sender_key_name_or_address>` must be an address.


You can validate the transaction's signatures by typing the following:

```bash
archwayd tx sign --validate-signatures signedSendTx.json
```

You can broadcast the signed transaction to a node by providing the JSON file to the following command:

```bash
archwayd tx broadcast --node=<node> signedSendTx.json
```
