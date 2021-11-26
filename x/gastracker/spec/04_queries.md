# Queries 

## Gas Tracking Query Wasm Plugin
To intersect query interactions within smart contract we create a custom plugin, this plugin performs query validation, tracks the gas. It may also:
- Send gas rebate to user
- Charge contract premium

Depending on wether the features are enabled for the contract or within the chain.


### Gas Tracking Query Request Wrapper
Our custom WasmEngine uses a custom wrapper with a magic string, that allows further validation of the query

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

