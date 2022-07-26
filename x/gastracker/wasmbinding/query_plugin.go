package wasmbinding

import (
	"encoding/json"
	"fmt"

	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/archway-network/archway/x/gastracker"
	"github.com/archway-network/archway/x/gastracker/wasmbinding/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

//var _ wasmKeeper.CustomQuerier = (*QueryPlugin)(nil).DispatchQuery

// CustomQueryPlugin creates a new custom query plugin for WASM bindings.
func CustomQueryPlugin(gtKeeper ContractMetadataReader) *wasmKeeper.QueryPlugins {
	qp := NewQueryPlugin(gtKeeper)

	return &wasmKeeper.QueryPlugins{
		Custom: qp.DispatchQuery,
	}
}

// QueryPlugin provides custom WASM queries.
type QueryPlugin struct {
	gtKeeper ContractMetadataReader
}

// NewQueryPlugin creates a new QueryPlugin.
func NewQueryPlugin(gtKeeper ContractMetadataReader) *QueryPlugin {
	return &QueryPlugin{
		gtKeeper: gtKeeper,
	}
}

// DispatchQuery validates and executes a custom WASM query.
func (qp QueryPlugin) DispatchQuery(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	// Parse and validate the input
	var req types.Query
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, sdkErrors.Wrap(gastracker.ErrInvalidRequest, fmt.Sprintf("request JSON unmarshal: %v", err))
	}
	if err := req.Validate(); err != nil {
		return nil, sdkErrors.Wrap(gastracker.ErrInvalidRequest, fmt.Sprintf("request validation: %v", err))
	}

	// Execute the custom query (one of)
	var resData interface{}
	var resErr error
	switch {
	case req.Metadata != nil:
		resData, resErr = qp.getContractMetadata(ctx, *req.Metadata)
	default:
		resErr = sdkErrors.Wrap(gastracker.ErrInvalidRequest, "unknown request")
	}
	if resErr != nil {
		return nil, resErr
	}

	// Encode the response
	res, err := json.Marshal(resData)
	if err != nil {
		return nil, sdkErrors.Wrap(gastracker.ErrInternal, fmt.Sprintf("response JSON marshal: %v", err))
	}

	return res, nil
}

// getContractMetadata returns the contract metadata.
func (qp QueryPlugin) getContractMetadata(ctx sdk.Context, req types.ContractMetadataRequest) (types.ContractMetadataResponse, error) {
	meta, err := qp.gtKeeper.GetContractMetadata(ctx, req.GetContractAddress())
	if err != nil {
		return types.ContractMetadataResponse{}, err
	}

	return types.NewContractMetadataResponse(meta), nil
}
