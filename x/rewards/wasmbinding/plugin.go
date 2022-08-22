package wasmbinding

import (
	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
)

// GetCustomWasmMsgOption returns a WASM option for the custom x/rewards module msg handler.
func GetCustomWasmMsgOption(rKeeper RewardsReaderWriter) wasmKeeper.Option {
	return wasmKeeper.WithMessageHandlerDecorator(CustomMessageDecorator(rKeeper))
}

// GetCustomWasmQueryOption returns a WASM option for the custom x/rewards module querier.
func GetCustomWasmQueryOption(rKeeper RewardsReaderWriter) wasmKeeper.Option {
	return wasmKeeper.WithQueryPlugins(CustomQueryPlugin(rKeeper))
}
