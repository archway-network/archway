package types

import (
	"strings"

	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
)

// ICAOwnerFromPort returns the owner address from the port
func ICAOwnerFromPort(port string) string {
	return strings.TrimPrefix(port, icatypes.ControllerPortPrefix)
}
