package cwgrant

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/collections"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/archway-network/archway/internal/collcompat"
	"github.com/archway-network/archway/x/cwgrant/types"
)

type WasmdKeeper interface {
	HasContractInfo(ctx sdk.Context, contractAddress sdk.AccAddress) bool
	Sudo(ctx sdk.Context, contractAddress sdk.AccAddress, msg []byte) ([]byte, error)
}

type Keeper struct {
	wasmdKeeper WasmdKeeper

	cdc codec.BinaryCodec

	Schema            collections.Schema
	GrantingContracts collections.KeySet[[]byte]
}

func NewKeeper(cdc codec.BinaryCodec, storeKey storetypes.StoreKey, wasmdKeeper WasmdKeeper) Keeper {
	schemaBuilder := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))
	k := Keeper{
		cdc:               cdc,
		wasmdKeeper:       wasmdKeeper,
		GrantingContracts: collections.NewKeySet(schemaBuilder, types.GrantersPrefix, "granting_contracts", collections.BytesKey),
	}
	schema, err := schemaBuilder.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// RegisterAsGranter registers the contract as a granter.
func (k Keeper) RegisterAsGranter(ctx context.Context, granter sdk.AccAddress) error {
	// we want to assess that the granter is a CW contract.
	if !k.wasmdKeeper.HasContractInfo(sdk.UnwrapSDKContext(ctx), granter) {
		return types.ErrNotAContract
	}
	isGranter, err := k.IsGrantingContract(ctx, granter)
	if err != nil {
		return err
	}
	if isGranter {
		return types.ErrAlreadyGranter.Wrapf("address %s", granter.String())
	}
	return k.GrantingContracts.Set(ctx, granter)
}

// IsGrantingContract checks if the provided granter address is one of the registered granting contracts.
func (k Keeper) IsGrantingContract(ctx context.Context, granter sdk.AccAddress) (bool, error) {
	return k.GrantingContracts.Has(ctx, granter)
}

// RequestGrant will signal to the contract that there's a grant request for a set of messages and the fees.
// In case the contract does not accept the grant then an error is returned.
func (k Keeper) RequestGrant(ctx context.Context, grantingContract sdk.AccAddress, txMsgs []sdk.Msg, wantFees sdk.Coins) error {
	msg, err := types.NewSudoMsg(k.cdc, wantFees, txMsgs)
	if err != nil {
		return err
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = k.wasmdKeeper.Sudo(sdk.UnwrapSDKContext(ctx), grantingContract, msgBytes)
	return err
}

// ImportState imports the state, assumes all contracts provided are valid.
func (k Keeper) ImportState(ctx context.Context, state *types.GenesisState) error {
	for i, addrStr := range state.GrantingContracts {
		addr, err := sdk.AccAddressFromBech32(addrStr)
		if err != nil {
			return fmt.Errorf("invalid address at index %d, %s: %w", i, addrStr, err)
		}
		err = k.RegisterAsGranter(ctx, addr)
		if err != nil {
			return err
		}
	}

	return nil
}

// ExportState exports the state.
func (k Keeper) ExportState(ctx context.Context) (*types.GenesisState, error) {
	s := new(types.GenesisState)
	err := k.GrantingContracts.Walk(ctx, nil, func(key []byte) (stop bool, err error) {
		addrStr := sdk.AccAddress(key).String()
		s.GrantingContracts = append(s.GrantingContracts, addrStr)
		return false, nil
	})
	return s, err
}
