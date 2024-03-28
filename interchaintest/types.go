package interchaintest

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type InterchainAccountAccountQueryResponse struct {
	Address string `json:"address"`
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

type CwErrorParams struct {
	ErrorStoredTime    string   `json:"error_stored_time"`
	SubscriptionFee    sdk.Coin `json:"subscription_fee"`
	SubscriptionPeriod string   `json:"subscription_period"`
}

type CwErrorIsSubscribed struct {
	Subscribed            bool   `json:"subscribed"`
	SubscriptionValidTill string `json:"subscription_valid_till"`
}

type CWErrorResponse struct {
	Errors []CWError `json:"errors"`
}

type CWError struct {
	ModuleName      string `json:"module_name"`
	ErrorCode       uint32 `json:"error_code"`
	ContractAddress string `json:"contract_address"`
	InputPayload    string `json:"input_payload"`
	ErrorMessage    string `json:"error_message"`
}
