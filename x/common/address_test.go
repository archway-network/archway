package common_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/archway-network/archway/x/common"
	"github.com/archway-network/archway/x/common/testutil"
)

func TestAddress(t *testing.T) {
	require.NotPanics(t, func() {
		_, addrs := testutil.PrivKeyAddressPairs(5)
		strs := common.AddrsToStrings(addrs...)
		addrsOut := common.StringsToAddrs(strs...)
		require.EqualValues(t, addrs, addrsOut)
	})
}
