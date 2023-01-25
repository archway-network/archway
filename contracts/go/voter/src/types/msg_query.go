package types

import (
	"github.com/CosmWasm/cosmwasm-go/std"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"

	archwayCustomTypes "github.com/archway-network/voter/src/pkg/archway/custom"
	"github.com/archway-network/voter/src/state"
)

// MsgQuery is handled by the Query entrypoint.
type MsgQuery struct {
	// Params returns the current contract parameters.
	Params *struct{} `json:",omitempty"`
	// Voting returns a voting meta.
	Voting *QueryVotingRequest `json:",omitempty"`
	// Results returns a voting summary.
	Tally *QueryTallyRequest `json:",omitempty"`
	// Open returns all open voting IDs.
	Open *struct{} `json:",omitempty"`
	// ReleaseStats returns the current Release operations stats.
	ReleaseStats *struct{} `json:",omitempty"`
	// IBCStats returns sent IBC packets stats for a given senderAddress.
	IBCStats *QueryIBCStatsRequest `json:",omitempty"`
	// WithdrawStats returns the current Withdraw operations stats.
	WithdrawStats *struct{} `json:",omitempty"`

	// APIVerifySecp256k1Signature calls api.VerifySecp256k1Signature and returns verification result.
	APIVerifySecp256k1Signature *QueryAPIVerifySecp256k1SignatureRequest `json:",omitempty"`
	// APIRecoverSecp256k1PubKey calls api.RecoverSecp256k1PubKey and returns pubKey recovered.
	APIRecoverSecp256k1PubKey *QueryAPIRecoverSecp256k1PubKeyRequest `json:",omitempty"`
	// APIVerifyEd25519Signature calls api.VerifyEd25519Signature and returns verification result.
	APIVerifyEd25519Signature *QueryAPIVerifyEd25519SignatureRequest `json:",omitempty"`
	// APIVerifyEd25519Signatures calls api.VerifyEd25519Signatures and returns verification result.
	APIVerifyEd25519Signatures *QueryAPIVerifyEd25519SignaturesRequest `json:",omitempty"`

	// CustomCustom calls WASM bindings custom query.
	CustomCustom stdTypes.RawMessage
	// CustomMetadata calls WASM bindings Metadata query.
	CustomMetadata *CustomMetadataRequest `json:",omitempty"`
	// CustomRewardsRecords calls WASM bindings RewardsRecords query (using contractAddress as the rewardsAddress).
	CustomRewardsRecords *CustomRewardsRecordsRequest `json:",omitempty"`
	// CustomGovVote calls WASM bindings GovVote.
	CustomGovVoteRequest *CustomGovVoteRequest `json:",omitempty"`
}

type CustomGovVoteRequest struct {
	ProposalID uint64 `json:",omitempty"`
	Voter      string `json:",omitempty"`
}

// QueryParamsResponse defines MsgQuery.Params response.
type QueryParamsResponse struct {
	Params
}

type (
	// QueryVotingRequest defines MsgQuery.Voting request.
	QueryVotingRequest struct {
		ID uint64
	}

	// QueryVotingResponse defines MsgQuery.Voting response.
	QueryVotingResponse struct {
		state.Voting
	}
)

type (
	// QueryTallyRequest defines MsgQuery.Tally request.
	QueryTallyRequest struct {
		ID uint64
	}

	// QueryTallyResponse defines MsgQuery.Tally response.
	QueryTallyResponse struct {
		Open  bool
		Votes []VoteTally
	}

	VoteTally struct {
		Option   string
		TotalYes uint32
		TotalNo  uint32
	}
)

// QueryOpenResponse defines MsgQuery.Open response.
type QueryOpenResponse struct {
	Ids []uint64
}

// QueryReleaseStatsResponse defines MsgQuery.ReleaseStats response.
type QueryReleaseStatsResponse struct {
	state.ReleaseStats
}

// QueryWithdrawStatsResponse defines MsgQuery.WithdrawStats response.
type QueryWithdrawStatsResponse struct {
	state.WithdrawStats
}

type (
	// QueryIBCStatsRequest defines MsgQuery.IBCStats request.
	QueryIBCStatsRequest struct {
		From string
	}

	// QueryIBCStatsResponse defines MsgQuery.IBCStats response.
	QueryIBCStatsResponse struct {
		Stats []state.IBCStats
	}
)

type (
	// QueryAPIVerifySecp256k1SignatureRequest defines MsgQuery.APIVerifySecp256k1Signature request.
	QueryAPIVerifySecp256k1SignatureRequest struct {
		Hash      []byte
		Signature []byte
		PubKey    []byte
	}

	// QueryAPIVerifySecp256k1SignatureResponse defines MsgQuery.APIVerifySecp256k1Signature response.
	QueryAPIVerifySecp256k1SignatureResponse struct {
		Valid bool
	}
)

type (
	// QueryAPIRecoverSecp256k1PubKeyRequest defines MsgQuery.APIRecoverSecp256k1PubKey request.
	QueryAPIRecoverSecp256k1PubKeyRequest struct {
		Hash          []byte
		Signature     []byte
		RecoveryParam std.Secp256k1RecoveryParam
	}

	// QueryAPIRecoverSecp256k1PubKeyResponse defines MsgQuery.APIRecoverSecp256k1PubKey response.
	QueryAPIRecoverSecp256k1PubKeyResponse struct {
		PubKey []byte
	}
)

type (
	// QueryAPIVerifyEd25519SignatureRequest defines MsgQuery.APIVerifyEd25519Signature request.
	QueryAPIVerifyEd25519SignatureRequest struct {
		Message   []byte
		Signature []byte
		PubKey    []byte
	}

	// QueryAPIVerifyEd25519SignatureResponse defines MsgQuery.APIVerifyEd25519Signature response.
	QueryAPIVerifyEd25519SignatureResponse struct {
		Valid bool
	}
)

type (
	// QueryAPIVerifyEd25519SignaturesRequest defines MsgQuery.APIVerifyEd25519Signatures request.
	QueryAPIVerifyEd25519SignaturesRequest struct {
		Messages   [][]byte
		Signatures [][]byte
		PubKeys    [][]byte
	}

	// QueryAPIVerifyEd25519SignaturesResponse defines MsgQuery.APIVerifyEd25519Signatures response.
	QueryAPIVerifyEd25519SignaturesResponse struct {
		Valid bool
	}
)

// CustomCustomResponse defines MsgQuery.CustomCustom response.
type CustomCustomResponse struct {
	Response stdTypes.RawMessage
}

type (
	// CustomMetadataRequest defines MsgQuery.CustomMetadata request.
	CustomMetadataRequest struct {
		// UseStargateQuery is a flag to indicate that the Stargate query should be used (Custom query otherwise).
		UseStargateQuery bool
	}

	// CustomMetadataResponse defines MsgQuery.CustomMetadata response.
	CustomMetadataResponse struct {
		archwayCustomTypes.ContractMetadataResponse
	}
)

type (
	// CustomRewardsRecordsRequest defines MsgQuery.CustomRewardsRecords request.
	CustomRewardsRecordsRequest struct {
		Pagination *archwayCustomTypes.PageRequest
	}

	// CustomRewardsRecordsResponse defines MsgQuery.CustomRewardsRecords response.
	CustomRewardsRecordsResponse struct {
		archwayCustomTypes.RewardsRecordsResponse
	}
)
