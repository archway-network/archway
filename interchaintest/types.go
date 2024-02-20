package interchaintest

type InterchainAccountAccountQueryResponse struct {
	InterchainAccountAddress string `json:"interchain_account_address"`
}

type QueryMsg struct {
	DumpState *struct{} `json:"dump_state"`
}

type QueryVoteResponse struct {
	ProposalID string       `json:"proposal_id"`
	Voter      string       `json:"voter"`
	Options    []VoteOption `json:"options"`
	Metadata   string       `json:"metadata"`
}

type VoteOption struct {
	Option string `json:"option"`
	Weight string `json:"weight"`
}
