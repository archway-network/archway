package common

import (
	"fmt"
	"time"

	collcodec "cosmossdk.io/collections/codec"
	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	DecValueEncoder collcodec.KeyCodec[math.LegacyDec] = decValueEncoder{}
	TimeKeyEncoder  collcodec.KeyCodec[time.Time]      = timeKeyEncoder{}
)

func HumanizeBytes(bz []byte) string {
	return fmt.Sprintf("\nbytesAsHex: %x", bz)
}

// Collection Codecs

// math.LegacyDec

type decValueEncoder struct{}

func (decValueEncoder) Stringify(value math.LegacyDec) string {
	return value.String()
}
func (decValueEncoder) Encode(buffer []byte, value math.LegacyDec) (int, error) {
	b, err := value.Marshal()
	copy(buffer, b)
	return len(b), err
}
func (decValueEncoder) Decode(b []byte) (int, math.LegacyDec, error) {
	dec := new(math.LegacyDec)
	err := dec.Unmarshal(b)
	return len(b), *dec, err
}
func (decValueEncoder) EncodeJSON(value math.LegacyDec) ([]byte, error) {
	b, err := value.MarshalJSON()
	return b, err
}
func (decValueEncoder) DecodeJSON(b []byte) (math.LegacyDec, error) {
	dec := new(math.LegacyDec)
	err := dec.UnmarshalJSON(b)
	return *dec, err
}
func (dve decValueEncoder) EncodeNonTerminal(buffer []byte, key math.LegacyDec) (int, error) {
	return dve.Encode(buffer, key)
}
func (dve decValueEncoder) DecodeNonTerminal(buffer []byte) (int, math.LegacyDec, error) {
	return dve.Decode(buffer)
}
func (decValueEncoder) Size(key math.LegacyDec) int {
	b, _ := key.Marshal()
	return len(b)
}
func (tke decValueEncoder) SizeNonTerminal(key math.LegacyDec) int {
	return tke.Size(key)
}
func (d decValueEncoder) KeyType() string {
	return "math.LegacyDec"
}

// std.time.Time

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
