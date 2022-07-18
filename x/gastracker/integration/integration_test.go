package integration

import (
	voterTypes "github.com/CosmWasm/cosmwasm-go/example/voter/src/types"
	cwMath "github.com/CosmWasm/cosmwasm-go/std/math"
	cwSdkTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/x/gastracker"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"testing"
)

func TestRewardsCollection(t *testing.T) {
	chain := e2eTesting.NewTestChain(t, 1)
	owner := chain.GetAccount(0)
	addr := owner.Address.String()
	id := chain.UploadContract(owner, "../../../e2e/contracts/voter.wasm", wasmtypes.DefaultUploadAccess)
	chain.InstantiateContract(owner, id, "", "voter", nil, voterTypes.MsgInstantiate{Params: voterTypes.Params{
		OwnerAddr: addr,
		NewVotingCost: cwSdkTypes.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: cwMath.NewUint128FromUint64(100),
		}.String(),
		VoteCost: cwSdkTypes.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: cwMath.NewUint128FromUint64(100),
		}.String(),
		IBCSendTimeout: 30000000000, // 30s‰
	}})

	chain.NextBlock(0)

	// now what we do next is that
	// we check the balance of gastracker
	balance := chain.GetBalance(authtypes.NewEmptyModuleAccount(gastracker.ModuleName).GetAddress())
	t.Logf("%s", balance)
}
