# Transactions

## Tx Broadcasting

When broadcasting transactions, `archwayd` accepts a `--broadcast-mode` flag. This
flag can have a value of `sync` (default), `async`, or `block`, where `sync` makes
the client return a CheckTx response, `async` makes the client return immediately,
and `block` makes the client wait for the tx to be committed (or timing out).

It is important to note that the `block` mode should **not** be used in most
circumstances. This is because broadcasting can timeout but the tx may still be
included in a block. This can result in many undesirable situations. Therefore, it
is best to use `sync` or `async` and query by tx hash to determine when the tx
is included in a block.

## Fees & Gas

Each transaction may either supply fees or gas prices, but not both.

Validator's have a minimum gas price (multi-denom) configuration and they use
this value when when determining if they should include the transaction in a block during `CheckTx`, where `gasPrices >= minGasPrices`. Note, your transaction must supply fees that are greater than or equal to **any** of the denominations the validator requires.

**Note**: With such a mechanism in place, validators may start to prioritize
txs by `gasPrice` in the mempool, so providing higher fees or gas prices may yield higher tx priority.

e.g.

```bash
archwayd tx bank send ... --fees=50000ARCH
```

or

```bash
archwayd tx bank send ... --gas-prices=0.0025ARCH
```