package types

import (
	"errors"

	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"

	"github.com/archway-network/voter/src/pkg"
)

const (
	IBCVersion = "voter-1"
)

var (
	IBCAckDataOK      = []byte{0x01}
	IBCAckDataFailure = []byte{0x00}
)

// ValidateIBCChannelParams validates the IBC channel params.
func ValidateIBCChannelParams(ch stdTypes.IBCChannel, checkCounterparty bool) error {
	if err := pkg.ValidateChannelID(ch.Endpoint.ChannelID); err != nil {
		return errors.New("invalid endpoint channelID (" + ch.Endpoint.ChannelID + "): " + err.Error())
	}

	if checkCounterparty {
		if err := pkg.ValidateChannelID(ch.CounterpartyEndpoint.ChannelID); err != nil {
			return errors.New("counterparty: invalid channelID (" + ch.CounterpartyEndpoint.ChannelID + "): " + err.Error())
		}
	}

	if err := ValidateIBCVersion(ch.Version); err != nil {
		return err
	}

	if ch.Order != stdTypes.Unordered {
		return errors.New("ordered channels are not supported")
	}

	return nil
}

// ValidateIBCVersion validates the IBC protocol version.
func ValidateIBCVersion(version string) error {
	if version != IBCVersion {
		return errors.New("invalid IBC version (" + version + "), expected " + IBCVersion)
	}

	return nil
}
