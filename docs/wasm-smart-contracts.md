# WASM Smart Contracts

## Build

To build the smart contract, navigate to your project's directory and execute this:

```bash
cargo build
```

## Test

```bash
cargo unit-test
```
Runs the tests.

## Dry Run

```bash
cargo wasm
```
>**Tip:** use dry run before deploying to gauge whether the deployment will succeed. This is useful because of speed, as dry running is a lot faster.

## Deploying 

```bash
archwayd tx  wasm  store <wasmFilePath>\
--from <walletLabel>\
--chain-id <chainId>\
--node <node>\
--gas-prices <gasPrices>\
--gas <gas>\
--gas-adjustment <gasAdjustment>
```

Where `<wasmFilePath>` is the path to the wasm file, 
`<node>` is the RPC address of the node,
`<gasPrices>` indicates the gas prices, `<gas>` refers to the gas mode, and `<gasAdjustment>` is self explanatory.

>**Warning:** make sure your wallet has enough tokens before deploying.
