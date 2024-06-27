package keeper

import (
	"path/filepath"

	"cosmossdk.io/collections"
	"cosmossdk.io/log"
	"github.com/archway-network/archway/internal/collcompat"
	"github.com/archway-network/archway/x/cwregistry/types"
	"github.com/cometbft/cometbft/libs/os"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	Keeper struct {
		Codec      codec.BinaryCodec
		storeKey   storetypes.StoreKey
		wasmKeeper types.WasmKeeper
		logger     log.Logger
		dataRoot   string
		Schema     collections.Schema

		// CodeMetadata key: CodeMetadataKeyPrefix + codeID | value: CodeMetadata
		CodeMetadata collections.Map[uint64, types.CodeMetadata]
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	wasmKeeper types.WasmKeeper,
	homePath string,
	logger log.Logger,
) Keeper {
	sb := collections.NewSchemaBuilder(collcompat.NewKVStoreService(storeKey))

	k := Keeper{
		Codec:      cdc,
		storeKey:   storeKey,
		wasmKeeper: wasmKeeper,
		logger:     logger.With("module", "x/"+types.ModuleName),
		dataRoot:   filepath.Join(homePath, "registry"),
		CodeMetadata: collections.NewMap(
			sb,
			types.CodeMetadataKeyPrefix,
			"code_metadata",
			collections.Uint64Key,
			collcompat.ProtoValue[types.CodeMetadata](cdc),
		),
	}
	err := os.EnsureDir(k.dataRoot, 0755)
	if err != nil {
		panic(err)
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
	k.wasmKeeper = wk
}
