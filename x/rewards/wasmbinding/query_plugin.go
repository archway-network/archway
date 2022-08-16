package wasmbinding

import (
	"encoding/json"
	"fmt"

	wasmKeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

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
	case req.RewardsRecords != nil:
		var pageReq *query.PageRequest
		if req.RewardsRecords.Pagination != nil {
			req := req.RewardsRecords.Pagination.ToSDK()
			pageReq = &req
		}

		resData, resErr = qp.getRewardsRecords(ctx, req.RewardsRecords.MustGetRewardsAddress(), pageReq)
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

// getRewardsRecords returns the paginated list of types.RewardsRecord objects for a given account address.
func (qp QueryPlugin) getRewardsRecords(ctx sdk.Context, rewardsAddr sdk.AccAddress, pageReq *query.PageRequest) (types.RewardsRecordsResponse, error) {
	maxWithdrawRecords := qp.rewardsKeeper.MaxWithdrawRecords(ctx)

	if pageReq == nil {
		pageReq = &query.PageRequest{
			Limit: maxWithdrawRecords,
		}
	}
	if pageReq.Limit > maxWithdrawRecords {
		return types.RewardsRecordsResponse{}, sdkErrors.Wrapf(rewardsTypes.ErrInvalidRequest, "max records (%d) query limit exceeded", maxWithdrawRecords)
	}

	records, pageResp, err := qp.rewardsKeeper.GetRewardsRecords(ctx, rewardsAddr, pageReq)
	if err != nil {
		return types.RewardsRecordsResponse{}, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, err.Error())
	}

	return types.NewRewardsRecordsResponse(records, *pageResp), nil
}
