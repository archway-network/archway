package rewards

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/archway-network/archway/wasmbinding/rewards/types"
	rewardsTypes "github.com/archway-network/archway/x/rewards/types"
)

// KeeperReaderExpected defines the x/rewards keeper expected read operations.
type KeeperReaderExpected interface {
	GetContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) *rewardsTypes.ContractMetadata
	GetRewardsRecords(ctx sdk.Context, rewardsAddr sdk.AccAddress, pageReq *query.PageRequest) ([]rewardsTypes.RewardsRecord, *query.PageResponse, error)
	MaxWithdrawRecords(ctx sdk.Context) uint64
}

// QueryHandler provides a custom WASM query handler for the x/rewards module.
type QueryHandler struct {
	rewardsKeeper KeeperReaderExpected
}

// NewQueryHandler creates a new QueryHandler instance.
func NewQueryHandler(rk KeeperReaderExpected) QueryHandler {
	return QueryHandler{
		rewardsKeeper: rk,
	}
}

// DispatchQuery validates and executes a custom WASM query.
func (h QueryHandler) DispatchQuery(ctx sdk.Context, req types.Query) (interface{}, error) {
	// Validate the input
	if err := req.Validate(); err != nil {
		return nil, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, fmt.Sprintf("x/rewards: sub-query validation: %v", err))
	}

	// Execute request (one of)
	switch {
	case req.Metadata != nil:
		return h.getContractMetadata(ctx, req.Metadata.MustGetContractAddress())
	case req.RewardsRecords != nil:
		var pageReq *query.PageRequest
		if req.RewardsRecords.Pagination != nil {
			req := req.RewardsRecords.Pagination.ToSDK()
			pageReq = &req
		}

		return h.getRewardsRecords(ctx, req.RewardsRecords.MustGetRewardsAddress(), pageReq)
	default:
		return nil, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, "x/rewards: unknown request")
	}
}

// getContractMetadata returns the contract metadata.
func (h QueryHandler) getContractMetadata(ctx sdk.Context, contractAddr sdk.AccAddress) (types.ContractMetadataResponse, error) {
	meta := h.rewardsKeeper.GetContractMetadata(ctx, contractAddr)
	if meta == nil {
		return types.ContractMetadataResponse{}, rewardsTypes.ErrMetadataNotFound
	}

	return types.NewContractMetadataResponse(*meta), nil
}

// getRewardsRecords returns the paginated list of types.RewardsRecord objects for a given account address.
func (h QueryHandler) getRewardsRecords(ctx sdk.Context, rewardsAddr sdk.AccAddress, pageReq *query.PageRequest) (types.RewardsRecordsResponse, error) {
	maxWithdrawRecords := h.rewardsKeeper.MaxWithdrawRecords(ctx)

	if pageReq == nil {
		pageReq = &query.PageRequest{
			Limit: maxWithdrawRecords,
		}
	}
	if pageReq.Limit > maxWithdrawRecords {
		return types.RewardsRecordsResponse{}, sdkErrors.Wrapf(rewardsTypes.ErrInvalidRequest, "max records (%d) query limit exceeded", maxWithdrawRecords)
	}

	records, pageResp, err := h.rewardsKeeper.GetRewardsRecords(ctx, rewardsAddr, pageReq)
	if err != nil {
		return types.RewardsRecordsResponse{}, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, err.Error())
	}

	return types.NewRewardsRecordsResponse(records, *pageResp), nil
}
