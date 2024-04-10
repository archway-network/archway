# Errors

The module exposes the following error codes which are used with the x/cwerrors module in case of error callback failures.

```proto
enum ModuleErrors {
  // ERR_UNKNOWN is the default error code
  ERR_UNKNOWN = 0;
  // ERR_CALLBACK_EXECUTION_FAILED is the error code for when the error callback fails
  ERR_CALLBACK_EXECUTION_FAILED = 1;
}
```