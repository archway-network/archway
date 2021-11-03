# Query Transactions

**Note:** These commands work only if you are running a full node on your machine or you are joined a network.

## Matching a Set of Events

You can use the transaction search command to query for transactions that match a
specific set of `events`, which are added on every transaction.

Each event is composed by a key-value pair in the form of `{eventType}.{eventAttribute}={value}`.
Events can also be combined to query for a more specific result using the `&` symbol.

You can query transactions by `events` as follows:

```bash
archwayd query txs --events='message.sender=archway1...'
```

And for using multiple `events`:

```bash
archwayd query txs --events='message.sender=archway1...&message.action=withdraw_delegator_reward'
```

The pagination is supported as well via `page` and `limit`:

```bash
archwayd query txs --events='message.sender=archway1...' --page=1 --limit=20
```
**Tip:** The action tag always equals the message type returned by the `Type()` function of the relevant message.

You can find a list of available `events` on each of the SDK modules:

- [Staking events](https://github.com/cosmos/cosmos-sdk/blob/master/x/staking/spec/07_events.md)
- [Governance events](https://github.com/cosmos/cosmos-sdk/blob/master/x/gov/spec/04_events.md)
- [Slashing events](https://github.com/cosmos/cosmos-sdk/blob/master/x/slashing/spec/06_events.md)
- [Distribution events](https://github.com/cosmos/cosmos-sdk/blob/master/x/distribution/spec/06_events.md)
- [Bank events](https://github.com/cosmos/cosmos-sdk/blob/master/x/bank/spec/04_events.md)


## Matching a Transaction's Hash

You can also query a single transaction by its hash using the following command:

```bash
archwayd query tx [hash]
```
