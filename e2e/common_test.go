package e2e

import (
	"testing"

	cwTypes "github.com/CosmWasm/cosmwasm-go/std/types"
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	suite.Suite

	chainA *e2eTesting.TestChain
	chainB *e2eTesting.TestChain
}

func (s *E2ETestSuite) SetupTest() {
	s.chainA = e2eTesting.NewTestChain(s.T(), 1)
	s.chainB = e2eTesting.NewTestChain(s.T(), 2)
}

// CosmWasmCoinsToSDK converts CosmWasm SDK coins to the Cosmos SDK coins.
func (s *E2ETestSuite) CosmWasmCoinsToSDK(cwCoins ...cwTypes.Coin) sdk.Coins {
	coins := sdk.NewCoins()
	for _, cwCoin := range cwCoins {
		amt, ok := sdk.NewIntFromString(cwCoin.Amount.String())
		s.Require().True(ok)

		coins = coins.Add(sdk.NewCoin(cwCoin.Denom, amt))
	}

	return coins
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
