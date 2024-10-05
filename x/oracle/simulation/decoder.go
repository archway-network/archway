package simulation

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	gogotypes "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/collections"

	"github.com/archway-network/archway/x/common"
	"github.com/archway-network/archway/x/oracle/types"
)

// NewDecodeStore returns a decoder function closure that unmarshals the KVPair's
// Value to the corresponding oracle type.
func NewDecodeStore(cdc codec.Codec) func(kvA, kvB kv.Pair) string {
	return func(kvA, kvB kv.Pair) string {
		switch kvA.Key[0] {
		case 1:
			_, a, err := common.DecValueEncoder.Decode(kvA.Value)
			if err != nil {
				panic(err)
			}
			_, b, err := common.DecValueEncoder.Decode(kvB.Value)
			if err != nil {
				panic(err)
			}
			return fmt.Sprintf("%v\n%v", a, b)
		case 2:
			return fmt.Sprintf("%v\n%v", sdk.AccAddress(kvA.Value), sdk.AccAddress(kvB.Value))
		case 3:
			var counterA, counterB gogotypes.UInt64Value
			cdc.MustUnmarshal(kvA.Value, &counterA)
			cdc.MustUnmarshal(kvB.Value, &counterB)
			return fmt.Sprintf("%v\n%v", counterA.Value, counterB.Value)
		case 4:
			var prevoteA, prevoteB types.AggregateExchangeRatePrevote
			cdc.MustUnmarshal(kvA.Value, &prevoteA)
			cdc.MustUnmarshal(kvB.Value, &prevoteB)
			return fmt.Sprintf("%v\n%v", prevoteA, prevoteB)
		case 5:
			var voteA, voteB types.AggregateExchangeRateVote
			cdc.MustUnmarshal(kvA.Value, &voteA)
			cdc.MustUnmarshal(kvB.Value, &voteB)
			return fmt.Sprintf("%v\n%v", voteA, voteB)
		case 6:
			_, a, err := collections.StringKey.Decode(kvA.Key[1:]) // TODO (spekalsg3): or decode json?
			if err != nil {
				panic(err)
			}
			_, b, err := collections.StringKey.Decode(kvB.Key[1:])
			if err != nil {
				panic(err)
			}
			return fmt.Sprintf("%s\n%s", a, b)
		default:
			panic(fmt.Sprintf("invalid oracle key prefix %X", kvA.Key[:1]))
		}
	}
}
