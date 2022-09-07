package state

import (
	"errors"

	"github.com/CosmWasm/cosmwasm-go/std"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"

	"github.com/archway-network/voter/src/pkg"
)

// ReleaseStatsKey is the storage key for storing ReleaseStats.
var ReleaseStatsKey = []byte("ReleaseStats")

// ReleaseStats keeps Release execute stats.
type ReleaseStats struct {
	// Count is a total number of successful contract funds releases.
	Count uint64
	// TotalAmount is a total amount of funds raised by the creator.
	TotalAmount []stdTypes.Coin
}

// AddRelease increments stats by a single release operation.
func (s *ReleaseStats) AddRelease(amount []stdTypes.Coin) {
	s.Count++
	s.TotalAmount = pkg.AddCoins(s.TotalAmount, amount...)
}

// GetReleaseStats returns ReleaseStats state.
func GetReleaseStats(storage std.Storage) (releaseStats ReleaseStats, retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("releaseStats state get: " + retErr.Error())
		}
	}()

	bz := storage.Get(ReleaseStatsKey)
	if bz == nil {
		return
	}

	if err := releaseStats.UnmarshalJSON(bz); err != nil {
		retErr = errors.New("object JSON unmarshal")
		return
	}

	return
}

// SetReleaseStats sets ReleaseStats state.
func SetReleaseStats(storage std.Storage, releaseStats ReleaseStats) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("releaseStats state set: " + retErr.Error())
		}
	}()

	bz, err := releaseStats.MarshalJSON()
	if err != nil {
		retErr = errors.New("object JSON marshal: " + err.Error())
		return
	}

	storage.Set(ReleaseStatsKey, bz)

	return
}
