package cli_test

import (
	"fmt"
	"testing"

	"github.com/archway-network/archway/x/oracle/client/cli"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/stretchr/testify/require"
)

func TestAddGenesisPricefeederDelegation(t *testing.T) {
	tests := []struct {
		name        string
		validator   string
		pricefeeder string

		expectErr bool
	}{
		{
			name:        "valid",
			validator:   "cosmosvaloper1lg6qclqn8fayp7t7rwxha6hgfhawxm5eh4ued7",
			pricefeeder: "cosmos18lmsapqp03fnvlf6436khg0d9gzhgrfrkcyr3a",
			expectErr:   false,
		},
		{
			name:        "invalid pricefeeder",
			validator:   "cosmosvaloper1lg6qclqn8fayp7t7rwxha6hgfhawxm5eh4ued7",
			pricefeeder: "cosmos1foobar",
			expectErr:   true,
		},
		{
			name:        "empty pricefeeder",
			validator:   "cosmosvaloper1lg6qclqn8fayp7t7rwxha6hgfhawxm5eh4ued7",
			pricefeeder: "",
			expectErr:   true,
		},
		{
			name:        "invalid validator",
			validator:   "cosmosvaloper1foobar",
			pricefeeder: "cosmos18lmsapqp03fnvlf6436khg0d9gzhgrfrkcyr3a",
			expectErr:   true,
		},
		{
			name:        "empty validator",
			validator:   "",
			pricefeeder: "cosmos18lmsapqp03fnvlf6436khg0d9gzhgrfrkcyr3a",
			expectErr:   true,
		},
	}

	for i, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			chain := e2eTesting.NewTestChain(t, i)
			ctx := chain.SetupClientCtx()
			cmd := cli.AddGenesisPricefeederDelegationCmd(t.TempDir())
			cmd.SetArgs([]string{
				fmt.Sprintf("--%s=%s", cli.FlagValidator, tc.validator),
				fmt.Sprintf("--%s=%s", cli.FlagPricefeeder, tc.pricefeeder),
				fmt.Sprintf("--%s=home", flags.FlagHome),
			})

			if tc.expectErr {
				require.Error(t, cmd.ExecuteContext(ctx))
			} else {
				require.NoError(t, cmd.ExecuteContext(ctx))
			}
		})
	}
}
