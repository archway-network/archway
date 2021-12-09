# Queries 

## Gas Tracking Query Wasm Plugin
To track  smart contract query we create a custom plugin, which tracks the gas. It may also:
- Send gas rebate to user
- Charge contract premium

Depending on whether the features are enabled for the contract or within the chain.

### Gas Tracking Query Request Wrapper
Our custom WasmEngine uses a custom wrapper that allows to determine whether a query is coming from tx or if is part of the normal rpc query.

```
// Custom wrapper around Query request
message GasTrackingQueryRequestWrapper {
  string magic_string = 1;
  bytes query_request = 2;
}
```

### Gas Tracking Query Response Wrapper
Our Custom Response Wrapper provides infomration on the gas consumed for this query
```

// Custom wrapper around Query result that also gives gas consumption
message GasTrackingQueryResultWrapper {
  uint64 gas_consumed = 1;
  bytes query_response = 2;
}

```

