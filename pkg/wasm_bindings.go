package pkg

import (
	"encoding/json"
	"errors"

	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CustomQuerierExpected defines the expected interface for a custom WASM bindings querier.
type CustomQuerierExpected interface {
	DispatchQuery(ctx sdk.Context, request json.RawMessage) ([]byte, error)
}

// CustomQueryDispatcherPluginOption returns a WASM option for a custom query dispatcher.
// The wasmd keeper supports only one custom query handler and has no dispatching for it (no Next handler calls).
// This option builds a dispatcher to iterate over all registered custom query handlers.
// CONTRACT: a custom query handler must return the wasmVmTypes.UnsupportedRequest error if it does not support the query.
func CustomQueryDispatcherPluginOption(queriers ...CustomQuerierExpected) wasmKeeper.Option {
	return wasmKeeper.WithQueryPlugins(
		&wasmKeeper.QueryPlugins{
			Custom: NewCustomWasmQueryDispatcher(queriers...).DispatchQuery,
		},
	)
}

// CustomWasmQueryDispatcher is a custom WASM query dispatcher for multiple modules.
type CustomWasmQueryDispatcher struct {
	queriers []CustomQuerierExpected
}

// NewCustomWasmQueryDispatcher creates a new CustomWasmQueryDispatcher instance.
func NewCustomWasmQueryDispatcher(queriers ...CustomQuerierExpected) *CustomWasmQueryDispatcher {
	return &CustomWasmQueryDispatcher{
		queriers: queriers,
	}
}

// DispatchQuery iterates over all registered custom query handlers and dispatches the query to the first one that supports it.
func (q CustomWasmQueryDispatcher) DispatchQuery(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	for _, querier := range q.queriers {
		res, err := querier.DispatchQuery(ctx, request)
		if err != nil {
			var unsupportedErr *wasmVmTypes.UnsupportedRequest
			if errors.As(err, &unsupportedErr) {
				continue
			}
			return nil, err
		}

		return res, nil
	}

	return nil, wasmVmTypes.UnsupportedRequest{Kind: "unknown CustomWasmQuerier variant"}
}
