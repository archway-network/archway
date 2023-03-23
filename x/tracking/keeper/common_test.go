package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	e2etesting "github.com/archway-network/archway/e2e/testing"
)

type KeeperTestSuite struct {
	suite.Suite

	chain *e2etesting.TestChain
}

func (s *KeeperTestSuite) SetupTest() {
	s.chain = e2etesting.NewTestChain(s.T(), 1)
}

func TestTrackingKeeper(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
