package src

import (
	"encoding/json"

	"github.com/CosmWasm/cosmwasm-go/std"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"

	archwayCustomTypes "github.com/archway-network/voter/src/pkg/archway/custom"
	archwayProtoTypes "github.com/archway-network/voter/src/pkg/archway/proto"
	"github.com/archway-network/voter/src/state"
	"github.com/archway-network/voter/src/types"
)

// queryParams handles MsgQuery.Params query.
func queryParams(deps *std.Deps) (*types.QueryParamsResponse, error) {
	stateParams, err := state.GetParams(deps.Storage)
	if err != nil {
		return nil, err
	}

	queryParams, err := types.NewParamsFromState(deps.Api, stateParams)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &types.QueryParamsResponse{
		Params: queryParams,
	}, nil
}

// queryVoting handles MsgQuery.Voting query.
func queryVoting(deps *std.Deps, req types.QueryVotingRequest) (*types.QueryVotingResponse, error) {
	voting, err := state.GetVoting(deps.Storage, req.ID)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	if voting == nil {
		return nil, stdTypes.NotFound{Kind: "voting"}
	}

	return &types.QueryVotingResponse{
		Voting: *voting,
	}, nil
}

// queryTally handles MsgQuery.Tally query.
func queryTally(deps *std.Deps, env stdTypes.Env, req types.QueryTallyRequest) (*types.QueryTallyResponse, error) {
	voting, err := state.GetVoting(deps.Storage, req.ID)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	if voting == nil {
		return nil, stdTypes.NotFound{Kind: "voting"}
	}

	resp := types.QueryTallyResponse{
		Open:  !voting.IsClosed(env.Block.Time),
		Votes: make([]types.VoteTally, 0, len(voting.Tallies)),
	}

	for _, tally := range voting.Tallies {
		resp.Votes = append(
			resp.Votes,
			types.VoteTally{
				Option:   tally.Option,
				TotalYes: uint32(len(tally.YesAddrs)),
				TotalNo:  uint32(len(tally.NoAddrs)),
			},
		)
	}

	return &resp, nil
}

// queryOpen handles MsgQuery.Open query.
func queryOpen(deps *std.Deps, env stdTypes.Env) (*types.QueryOpenResponse, error) {
	ids := make([]uint64, 0)

	err := state.IterateVotings(deps.Storage, func(voting state.Voting) (stop bool) {
		if !voting.IsClosed(env.Block.Time) {
			ids = append(ids, voting.ID)
		}
		return false
	})
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &types.QueryOpenResponse{
		Ids: ids,
	}, nil
}

// queryReleaseStats handles MsgQuery.ReleaseStats query.
func queryReleaseStats(deps *std.Deps) (*types.QueryReleaseStatsResponse, error) {
	stats, err := state.GetReleaseStats(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &types.QueryReleaseStatsResponse{
		ReleaseStats: stats,
	}, nil
}

// queryWithdrawStats handles MsgQuery.WithdrawStats query.
func queryWithdrawStats(deps *std.Deps) (*types.QueryWithdrawStatsResponse, error) {
	stats, err := state.GetWithdrawStats(deps.Storage)
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &types.QueryWithdrawStatsResponse{
		WithdrawStats: stats,
	}, nil
}

// queryIBCStats handles MsgQuery.IBCStats query.
func queryIBCStats(deps *std.Deps, req types.QueryIBCStatsRequest) (*types.QueryIBCStatsResponse, error) {
	var stats []state.IBCStats
	err := state.IterateIBCStats(deps.Storage, req.From, func(ibcStats state.IBCStats) (stop bool) {
		stats = append(stats, ibcStats)
		return false
	})
	if err != nil {
		return nil, types.NewErrInternal(err.Error())
	}

	return &types.QueryIBCStatsResponse{
		Stats: stats,
	}, nil
}

// queryAPIVerifySecp256k1Signature defines MsgQuery.APIVerifySecp256k1Signature query.
func queryAPIVerifySecp256k1Signature(deps *std.Deps, req types.QueryAPIVerifySecp256k1SignatureRequest) (*types.QueryAPIVerifySecp256k1SignatureResponse, error) {
	ok, err := deps.Api.VerifySecp256k1Signature(req.Hash, req.Signature, req.PubKey)
	if err != nil {
		return nil, types.NewErrInvalidRequest(err.Error())
	}

	return &types.QueryAPIVerifySecp256k1SignatureResponse{
		Valid: ok,
	}, nil
}

// queryAPIRecoverSecp256k1PubKey defines MsgQuery.APIRecoverSecp256k1PubKey query.
func queryAPIRecoverSecp256k1PubKey(deps *std.Deps, req types.QueryAPIRecoverSecp256k1PubKeyRequest) (*types.QueryAPIRecoverSecp256k1PubKeyResponse, error) {
	pubKey, err := deps.Api.RecoverSecp256k1PubKey(req.Hash, req.Signature, req.RecoveryParam)
	if err != nil {
		return nil, types.NewErrInvalidRequest(err.Error())
	}

	return &types.QueryAPIRecoverSecp256k1PubKeyResponse{
		PubKey: pubKey,
	}, nil
}

// queryAPIVerifyEd25519Signature defines MsgQuery.VerifyEd25519Signature query.
func queryAPIVerifyEd25519Signature(deps *std.Deps, req types.QueryAPIVerifyEd25519SignatureRequest) (*types.QueryAPIVerifyEd25519SignatureResponse, error) {
	ok, err := deps.Api.VerifyEd25519Signature(req.Message, req.Signature, req.PubKey)
	if err != nil {
		return nil, types.NewErrInvalidRequest(err.Error())
	}

	return &types.QueryAPIVerifyEd25519SignatureResponse{
		Valid: ok,
	}, nil
}

// queryAPIVerifyEd25519Signatures defines MsgQuery.VerifyEd25519Signatures query.
func queryAPIVerifyEd25519Signatures(deps *std.Deps, req types.QueryAPIVerifyEd25519SignaturesRequest) (*types.QueryAPIVerifyEd25519SignaturesResponse, error) {
	ok, err := deps.Api.VerifyEd25519Signatures(req.Messages, req.Signatures, req.PubKeys)
	if err != nil {
		return nil, types.NewErrInvalidRequest(err.Error())
	}

	return &types.QueryAPIVerifyEd25519SignaturesResponse{
		Valid: ok,
	}, nil
}

// queryCustomCustom defines CustomQuery.Custom query.
func queryCustomCustom(deps *std.Deps, req stdTypes.RawMessage) (*types.CustomCustomResponse, error) {
	reqRaw := stdTypes.QueryRequest{
		Custom: req,
	}
	reqRawBz, err := reqRaw.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("query JSON marshal: " + err.Error())
	}

	resBz, err := deps.Querier.RawQuery(reqRawBz)
	if err != nil {
		return nil, types.NewErrInternal("raw query: " + err.Error())
	}

	return &types.CustomCustomResponse{
		Response: resBz,
	}, nil
}

// queryCustomMetadata defines CustomQuery.Metadata query.
func queryCustomMetadata(deps *std.Deps, env stdTypes.Env, req types.CustomMetadataRequest) (*types.CustomMetadataResponse, error) {
	if req.UseStargateQuery {
		return queryCustomMetadataStargate(deps, env)
	}

	return queryCustomMetadataCustom(deps, env)
}

// queryCustomMetadataCustom returns a contract metadata using Custom plugin query.
func queryCustomMetadataCustom(deps *std.Deps, env stdTypes.Env) (*types.CustomMetadataResponse, error) {
	customReq := archwayCustomTypes.CustomQuery{
		ContractMetadata: &archwayCustomTypes.ContractMetadataRequest{
			ContractAddress: env.Contract.Address,
		},
	}

	customReqBz, err := customReq.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("request JSON marshal: " + err.Error())
	}

	reqRaw := stdTypes.QueryRequest{
		Custom: customReqBz,
	}
	reqRawBz, err := reqRaw.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("query JSON marshal: " + err.Error())
	}

	resBz, err := deps.Querier.RawQuery(reqRawBz)
	if err != nil {
		return nil, types.NewErrInternal("raw query: " + err.Error())
	}

	var res archwayCustomTypes.ContractMetadataResponse
	if err := res.UnmarshalJSON(resBz); err != nil {
		return nil, types.NewErrInternal("raw query result unmarshal: " + err.Error())
	}

	return &types.CustomMetadataResponse{
		ContractMetadataResponse: res,
	}, nil
}

// queryCustomMetadataStargate returns a contract metadata using Stargate query.
func queryCustomMetadataStargate(deps *std.Deps, env stdTypes.Env) (*types.CustomMetadataResponse, error) {
	const queryPath = "/archway.rewards.v1.Query/ContractMetadata"

	queryData := archwayProtoTypes.QueryContractMetadataRequest{
		ContractAddress: env.Contract.Address,
	}
	queryDataBz, err := queryData.Marshal()
	if err != nil {
		return nil, types.NewErrInternal("query data Proto marshal: " + err.Error())
	}

	query := stdTypes.QueryRequest{
		Stargate: &stdTypes.StargateQuery{
			Path: queryPath,
			Data: queryDataBz,
		},
	}
	queryBz, err := query.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("query JSON marshal: " + err.Error())
	}

	resBz, err := deps.Querier.RawQuery(queryBz)
	if err != nil {
		return nil, types.NewErrInternal("raw query: " + err.Error())
	}

	var res archwayProtoTypes.QueryContractMetadataResponse
	if err := res.Unmarshal(resBz); err != nil {
		return nil, types.NewErrInternal("raw query result Proto unmarshal: " + err.Error())
	}

	return &types.CustomMetadataResponse{
		ContractMetadataResponse: archwayCustomTypes.ContractMetadataResponse{
			OwnerAddress:   res.Metadata.OwnerAddress,
			RewardsAddress: res.Metadata.RewardsAddress,
		},
	}, nil
}

// queryCustomRewardsRecords defines CustomQuery.RewardsRecords query.
func queryCustomRewardsRecords(deps *std.Deps, env stdTypes.Env, req types.CustomRewardsRecordsRequest) (*types.CustomRewardsRecordsResponse, error) {
	customReq := archwayCustomTypes.CustomQuery{
		RewardsRecords: &archwayCustomTypes.RewardsRecordsRequest{
			RewardsAddress: env.Contract.Address,
			Pagination:     req.Pagination,
		},
	}

	customReqBz, err := customReq.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("request JSON marshal: " + err.Error())
	}

	reqRaw := stdTypes.QueryRequest{
		Custom: customReqBz,
	}
	reqRawBz, err := reqRaw.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("query JSON marshal: " + err.Error())
	}

	resBz, err := deps.Querier.RawQuery(reqRawBz)
	if err != nil {
		return nil, types.NewErrInternal("raw query: " + err.Error())
	}

	var res archwayCustomTypes.RewardsRecordsResponse
	if err := res.UnmarshalJSON(resBz); err != nil {
		return nil, types.NewErrInternal("raw query result unmarshal: " + err.Error())
	}

	return &types.CustomRewardsRecordsResponse{
		RewardsRecordsResponse: res,
	}, nil
}

func queryCustomGovVote(deps *std.Deps, env stdTypes.Env, req types.CustomGovVoteRequest) (*json.RawMessage, error) {
	customReq := archwayCustomTypes.CustomQuery{
		GovVote: &archwayCustomTypes.GovVoteRequest{
			ProposalID: req.ProposalID,
			Voter:      req.Voter,
		},
	}

	customReqBz, err := customReq.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("request JSON marshal: " + err.Error())
	}

	reqRaw := stdTypes.QueryRequest{
		Custom: customReqBz,
	}
	reqRawBz, err := reqRaw.MarshalJSON()
	if err != nil {
		return nil, types.NewErrInternal("query JSON marshal: " + err.Error())
	}

	resBz, err := deps.Querier.RawQuery(reqRawBz)
	if err != nil {
		return nil, types.NewErrInternal("raw query: " + err.Error())
	}

	asRawMsg := json.RawMessage(resBz)
	return &asRawMsg, nil
}
