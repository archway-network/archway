package state

import (
	"errors"

	"github.com/CosmWasm/cosmwasm-go/std"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"

	"github.com/archway-network/voter/src/pkg"
)

// WithdrawStatsKey is the storage key for storing WithdrawStats.
var WithdrawStatsKey = []byte("WithdrawStats")

// WithdrawStats keeps Release execute stats.
type WithdrawStats struct {
	// Count is a total number of successful withdrawal operations.
	Count uint64
	// TotalAmount is a total amount of funds distributed.
	TotalAmount []stdTypes.Coin
	// TotalRecordsUsed is a total amount of RewardsRecords used in withdrawal operations.
	TotalRecordsUsed uint64
}

// AddWithdraw increments stats by a single withdraw operation.
func (s *WithdrawStats) AddWithdraw(amount []stdTypes.Coin, recordsUsed uint64) {
	s.Count++
	s.TotalAmount = pkg.AddCoins(s.TotalAmount, amount...)
	s.TotalRecordsUsed += recordsUsed
}

// GetWithdrawStats returns WithdrawStats state.
func GetWithdrawStats(storage std.Storage) (withdrawStats WithdrawStats, retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("withdrawStats state get: " + retErr.Error())
		}
	}()

	bz := storage.Get(WithdrawStatsKey)
	if bz == nil {
		return
	}

	if err := withdrawStats.UnmarshalJSON(bz); err != nil {
		retErr = errors.New("object JSON unmarshal")
		return
	}

	return
}

// SetWithdrawStats sets ReleaseStats state.
func SetWithdrawStats(storage std.Storage, withdrawStats WithdrawStats) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("withdrawStats state set: " + retErr.Error())
		}
	}()

	bz, err := withdrawStats.MarshalJSON()
	if err != nil {
		retErr = errors.New("object JSON marshal: " + err.Error())
		return
	}

	storage.Set(WithdrawStatsKey, bz)

	return
}
