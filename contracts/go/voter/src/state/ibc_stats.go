package state

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/CosmWasm/cosmwasm-go/std"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
)

// IBCStatsKey is the storage key prefix for storing IBCStats.
var IBCStatsKey = []byte("IBCStats")

const (
	IBCPkgSentStatus     IBCSendStatus = "sent"
	IBCPkgAckedStatus    IBCSendStatus = "acked"
	IBCPkgRejectedStatus IBCSendStatus = "rejected"
	IBCPkgTimedOutStatus IBCSendStatus = "timed out"
)

type (
	IBCSendStatus string

	// IBCStats keeps send IBC packets status.
	IBCStats struct {
		// VotingID is a unique voting ID.
		VotingID uint64
		// From is a vote sender address.
		From string
		// Status defines the current IBC send packet status.
		Status IBCSendStatus
		// CreatedAt is a timestamp IBC packet was created at [UNIX time in ns].
		CreatedAt uint64
	}
)

// NewIBCStats creates a new IBCStats with "sent" status.
func NewIBCStats(senderAddr string, votingID uint64, env stdTypes.Env) IBCStats {
	return IBCStats{
		VotingID:  votingID,
		From:      senderAddr,
		Status:    IBCPkgSentStatus,
		CreatedAt: env.Block.Time,
	}
}

// GetIBCStats returns IBCStats state.
func GetIBCStats(storage std.Storage, senderAddr string, votingID uint64) (ibcStats IBCStats, retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("ibcStats state get: " + retErr.Error())
		}
	}()

	key := buildIBCStatsKey(senderAddr, votingID)
	bz := storage.Get(key)
	if bz == nil {
		retErr = errors.New("not found")
		return
	}

	if err := ibcStats.UnmarshalJSON(bz); err != nil {
		retErr = errors.New("object JSON unmarshal")
		return
	}

	return
}

// SetIBCStats sets IBCStats state.
func SetIBCStats(storage std.Storage, ibcStats IBCStats) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("ibcStats state set: " + retErr.Error())
		}
	}()

	bz, err := ibcStats.MarshalJSON()
	if err != nil {
		retErr = errors.New("object JSON marshal: " + err.Error())
		return
	}

	key := buildIBCStatsKey(ibcStats.From, ibcStats.VotingID)
	storage.Set(key, bz)

	return
}

// IterateIBCStats iterates over all stored IBCStats objects by senderAddr.
func IterateIBCStats(storage std.Storage, senderAddr string, fn func(ibcStats IBCStats) (stop bool)) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("range over ibcStats: " + retErr.Error())
		}
	}()

	keyStart := buildIBCStatsPrefix(senderAddr)
	keyEnd := buildIBCStatsKey(senderAddr, math.MaxUint64)
	iter := storage.Range(keyStart, keyEnd, std.Ascending)
	for {
		_, bz, err := iter.Next()
		if err != nil {
			break
		}

		var ibcStats IBCStats
		if err := ibcStats.UnmarshalJSON(bz); err != nil {
			retErr = errors.New("object JSON unmarshal")
			return
		}

		if fn(ibcStats) {
			break
		}
	}

	return
}

// buildIBCStatsPrefix builds IBCStats storage key prefix.
func buildIBCStatsPrefix(senderAddr string) []byte {
	return append(IBCStatsKey, []byte(senderAddr)...)
}

// buildIBCStatsKey builds IBCStats storage key.
func buildIBCStatsKey(senderAddr string, votingID uint64) []byte {
	id := make([]byte, 8)
	binary.LittleEndian.PutUint64(id, votingID)

	return append(buildIBCStatsPrefix(senderAddr), id...)
}
