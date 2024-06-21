package types

import (
	context "context"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// WasmKeeper defines the expected interface needed to interact with the wasm module.
type WasmKeeper interface {
	// GetContractInfo returns the contract info for the given address
	GetContractInfo(ctx context.Context, contractAddress sdk.AccAddress) *wasmtypes.ContractInfo
	// GetCodeInfo returns the code info for the given codeID
	GetCodeInfo(ctx context.Context, codeID uint64) *wasmtypes.CodeInfo
}
