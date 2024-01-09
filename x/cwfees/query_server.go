package cwfees

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/cwfees/types"
)

var _ types.QueryServer = (*queryServer)(nil)

func NewQueryServer(k Keeper) types.QueryServer { return queryServer{k: k} }

type queryServer struct{ k Keeper }

func (q queryServer) IsGrantingContract(ctx context.Context, request *types.IsGrantingContractRequest) (*types.IsGrantingContractResponse, error) {
	addr, err := sdk.AccAddressFromBech32(request.ContractAddress)
	if err != nil {
		return nil, err
	}
	isGrantingContract, err := q.k.IsGrantingContract(ctx, addr)
	return &types.IsGrantingContractResponse{IsGrantingContract: isGrantingContract}, err
}
