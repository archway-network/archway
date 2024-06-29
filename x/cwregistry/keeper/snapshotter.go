package keeper

import (
	"fmt"
	"io"
	"math"

	"cosmossdk.io/log"
	snapshot "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/archway-network/archway/x/cwregistry/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ snapshot.ExtensionSnapshotter = &CwRegistrySnapshotter{}

const SnapshotFormat = 1

type CwRegistrySnapshotter struct {
	registryKeeper *Keeper
	cms            storetypes.MultiStore
}

func NewCwRegistrySnapshotter(k *Keeper, cms storetypes.MultiStore) *CwRegistrySnapshotter {
	return &CwRegistrySnapshotter{
		registryKeeper: k,
		cms:            cms,
	}
}

func (s *CwRegistrySnapshotter) SnapshotName() string {
	return types.ModuleName
}

func (s *CwRegistrySnapshotter) SnapshotFormat() uint32 {
	return SnapshotFormat
}

func (s *CwRegistrySnapshotter) SupportedFormats() []uint32 {
	return []uint32{SnapshotFormat}
}

func (s *CwRegistrySnapshotter) SnapshotExtension(height uint64, payloadWriter snapshot.ExtensionPayloadWriter) error {
	cacheMS, err := s.cms.CacheMultiStoreWithVersion(int64(height))
	if err != nil {
		return err
	}
	ctx := sdk.NewContext(cacheMS, tmproto.Header{}, false, log.NewNopLogger())
	codeMetadata, err := s.registryKeeper.GetAllCodeMetadata(ctx)
	if err != nil {
		return err
	}
	for _, metadata := range codeMetadata {
		schema, err := s.registryKeeper.GetSchema(metadata.CodeId)
		if err != nil {
			return err
		}
		compressedSchema, err := ioutils.GzipIt([]byte(schema))
		if err != nil {
			return err
		}
		snapshotPayload := types.NewSnapshot(metadata.CodeId, compressedSchema)
		payloadData, err := snapshotPayload.Marshal()
		if err != nil {
			return err
		}
		err = payloadWriter(payloadData)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *CwRegistrySnapshotter) RestoreExtension(height uint64, format uint32, payloadReader snapshot.ExtensionPayloadReader) error {
	if format != SnapshotFormat {
		return snapshot.ErrUnknownFormat
	}
	for {
		payload, err := payloadReader()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return fmt.Errorf("cannot read blob from the payload: %w", err)
		}
		var snapshotPayload types.SnapshotPayload
		err = snapshotPayload.Unmarshal(payload)
		if err != nil {
			return err
		}
		if !ioutils.IsGzip(snapshotPayload.Schema) {
			return types.ErrInvalidSnapshotPayload
		}
		schema, err := ioutils.Uncompress(snapshotPayload.Schema, math.MaxInt64)
		if err != nil {
			return err
		}
		err = s.registryKeeper.SetSchema(snapshotPayload.CodeId, string(schema))
		if err != nil {
			return err
		}
	}
}
