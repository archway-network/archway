package action

import (
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/NibiruChain/collections"

	"github.com/archway-network/archway/app"
	"github.com/archway-network/archway/x/common/asset"
	"github.com/archway-network/archway/x/common/testutil/action"
	"github.com/archway-network/archway/x/oracle/types"
)

func SetOraclePrice(pair asset.Pair, price math.LegacyDec) action.Action {
	return &setPairPrice{
		Pair:  pair,
		Price: price,
	}
}

type setPairPrice struct {
	Pair  asset.Pair
	Price math.LegacyDec
}

func (s setPairPrice) Do(app *app.ArchwayApp, ctx sdk.Context) (sdk.Context, error) {
	app.Keepers.OracleKeeper.SetPrice(ctx, s.Pair, s.Price)

	return ctx, nil
}

func InsertOraclePriceSnapshot(pair asset.Pair, time time.Time, price math.LegacyDec) action.Action {
	return &insertOraclePriceSnapshot{
		Pair:  pair,
		Time:  time,
		Price: price,
	}
}

type insertOraclePriceSnapshot struct {
	Pair  asset.Pair
	Time  time.Time
	Price math.LegacyDec
}

func (s insertOraclePriceSnapshot) Do(app *app.ArchwayApp, ctx sdk.Context) (sdk.Context, error) {
	app.Keepers.OracleKeeper.PriceSnapshots.Insert(ctx, collections.Join(s.Pair, s.Time), types.PriceSnapshot{
		Pair:        s.Pair,
		Price:       s.Price,
		TimestampMs: s.Time.UnixMilli(),
	})

	return ctx, nil
}
