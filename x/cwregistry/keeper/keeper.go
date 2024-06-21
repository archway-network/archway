package keeper

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/log"

	"github.com/archway-network/archway/internal/collcompat"
	"github.com/archway-network/archway/x/cwregistry/types"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		Codec      codec.BinaryCodec
		storeKey   storetypes.StoreKey
		sudoKeeper types.WasmKeeper
		logger     log.Logger
		Schema     collections.Schema

		// CodeMetadata key: CodeMetadataKeyPrefix + codeID | value: CodeMetadata
		CodeMetadata collections.Map[uint32, types.CodeMetadata]
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	sudoKeeper types.WasmKeeper,
	logger log.Logger,
) Keeper {
	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))

	k := Keeper{
		Codec:      cdc,
		storeKey:   storeKey,
		sudoKeeper: sudoKeeper,
		logger:     logger.With("module", "x/"+types.ModuleName),
		CodeMetadata: collections.NewMap(
			sb,
			types.CodeMetadataKeyPrefix,
			"code_metadatas",
			collections.Uint32Key,
			collcompat.ProtoValue[types.CodeMetadata](cdc),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema
	return k
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return k.logger
}

// SetWasmKeeper sets the given wasm keeper.
// NOTE: Only for testing purposes
func (k *Keeper) SetWasmKeeper(wk types.WasmKeeper) {
	k.sudoKeeper = wk
}
