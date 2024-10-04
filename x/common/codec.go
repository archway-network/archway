package common

import (
	"time"

	collcodec "cosmossdk.io/collections/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	TimeKeyEncoder collcodec.KeyCodec[time.Time] = timeKeyEncoder{}
)

// Collection Codecs

type timeKeyEncoder struct{}

func (timeKeyEncoder) Stringify(t time.Time) string {
	return t.String()
}
func (timeKeyEncoder) Encode(buffer []byte, t time.Time) (int, error) {
	buf := sdk.FormatTimeBytes(t)
	copy(buffer, buf)
	return len(buf), nil
}
func (timeKeyEncoder) Decode(b []byte) (int, time.Time, error) {
	t, err := sdk.ParseTimeBytes(b)
	return len(b), t, err
}
func (timeKeyEncoder) EncodeJSON(value time.Time) ([]byte, error) {
	return value.MarshalJSON()
}
func (tke timeKeyEncoder) DecodeJSON(b []byte) (time.Time, error) {
	t := new(time.Time)
	err := t.UnmarshalJSON(b)
	return *t, err
}
func (tke timeKeyEncoder) EncodeNonTerminal(buffer []byte, key time.Time) (int, error) {
	return tke.Encode(buffer, key)
}
func (tke timeKeyEncoder) DecodeNonTerminal(buffer []byte) (int, time.Time, error) {
	return tke.Decode(buffer)
}
func (timeKeyEncoder) Size(key time.Time) int {
	return len(sdk.FormatTimeString(key))
}
func (timeKeyEncoder) KeyType() string {
	return "archway.timeKeyEncoder"
}
func (tke timeKeyEncoder) SizeNonTerminal(key time.Time) int {
	return tke.Size(key)
}
