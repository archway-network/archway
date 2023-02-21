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
	GetFlatFee(ctx sdk.Context, contractAddr sdk.AccAddress) (sdk.Coin, bool)
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

// GetContractMetadata returns the contract metadata.
func (h QueryHandler) GetContractMetadata(ctx sdk.Context, req types.ContractMetadataRequest) (types.ContractMetadataResponse, error) {
	if err := req.Validate(); err != nil {
		return types.ContractMetadataResponse{}, fmt.Errorf("metadata: %w", err)
	}

	meta := h.rewardsKeeper.GetContractMetadata(ctx, req.MustGetContractAddress())
	if meta == nil {
		return types.ContractMetadataResponse{}, rewardsTypes.ErrMetadataNotFound
	}

	return types.NewContractMetadataResponse(*meta), nil
}

func (h QueryHandler) GetFlatFee(ctx sdk.Context, req types.ContractFlatFeeRequest) (types.ContractFlatFeeResponse, error) {
	if err := req.Validate(); err != nil {
		return types.ContractFlatFeeResponse{}, fmt.Errorf("flatfee: %w", err)
	}

	flatfee, found := h.rewardsKeeper.GetFlatFee(ctx, req.MustGetContractAddress())
	if !found {
		return types.ContractFlatFeeResponse{}, rewardsTypes.ErrContractFlatFeeNotFound
	}
	return types.NewContractFlatFeeResponse(flatfee), nil
}

// GetRewardsRecords returns the paginated list of types.RewardsRecord objects for a given account address.
func (h QueryHandler) GetRewardsRecords(ctx sdk.Context, req types.RewardsRecordsRequest) (types.RewardsRecordsResponse, error) {
	if err := req.Validate(); err != nil {
		return types.RewardsRecordsResponse{}, fmt.Errorf("rewardsRecords: %w", err)
	}

	var pageReq *query.PageRequest
	if req.Pagination != nil {
		req := req.Pagination.ToSDK()
		pageReq = &req
	}

	records, pageResp, err := h.rewardsKeeper.GetRewardsRecords(ctx, req.MustGetRewardsAddress(), pageReq)
	if err != nil {
		return types.RewardsRecordsResponse{}, sdkErrors.Wrap(rewardsTypes.ErrInvalidRequest, err.Error())
	}

	return types.NewRewardsRecordsResponse(records, *pageResp), nil
}
