# Errors

The module exposes the following error codes which are used with the x/cwerrors module in case of ica tx failures.

```proto
enum ModuleErrors {
  // ERR_UNKNOWN is the default error code
  ERR_UNKNOWN = 0;
  // ERR_PACKET_TIMEOUT is the error code for packet timeout
  ERR_PACKET_TIMEOUT = 1;
  // ERR_EXEC_FAILURE is the error code for tx execution failure
  ERR_EXEC_FAILURE = 2;
}
```