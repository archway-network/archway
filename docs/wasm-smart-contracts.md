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

>**Note:** The produced wasm file is not optimized in size.
In order to have an optimized wasm file, we can use [rust-optimizer](https://github.com/CosmWasm/rust-optimizer).

## Deploying 

```bash
archwayd tx wasm store <wasmFilePath>\
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

## Instantiation

After deploying our wasm file, we need to instantiate it and get its address. 
To do so we need to run the following command.

**Note:** We have made some changes in execute command where we are passing extra parameters.

```bash
archwayd tx wasm instantiate <code_id>\
"{\"reward_address\": <address which will receive reward>,\
\"instantiation_request\": <base64 encoded json body>,\
\"gas_rebate_to_user\": <bool>,\
\"collect_premium\": <bool>,\
\"premium_percentage_charged\": <int in range of 1 to 200>}"\
--label test\
--from $MY_VALIDATOR_ADDRESS\
--keyring-backend test\
--chain-id testnet\
--fees 1stake
```

where `code_id` is a number that you will see in the output of deploying procedure which is `archwayd tx wasm store` command.

`instantiation_request` contains `InitMsg` of the smart contract in `JSON` format. It has to be encoded with `base64`.

If `gas_rebate_to_user` set to true, the developer does not get any reward, instead the user gets refunded. 
>**Note 1:** Due to a limitation in the Cosmos SDK, currently this feature doe not work; it will work with the new release of the SDK.

>**Note 2:** The inflation reward is paid always regardless of the value of this parameter.


If `collect_premium` set to true, the developer can charge an extra amount (in percentage). 
>**Note:** this parameter and `gas_rebate_to_user` cannot be true at the same time.

The premium charge is set in percentage with setting the `premium_percentage_charged` parameter which is calculated according to the gas usage and it can be a value from 1 to 200.

Here is an example of this command:
```bash
archwayd tx wasm instantiate 64 "{\"reward_address\": \"archway16w95tw2ueqdy0nvknkjv07zc287earxhwlykpt\", \"instantiation_request\": \"e30=\", \"gas_rebate_to_user\": false, \"collect_premium\": false, \"premium_percentage_charged\": 0}" --label test --from $MY_VALIDATOR_ADDRESS --keyring-backend test --chain-id testnet --fees 1stake
```

## More Info

- Other commands related to wasm can be found at [wasmd](https://docs.cosmwasm.com/) documentation.