//go:build precision_test

// NOTE: this test is run as a separate build tag as it modifies a global
// variable that is used in other tests. In a concurrent environment this
// might cause unforeseen issues, so we isolate it.
package upgrade052_test

import (
	e2eTesting "github.com/archway-network/archway/e2e/testing"
	archway "github.com/archway-network/archway/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrecisionBreakages(t *testing.T) {
	attoPrec := archway.DefaultPowerReduction
	microPrecision := sdk.NewInt(1_000_000)
	stake := func(shouldPass bool, errContains string) {
		chainBad := e2eTesting.NewTestChain(t, 1,
			e2eTesting.WithGenDefaultCoinBalance(attoPrec.MulRaw(1_000_000_000).String()),
			e2eTesting.WithBondAmount(attoPrec.MulRaw(1).String()),
			e2eTesting.WithDefaultFeeAmount(attoPrec.MulRaw(1).String()),
			e2eTesting.WithValidatorsNum(2),
		)
		acc := chainBad.GetAccount(1)

		_, _, _, err := chainBad.SendMsgs(acc, shouldPass, []sdk.Msg{
			&stakingtypes.MsgDelegate{
				DelegatorAddress: acc.Address.String(),
				ValidatorAddress: sdk.ValAddress(chainBad.GetCurrentValSet().Validators[0].Address).String(),
				Amount:           sdk.NewCoin("stake", attoPrec.MulRaw(100_000_000)),
			},
		})
		if !shouldPass {
			require.ErrorContains(t, err, errContains)
		}
	}
	// setup bad power reduction
	sdk.DefaultPowerReduction = microPrecision
	stake(false, "Int64() out of bound")
	// fix power reduction
	sdk.DefaultPowerReduction = attoPrec
	stake(true, "")
}
