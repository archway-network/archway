# CWRegistry

This module can be used by contract deployers to provide metadata about their contracts. The module stores contract metadata such as source code, schema, developer contact info.

## Concepts

Cosmwasm does not provide any way for a Cosmwasm smart contract developer to provide any metadata regarding their contracts. This has been explored in `x/wasm` before [^1], where during contract binary upload, a developer could provide the source code url. This feature was deprecated by Confio due to
1. Field was often unfilled or had erroneous values
2. No tooling to verify the contracts match the given information   
Due to the nature of wasm, it is also not possible to take a look at the source code of a deployed contract.

Once a contract is deployed, it is not easy for external parties to get the contract schema and endpoints[^2], especially so in the case when the contract is closed source or source URL not available. Having this information available on chain would enable the following
1. Make third party tooling of contracts easier to develop, such as code gen for UI
2. General purpose contract interaction tools
3. Indexers and block exploreres could use this information to better display the contract state.  
4. CW based multisigs and DAOs can perform contract interactions knowing what the contarct expects.

Currently, there is no way for a user/another developer to know who deployed a contract. In case they would like to contact the developer, there isnt any way to do it beyond the deployer address. Adding a field for security contact would help others report issues.

Most of the Cosmwasm chains run as permissioned Cosmwasm, which allows for the contract source to be connected to the binary in the governance proposal. However, in the permissionless approach of Archway, there is no builtin way to establish this connection.

## Contents

1. [State](./01_state.md)
2. [Messages](./02_messages.md)


## References

[RFC - x/cw-registry module](https://github.com/orgs/archway-network/discussions/16)




[^1]: [Question: Why was StoreCode.url removed from the tx msg?](https://github.com/CosmWasm/wasmd/issues/742)

[^2]: [Upload JSON schema alongside code contract](https://github.com/CosmWasm/wasmd/issues/241)