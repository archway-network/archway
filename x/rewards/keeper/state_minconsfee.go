package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// MinConsFeeState provides access to the Minimum Consensus Fee object storage operations.
type MinConsFeeState struct {
	stateStore storeTypes.KVStore
	cdc        codec.Codec
	ctx        sdk.Context
}

// SetFee creates or modifies the fee coin.
func (s MinConsFeeState) SetFee(feeCoin sdk.DecCoin) {
	s.stateStore.Set(
		types.MinConsFeeKey,
		s.cdc.MustMarshal(&feeCoin),
	)
}

// GetFee returns the fee coin if exists.
func (s MinConsFeeState) GetFee() (sdk.DecCoin, bool) {
	coinBz := s.stateStore.Get(types.MinConsFeeKey)
	if coinBz == nil {
		return sdk.DecCoin{}, false
	}

	var coin sdk.DecCoin
	s.cdc.MustUnmarshal(coinBz, &coin)

	return coin, true
}
