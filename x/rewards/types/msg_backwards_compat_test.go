package types

import (
	"encoding/hex"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

func TestBackwardsCompat(t *testing.T) {
	// defines the pre ux improvements contract metadata protobuf bytes
	const preDevUXContractMetadata = "0a08636f6e747261637412056f776e65721a0772657761726473"
	// defines the concrete metadata struct represented by the above protobuf bytes
	wantContractMetadata := &ContractMetadata{
		ContractAddress:  "contract",
		OwnerAddress:     "owner",
		RewardsAddress:   "rewards",
		WithdrawToWallet: false, // this was not present before, but we want it to default to false if not present in bytes.
	}
	// we assert that we can decode the pre ux improvements contract metadata
	// into a new contract metadata struct.
	md := &ContractMetadata{}
	protoBytes, err := hex.DecodeString(preDevUXContractMetadata)
	require.NoError(t, err)
	require.NoError(t, proto.Unmarshal(protoBytes, md))
	require.Equal(t, wantContractMetadata, md)
}
