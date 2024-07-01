package keeper_test

import (
	"testing"

	"github.com/archway-network/archway/pkg/testutils"
	"github.com/archway-network/archway/x/cwregistry/keeper"
	"github.com/archway-network/archway/x/cwregistry/types"
	"github.com/stretchr/testify/require"
)

func TestRegisterCode(t *testing.T) {
	k, ctx := testutils.CWRegistryKeeper(t)
	wasmKeeper := testutils.NewMockContractViewer()
	k.SetWasmKeeper(wasmKeeper)
	msgSrvr := keeper.NewMsgServerImpl(k)
	senderAddr := testutils.AccAddress().String()
	wasmKeeper.AddCodeAdmin(1, senderAddr)

	// TEST: Empty
	msg := types.MsgRegisterCode{}
	_, err := msgSrvr.RegisterCode(ctx, &msg)
	require.Error(t, err)

	// TEST Invalid sender
	msg = types.MsgRegisterCode{
		Sender:   "👻",
		CodeId:   1,
		Schema:   "schema",
		Contacts: []string{"contact1", "contact2"},
	}
	_, err = msgSrvr.RegisterCode(ctx, &msg)
	require.Error(t, err)

	// TEST: Success
	msg = types.MsgRegisterCode{
		Sender:   senderAddr,
		CodeId:   1,
		Schema:   "schema",
		Contacts: []string{"contact1", "contact2"},
	}
	_, err = msgSrvr.RegisterCode(ctx, &msg)
	require.NoError(t, err)
	require.True(t, k.HasCodeMetadata(ctx, 1))
}
