---

# ADR-009: Introduction of CWFees Module

Date: 2023-01-16

## Status

Accepted | Implemented.

## Context

The introduction of the CWFees module marks a significant advancement in how transaction fees are handled within the
ecosystem. Previously confined to entities registered under the standard x/feegrant module, the `FeeGranter` role within
a transaction's `AuthInfo.Fee` can now be assumed by a CosmWasm contract. This transition from a static,
binary-implemented fee grant logic to a dynamic, contract-based approach enhances flexibility, eliminating the need for
chain upgrades for modifications in fee grant logic.

## Decision

We have expanded the capabilities of the CWFees module by introducing two key message entry points: `RegisterAsGranter`
and `UnregisterAsGranter`, exclusively accessible to CosmWasm contracts. This enables a contract, for example during its
instantiation phase, to declare itself as a fee granter by issuing a `StargateMsg` containing a `RegisterAsGranter`
message. Once registered, the system acknowledges the contract as an eligible fee granter.

In the event of a transaction, the user can designate a registered CosmWasm contract as the `Tx.AuthInfo.Fee.Granter`.
The system, in turn, invokes this contract via sudo, providing it with critical information encapsulated in a structured
JSON format:

```golang
type SudoMsg struct {
CWGrant *CWGrant `json:"cw_grant"`
}

type CWGrant struct {
FeeRequested wasmVmTypes.Coins `json:"fee_requested"`
Msgs []CWGrantMessage `json:"msgs"`
}

type CWGrantMessage struct {
Sender string `json:"sender"`
TypeUrl string `json:"type_url"`
Msg []byte `json:"msg"`
}
```

This detailed information empowers the contract to make an informed decision regarding the grant request. If the contract
consents (i.e., no errors are returned), the runtime itself handles the transfer of fees from the contract to the
auth collector. Conversely, if the contract opts to decline the grant, it must issue an error response. Overall, the contract
is not required to do anything besides signaling it accepts or refuses the grant (no coin moving is required!).

## Consequences

### Positive
1. Grants developers enhanced control over the application of fee grants.
2. Enables transactions where contracts absorb the gas costs for users, creating a gas-less experience.
3. Supports the development of diverse incentive models.
4. Accommodates the use of multiple coin types for fee payments.

### Negative
1. Introduces a layer of complexity to the system's architecture.

### Security Considerations
- To mitigate potential risks, the gas usage within a CWGrant will be capped, preventing a CWFees contract from incurring excessive gas consumption.
