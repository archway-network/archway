package wasmbinding

import (
	"encoding/json"
	"fmt"

	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
	"github.com/archway-network/archway/x/rewards/wasmbinding/types"
)

//var _ wasmKeeper.CustomQuerier = (*QueryPlugin)(nil).DispatchQuery

// CustomQueryPlugin creates a new custom query plugin for WASM bindings.
func CustomQueryPlugin(gtKeeper RewardsReader) *wasmKeeper.QueryPlugins {
	qp := NewQueryPlugin(gtKeeper)

	return &wasmKeeper.QueryPlugins{
		Custom: qp.DispatchQuery,
	}
}

// QueryPlugin provides custom WASM queries.
type QueryPlugin struct {
	rewardsKeeper RewardsReader
}

// NewQueryPlugin creates a new QueryPlugin.
func NewQueryPlugin(rk RewardsReader) *QueryPlugin {
	return &QueryPlugin{
		rewardsKeeper: rk,
	}
}

// DispatchQuery validates and executes a custom WASM query.
func (qp QueryPlugin) DispatchQuery(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
	// Parse and validate the input
	var req types.Query
	if err := json.Unmarshal(request, &req); err != nil {
		return nil, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, fmt.Sprintf("request JSON unmarshal: %v", err))
	}
	if err := req.Validate(); err != nil {
		return nil, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, fmt.Sprintf("request validation: %v", err))
	}

	// Execute the custom query (one of)
	var resData interface{}
	var resErr error
	switch {
	case req.Metadata != nil:
		resData, resErr = qp.getContractMetadata(ctx, req.Metadata.MustGetContractAddress())
	case req.CurrentRewards != nil:
		resData, resErr = qp.getCurrentRewards(ctx, req.CurrentRewards.MustGetRewardsAddress())
	default:
		return nil, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, "unknown request")
	}
	if resErr != nil {
		return nil, resErr
	}

	// Encode the response
	res, err := json.Marshal(resData)
	if err != nil {
		return nil, sdkErrors.Wrap(rewardsTypes.ErrInternal, fmt.Sprintf("response JSON marshal: %v", err))
	}

	return res, nil
}

// getContractMetadata returns the contract metadata.
func (qp QueryPlugin) getContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) (types.ContractMetadataResponse, error) {
	meta := qp.rewardsKeeper.GetContractMetadata(ctx, contractAddr)
	if meta == nil {
		return types.ContractMetadataResponse{}, rewardsTypes.ErrMetadataNotFound
	}

	return types.NewContractMetadataResponse(*meta), nil
}

// getCurrentRewards returns the current rewards for a given account address.
func (qp QueryPlugin) getCurrentRewards(ctx sdk.Context, rewardsAddr sdk.AccAddress) (types.CurrentRewardsResponse, error) {
	rewards := qp.rewardsKeeper.GetCurrentRewards(ctx, rewardsAddr)

	return types.NewCurrentRewardsResponse(rewards), nil
}
