package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/archway-network/archway/x/oracle/asset"
	"github.com/archway-network/archway/x/oracle/types"
)

// querier is used as Keeper will have duplicate methods if used directly, and gRPC names take precedence over q
type querier struct {
	Keeper
}

// NewQuerier returns an implementation of the oracle QueryServer interface
// for the provided Keeper.
func NewQuerier(keeper Keeper) types.QueryServer {
	return &querier{Keeper: keeper}
}

var _ types.QueryServer = querier{}

// Params queries params of distribution module
func (q querier) Params(c context.Context, _ *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var params types.Params

	params, err := q.Keeper.Params.Get(ctx)
	if err != nil {
		return nil, err
	}

	return &types.QueryParamsResponse{Params: params}, nil
}

// ExchangeRate queries exchange rate of a pair
func (q querier) ExchangeRate(c context.Context, req *types.QueryExchangeRateRequest) (*types.QueryExchangeRateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if len(req.Pair) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty pair")
	}

	ctx := sdk.UnwrapSDKContext(c)
	exchangeRate, err := q.Keeper.GetExchangeRate(ctx, req.Pair)
	if err != nil {
		return nil, err
	}

	return &types.QueryExchangeRateResponse{ExchangeRate: exchangeRate}, nil
}

/*
Gets the time-weighted average price from ( ctx.BlockTime() - interval, ctx.BlockTime() ]
Note the open-ended right bracket.

If there's only one snapshot, then this function returns the price from that single snapshot.

Returns -1 if there's no price.
*/
func (q querier) ExchangeRateTwap(c context.Context, req *types.QueryExchangeRateRequest) (response *types.QueryExchangeRateResponse, err error) {
	if _, err = q.ExchangeRate(c, req); err != nil {
		return
	}

	ctx := sdk.UnwrapSDKContext(c)
	twap, err := q.Keeper.GetExchangeRateTwap(ctx, req.Pair)
	if err != nil {
		return &types.QueryExchangeRateResponse{}, err
	}
	return &types.QueryExchangeRateResponse{ExchangeRate: twap}, nil
}

// ExchangeRates queries exchange rates of all pairs
func (q querier) ExchangeRates(ctx context.Context, _ *types.QueryExchangeRatesRequest) (*types.QueryExchangeRatesResponse, error) {

	var exchangeRates types.ExchangeRateTuples
	q.Keeper.ExchangeRates.Walk(ctx, nil, func(key asset.Pair, value types.DatedPrice) (bool, error) {
		exchangeRates = append(exchangeRates, types.ExchangeRateTuple{
			Pair:         key,
			ExchangeRate: value.ExchangeRate,
		})
		return false, nil
	})

	return &types.QueryExchangeRatesResponse{ExchangeRates: exchangeRates}, nil
}

// Actives queries all pairs for which exchange rates exist
func (q querier) Actives(ctx context.Context, _ *types.QueryActivesRequest) (*types.QueryActivesResponse, error) {
	iter, err := q.Keeper.ExchangeRates.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	keys, err := iter.Keys()
	if err != nil {
		return nil, err
	}
	return &types.QueryActivesResponse{
		Actives: keys,
	}, nil
}

// VoteTargets queries the voting target list on current vote period
func (q querier) VoteTargets(c context.Context, _ *types.QueryVoteTargetsRequest) (*types.QueryVoteTargetsResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	targets, err := q.GetWhitelistedPairs(ctx)
	if err != nil {
		return nil, err
	}
	return &types.QueryVoteTargetsResponse{
		VoteTargets: targets,
	}, nil
}

// FeederDelegation queries the account address that the validator operator delegated oracle vote rights to
func (q querier) FeederDelegation(ctx context.Context, req *types.QueryFeederDelegationRequest) (*types.QueryFeederDelegationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	valueBytes, err := q.Keeper.FeederDelegations.Get(ctx, valAddr)
	var value sdk.Address
	if err == nil {
		value = sdk.AccAddress(valueBytes)
	} else {
		value = sdk.AccAddress(valAddr)
	}
	return &types.QueryFeederDelegationResponse{
		FeederAddr: value.String(),
	}, nil
}

// MissCounter queries oracle miss counter of a validator
func (q querier) MissCounter(ctx context.Context, req *types.QueryMissCounterRequest) (*types.QueryMissCounterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	value, err := q.MissCounters.Get(ctx, valAddr)
	if err != nil {
		value = 0
	}
	return &types.QueryMissCounterResponse{
		MissCounter: value,
	}, nil
}

// AggregatePrevote queries an aggregate prevote of a validator
func (q querier) AggregatePrevote(ctx context.Context, req *types.QueryAggregatePrevoteRequest) (*types.QueryAggregatePrevoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	prevote, err := q.Prevotes.Get(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryAggregatePrevoteResponse{
		AggregatePrevote: prevote,
	}, nil
}

// AggregatePrevotes queries aggregate prevotes of all validators
func (q querier) AggregatePrevotes(ctx context.Context, _ *types.QueryAggregatePrevotesRequest) (*types.QueryAggregatePrevotesResponse, error) {
	iter, err := q.Prevotes.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	values, err := iter.Values()
	if err != nil {
		return nil, err
	}
	return &types.QueryAggregatePrevotesResponse{
		AggregatePrevotes: values,
	}, nil
}

// AggregateVote queries an aggregate vote of a validator
func (q querier) AggregateVote(ctx context.Context, req *types.QueryAggregateVoteRequest) (*types.QueryAggregateVoteResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	vote, err := q.Keeper.Votes.Get(ctx, valAddr)
	if err != nil {
		return nil, err
	}

	return &types.QueryAggregateVoteResponse{
		AggregateVote: vote,
	}, nil
}

// AggregateVotes queries aggregate votes of all validators
func (q querier) AggregateVotes(ctx context.Context, _ *types.QueryAggregateVotesRequest) (*types.QueryAggregateVotesResponse, error) {
	iter, err := q.Keeper.Votes.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	values, err := iter.Values()
	if err != nil {
		return nil, err
	}
	return &types.QueryAggregateVotesResponse{
		AggregateVotes: values,
	}, nil
}
