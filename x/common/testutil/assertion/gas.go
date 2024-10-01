package assertion

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/app"
	"github.com/archway-network/archway/x/common/testutil/action"
)

type gasConsumedShouldBe struct {
	gasConsumed uint64
}

func (g gasConsumedShouldBe) Do(_ *app.ArchwayApp, ctx sdk.Context) (sdk.Context, error) {
	gasUsed := ctx.GasMeter().GasConsumed()
	if g.gasConsumed != gasUsed {
		return ctx, fmt.Errorf("gas consumed should be %d, but got %d", g.gasConsumed, gasUsed)
	}

	return ctx, nil
}

func GasConsumedShouldBe(gasConsumed uint64) action.Action {
	return &gasConsumedShouldBe{gasConsumed: gasConsumed}
}
