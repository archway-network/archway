package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// FlatFeeState
type FlatFeeState struct {
	stateStore storeTypes.KVStore
	cdc        codec.Codec
	ctx        sdk.Context
}

// SetFee
func (s FlatFeeState) SetFee(contractAddr sdk.AccAddress, feeCoin sdk.Coin) {
	store := prefix.NewStore(s.stateStore, types.FlatFeePrefix)
	store.Set(
		contractAddr.Bytes(),
		s.cdc.MustMarshal(&feeCoin),
	)
}

// GetFee
func (s FlatFeeState) GetFee(contractAddr sdk.AccAddress) (sdk.Coin, bool) {
	store := prefix.NewStore(s.stateStore, types.FlatFeePrefix)
	coinBz := store.Get(contractAddr.Bytes())
	if coinBz == nil {
		return sdk.Coin{}, false
	}

	var coin sdk.Coin
	s.cdc.MustUnmarshal(coinBz, &coin)

	return coin, true
}

// RemoveFee
func (s FlatFeeState) RemoveFee(contractAddr sdk.AccAddress) {
	store := prefix.NewStore(s.stateStore, types.FlatFeePrefix)
	store.Delete(contractAddr.Bytes())
}
