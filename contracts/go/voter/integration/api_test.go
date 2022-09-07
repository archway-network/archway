package integration

import (
	"path/filepath"

	"github.com/CosmWasm/cosmwasm-go/systest"
	mocks "github.com/CosmWasm/wasmvm/api"

	"github.com/archway-network/voter/src/types"
)

func (s *ContractTestSuite) TestAPIValidateAddress() {
	s.Run("Fail: invalid creator address", func() {
		contractPath := filepath.Join("..", ContractWasmFileName)
		creatorAddr := ValidAddr[:len(ValidAddr)-1]

		// Load
		instance := systest.NewInstance(s.T(),
			contractPath,
			15_000_000_000_000,
			nil,
		)

		env := mocks.MockEnv()
		info := mocks.MockInfo(creatorAddr, nil)
		params := s.genParams
		params.OwnerAddr = creatorAddr

		msg := types.MsgInstantiate{
			Params: params,
		}

		// Instantiate
		_, _, err := instance.Instantiate(env, info, msg)
		s.Require().Error(err)
	})
}
