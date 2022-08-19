package state

import (
	"encoding/binary"
	"errors"

	"github.com/CosmWasm/cosmwasm-go/std"
)

var (
	// LastVotingIDKey is the storage key for storing last unique Voting ID.
	LastVotingIDKey = []byte("LastVotingID")

	// VotingKey is the storage prefix for storing Voting objects.
	VotingKey = []byte("Voting")
)

type (
	// Voting defines a voting meta and its current progress.
	Voting struct {
		// ID is a unique voting ID.
		ID uint64
		// Name is a voting name (description).
		Name string
		// CreatorAddr is a voting creator address.
		CreatorAddr string
		// StartTime is a block time when a voting has started.
		StartTime uint64
		// EndTime defines when a voting ends (votes are declined after that timestamp and results can be queried).
		EndTime uint64
		// Tallies is the current voting state.
		Tallies []Tally
	}

	Tally struct {
		// Option is a voting option name.
		Option string
		// YesAddrs is an array of addresses voted "yes" (no duplicates allowed).
		YesAddrs []string
		// NoAddrs is an array of addresses voted "no" (no duplicates allowed).
		NoAddrs []string
	}
)

// NewVoting ...
func NewVoting(id uint64, name, creatorAddr string, curTime, votingDur uint64, voteOptions []string) Voting {
	v := Voting{
		ID:          id,
		Name:        name,
		CreatorAddr: creatorAddr,
		StartTime:   curTime,
		EndTime:     curTime + votingDur,
		Tallies:     make([]Tally, 0, len(voteOptions)),
	}

	for _, opt := range voteOptions {
		v.Tallies = append(v.Tallies, Tally{
			Option: opt,
		})
	}

	return v
}

// IsClosed checks if the voting is over.
func (v Voting) IsClosed(curTime uint64) bool {
	return curTime > v.EndTime
}

// HasVote checks if sender has already voted.
func (v Voting) HasVote(targetAddr string) bool {
	for _, tally := range v.Tallies {
		for _, addr := range tally.YesAddrs {
			if addr == targetAddr {
				return true
			}
		}
		for _, addr := range tally.NoAddrs {
			if addr == targetAddr {
				return true
			}
		}
	}

	return false
}

// AddYesVote appends a new "yes" vote.
func (v *Voting) AddYesVote(option, addr string) error {
	for i := 0; i < len(v.Tallies); i++ {
		if v.Tallies[i].Option != option {
			continue
		}
		v.Tallies[i].YesAddrs = append(v.Tallies[i].YesAddrs, addr)
		return nil
	}

	return errors.New("option not found")
}

// AddNoVote appends a new "no" vote.
func (v *Voting) AddNoVote(option, addr string) error {
	for i := 0; i < len(v.Tallies); i++ {
		if v.Tallies[i].Option != option {
			continue
		}
		v.Tallies[i].NoAddrs = append(v.Tallies[i].NoAddrs, addr)
		return nil
	}

	return errors.New("option not found")
}

// NextVotingID returns a next unique Voting ID.
func NextVotingID(storage std.Storage) (uint64, error) {
	data := storage.Get(LastVotingIDKey)
	if data == nil {
		return 0, nil
	}
	if len(data) != 8 {
		return 0, errors.New("invalid lastVotingID")
	}

	lastID := binary.LittleEndian.Uint64(data)

	return lastID + 1, nil
}

// SetLastVotingID sets LastVotingIDKey.
func SetLastVotingID(storage std.Storage, id uint64) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, id)

	storage.Set(LastVotingIDKey, data)
}

// GetVoting returns Voting state if exists.
func GetVoting(storage std.Storage, id uint64) (voting *Voting, retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("voting state get: " + retErr.Error())
		}
	}()

	key := buildVotingKey(id)
	bz := storage.Get(key)
	if bz == nil {
		return nil, nil
	}

	var obj Voting
	if err := obj.UnmarshalJSON(bz); err != nil {
		retErr = errors.New("object JSON unmarshal")
		return
	}
	voting = &obj

	return
}

// SetVoting sets Voting state.
func SetVoting(storage std.Storage, voting Voting) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("voting state set: " + retErr.Error())
		}
	}()

	bz, err := voting.MarshalJSON()
	if err != nil {
		retErr = errors.New("object JSON marshal: " + err.Error())
		return
	}

	key := buildVotingKey(voting.ID)
	storage.Set(key, bz)

	return
}

// IterateVotings iterates over all stored voting objects.
func IterateVotings(storage std.Storage, fn func(voting Voting) (stop bool)) (retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("range over voting state: " + retErr.Error())
		}
	}()

	iter := storage.Range(VotingKey, nil, std.Ascending)
	for {
		_, bz, err := iter.Next()
		if err != nil {
			break
		}

		var voting Voting
		if err := voting.UnmarshalJSON(bz); err != nil {
			retErr = errors.New("object JSON unmarshal")
			return
		}

		if fn(voting) {
			break
		}
	}

	return
}

// buildVotingKey builds Voting storage key.
func buildVotingKey(id uint64) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, id)

	return append(VotingKey, data...)
}
