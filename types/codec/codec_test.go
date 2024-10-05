package codec

import (
	"testing"
	"time"

	collcodec "cosmossdk.io/collections/codec"
	"github.com/stretchr/testify/require"
)

func assertBijective[T any](t *testing.T, encoder collcodec.KeyCodec[T], key T) {
	size := encoder.Size(key)
	encodedKey := make([]byte, size)

	write, err := encoder.Encode(encodedKey, key)
	require.NoError(t, err)

	read, decodedKey, err := encoder.Decode(encodedKey)
	require.NoError(t, err)

	require.Equal(t, write, size, "write bytes do not match the size")
	require.Equal(t, read, size, "read bytes do not match the size")
	require.Equal(t, len(encodedKey), size, "encoded key does not match the size")
	require.Equal(t, key, decodedKey, "encoding and decoding produces different keys")
}

func TestTimeKey(t *testing.T) {
	t.Run("bijective", func(t *testing.T) {
		key := time.Now()
		assertBijective[time.Time](t, TimeKeyEncoder, key.Round(0).UTC())
	})
}
