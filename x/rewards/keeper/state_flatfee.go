package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storeTypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/x/rewards/types"
)

// FlatFeeState provides access to the types.FlatFee objects storage operations.
type FlatFeeState struct {
	stateStore storeTypes.KVStore
	cdc        codec.Codec
	ctx        sdk.Context
}

// SetFee creates or modifies a types.FlatFee object.
func (s FlatFeeState) SetFee(contractAddr sdk.AccAddress, feeCoin sdk.Coin) {
	store := prefix.NewStore(s.stateStore, types.FlatFeePrefix)
	store.Set(
		contractAddr.Bytes(),
		s.cdc.MustMarshal(&feeCoin),
	)
}

// GetFee returns a types.FlatFee object by contract address.
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

// RemoveFee deletes a types.FlatFee object.
func (s FlatFeeState) RemoveFee(contractAddr sdk.AccAddress) {
	store := prefix.NewStore(s.stateStore, types.FlatFeePrefix)
	store.Delete(contractAddr.Bytes())
}

// Import initializes state from the genesis flat fees data.
func (s FlatFeeState) Import(flatFees []types.FlatFee) {
	for _, flatFee := range flatFees {
		contractAddr := flatFee.MustGetContractAddress()
		fee := flatFee.GetFlatFee()
		s.SetFee(contractAddr, fee)
	}
}

// Export returns the flat fees genesis data for the state.
func (s FlatFeeState) Export() []types.FlatFee {
	store := prefix.NewStore(s.stateStore, types.FlatFeePrefix)

	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var fees = make([]types.FlatFee, 0)
	for ; iterator.Valid(); iterator.Next() {
		var coin sdk.Coin
		contractAddr := sdk.AccAddress(iterator.Key())
		s.cdc.MustUnmarshal(iterator.Value(), &coin)

		fees = append(fees, types.FlatFee{
			ContractAddress: contractAddr.String(),
			FlatFee:         coin,
		})
	}

	return fees
}
