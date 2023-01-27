package wasmbinding_test

import (
	"encoding/json"
	"testing"

	cwMath "github.com/CosmWasm/cosmwasm-go/std/math"
	cwSdkTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	wasmdTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	voterTypes "github.com/archway-network/voter/src/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkGov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
)

func TestGovQuerier(t *testing.T) {
	// we create a vote which only contains the address of account 1
	// and we check if the contract can see the vote and match the result
	chain := e2eTesting.NewTestChain(t, 1)
	chain.GetApp().GovKeeper.SetVote(chain.GetContext(), sdkGov.Vote{
		ProposalId: 1,
		Voter:      chain.GetAccount(1).Address.String(),
		Option:     0,
		Options:    nil,
	})
	acc := chain.GetAccount(0)
	codeID := chain.UploadContract(acc, "../contracts/go/voter/code.wasm", wasmdTypes.DefaultUploadAccess)
	init := voterTypes.Params{
		OwnerAddr: acc.Address.String(),
		NewVotingCost: cwSdkTypes.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: cwMath.NewUint128FromUint64(10),
		}.String(),
		VoteCost: cwSdkTypes.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: cwMath.NewUint128FromUint64(10),
		}.String(),
		IBCSendTimeout: 30000000000, // 30s‰
	}
	contractAddr, _ := chain.InstantiateContract(acc, codeID, acc.Address.String(), "voter", nil, voterTypes.MsgInstantiate{Params: init})

	queryMsg := &voterTypes.MsgQuery{CustomGovVoteRequest: &voterTypes.CustomGovVoteRequest{
		ProposalID: 1,
		Voter:      chain.GetAccount(1).Address.String(),
	}}

	queryMsgBytes, err := json.Marshal(queryMsg)
	require.NoError(t, err)

	resp, err := chain.SmartQueryContract(contractAddr, true, json.RawMessage(queryMsgBytes))
	require.NoError(t, err)

	require.Contains(t, string(resp), chain.GetAccount(1).Address.String())
}
