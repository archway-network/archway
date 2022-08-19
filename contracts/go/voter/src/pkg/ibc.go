package pkg

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"unicode"
)

// ValidateChannelID validates IBC channel ID.
func ValidateChannelID(channelID string) error {
	if !strings.HasPrefix(channelID, "channel-") {
		return errors.New("invalid prefix (must be 'channel-')")
	}

	channelSequence := strings.TrimPrefix(channelID, "channel-")
	if len(channelSequence) < 1 || len(channelSequence) > 20 {
		return errors.New("invalid sequence (must be GTE 1 and LTE 20")
	}

	for _, c := range channelSequence {
		if !unicode.IsNumber(c) {
			return errors.New("found an unsupported char (number is expected): " + string(c))
		}
	}

	sequence, err := strconv.ParseUint(channelSequence, 10, 64)
	if err != nil {
		return errors.New("failed to parse sequence: " + err.Error())
	}
	if sequence > math.MaxUint32 {
		return errors.New("sequence is too large")
	}

	return nil
}
