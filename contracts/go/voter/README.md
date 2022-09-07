# Voter contract

The contract is a voting dApp where users have to pay for creating a new voting or for participating in an existing one.
A contract creator can release raised funds and buy himself a cup of coffee.

The purpose of creating this contract was to test every feature the [cosmwasm-go SDK](https://github.com/CosmWasm/cosmwasm-go) and the [CosmWasm wasmvm](https://github.com/CosmWasm/wasmvm) provide.
Voter also utilizes all the Archway protocol [WASM bindings](../../../x/rewards/spec/08_wasm_bindings.md) and is used for [end-to-end](../../../e2e/voter_test.go) testing of the protocol.

Use the [Makefile](./Makefile) to run Unit / Integration tests and build a WASM blob.

If you want to investigate the code, start with the [contract.go](./src/contract.go) file as it has all the entrypoints called by the WASM VM.
