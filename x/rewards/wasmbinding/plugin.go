package wasmbinding

import (
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

// GetCustomWasmOptions returns WASM options for the custom querier and for the custom msg handler.
func GetCustomWasmOptions(gasTrackerKeeper RewardsReaderWriter) []wasmKeeper.Option {
	return []wasmKeeper.Option{
		wasmKeeper.WithQueryPlugins(CustomQueryPlugin(gasTrackerKeeper)),
		wasmKeeper.WithMessageHandlerDecorator(CustomMessageDecorator(gasTrackerKeeper)),
	}
}
