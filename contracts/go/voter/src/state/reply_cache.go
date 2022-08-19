package state

import (
	"encoding/binary"
	"errors"

	"github.com/CosmWasm/cosmwasm-go/std"
)

var (
	// LastReplyIDKey is the storage key for storing last unique ReplyMsg ID.
	LastReplyIDKey = []byte("LastReplyID")

	// ReplyMsgTypeKey is the storage prefix for storing ReplyMsgType by ReplyID.
	ReplyMsgTypeKey = []byte("ReplyMsgType")
)

// ReplyMsgType defines the enum of supported reply types.
type ReplyMsgType uint8

const (
	ReplyMsgTypeBank     ReplyMsgType = iota + 1
	ReplyMsgTypeWithdraw ReplyMsgType = iota + 1
)

// GetReplyMsgType returns ReplyMsgType by Reply ID if found.
func GetReplyMsgType(storage std.Storage, replyID uint64) (msgType ReplyMsgType, found bool, retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("replyMsgType state get: " + retErr.Error())
		}
	}()

	key := buildReplyMsgTypeKeyKey(replyID)
	bz := storage.Get(key)
	if bz == nil {
		return
	}

	if len(bz) != 1 {
		retErr = errors.New("invalid object length")
		return
	}
	msgType, found = ReplyMsgType(bz[0]), true

	return
}

// SetReplyMsgType picks a next unique Reply ID and stores ReplyMsgType.
func SetReplyMsgType(storage std.Storage, msgType ReplyMsgType) (replyID uint64, retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("replyMsgType state set: " + retErr.Error())
		}
	}()

	id, err := nextReplyID(storage)
	if err != nil {
		retErr = err
		return
	}
	replyID = id

	key := buildReplyMsgTypeKeyKey(id)
	storage.Set(key, []byte{byte(msgType)})

	return
}

// nextReplyID returns a next unique Reply ID.
func nextReplyID(storage std.Storage) (uint64, error) {
	data := storage.Get(LastReplyIDKey)
	if data == nil {
		return 0, nil
	}
	if len(data) != 8 {
		return 0, errors.New("invalid lastReplyID")
	}

	lastID := binary.LittleEndian.Uint64(data)

	return lastID + 1, nil
}

// setLastReplyID sets LastReplyIDKey.
func setLastReplyID(storage std.Storage, id uint64) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, id)

	storage.Set(LastReplyIDKey, data)
}

// buildReplyMsgTypeKeyKey builds ReplyMsgType storage key by unique Reply ID.
func buildReplyMsgTypeKeyKey(id uint64) []byte {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, id)

	return append(ReplyMsgTypeKey, data...)
}
