package keeper

import (
	"github.com/archway-network/archway/x/cwregistry/types"
)

var _ types.QueryServer = &QueryServer{}

type QueryServer struct {
	keeper Keeper
}

// NewQueryServer creates a new gRPC query server.
func NewQueryServer(keeper Keeper) *QueryServer {
	return &QueryServer{
		keeper: keeper,
	}
}
