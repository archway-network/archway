package wasmbinding_test

import (
	"testing"

	wasmVmTypes "github.com/CosmWasm/wasmvm/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/assert"

	e2eTesting "github.com/archway-network/archway/e2e/testing"
	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/wasmbinding"
)

// TestWASMBindingPlugins tests common failure scenarios for custom querier and msg handler plugins.
// Happy paths are tested in the integration tests.
func TestWASMBindingPlugins(t *testing.T) {
	// Setup
	chain := e2eTesting.NewTestChain(t, 1)
	mockMessenger := testutils.NewMockMessenger()
	mockContractAddr := e2eTesting.GenContractAddresses(1)[0]
	ctx := chain.GetContext()

	// Create custom plugins
	rewardsKeeper := chain.GetApp().RewardsKeeper
	msgPlugin, queryPlugin := wasmbinding.BuildWasmMsgDecorator(rewardsKeeper), wasmbinding.BuildWasmQueryPlugin(rewardsKeeper)

	// Querier tests
	t.Run("Querier failure", func(t *testing.T) {
		t.Run("Invalid JSON request", func(t *testing.T) {
			_, err := queryPlugin.Custom(ctx, []byte("invalid"))
			assert.ErrorIs(t, err, sdkErrors.ErrInvalidRequest)
		})

		t.Run("Invalid request (not one of)", func(t *testing.T) {
			queryBz := []byte("{}")

			_, err := queryPlugin.Custom(ctx, queryBz)
			assert.ErrorIs(t, err, sdkErrors.ErrInvalidRequest)
		})
	})

	// Msg handler tests
	t.Run("MsgHandler failure", func(t *testing.T) {
		t.Run("Invalid JSON request", func(t *testing.T) {
			msg := wasmVmTypes.CosmosMsg{
				Custom: []byte("invalid"),
			}
			_, _, err := msgPlugin(mockMessenger).DispatchMsg(ctx, mockContractAddr, "", msg)
			assert.ErrorIs(t, err, sdkErrors.ErrInvalidRequest)
		})

		t.Run("Invalid request (not one of)", func(t *testing.T) {
			msg := wasmVmTypes.CosmosMsg{
				Custom: []byte("{}"),
			}
			_, _, err := msgPlugin(mockMessenger).DispatchMsg(ctx, mockContractAddr, "", msg)
			assert.ErrorIs(t, err, sdkErrors.ErrInvalidRequest)
		})
	})

	t.Run("MsgHandler OK", func(t *testing.T) {
		t.Run("No-op (non-custom msg)", func(t *testing.T) {
			msg := wasmVmTypes.CosmosMsg{}
			_, _, err := msgPlugin(mockMessenger).DispatchMsg(ctx, mockContractAddr, "", msg)
			assert.NoError(t, err)
		})
	})
}
