package pkg

import (
	"errors"
	"unicode"
)

// reDenomString is a reference regExp from the Cosmos SDK.
const reDenomString = `[a-zA-Z][a-zA-Z0-9/-]{2,127}`

// ValidateDenom validates Coin.Denom.
func ValidateDenom(denom string) error {
	if len(denom) < 2 || len(denom) > 127 {
		return errors.New("invalid len (must be GTE 2 and LTE 127")
	}

	for i, c := range denom {
		if c > unicode.MaxASCII {
			return errors.New("not ASCII char found")
		}

		switch i {
		case 0:
			if !unicode.IsLetter(c) {
				return errors.New("first char must be a letter")
			}
		default:
			if !unicode.IsLetter(c) && !unicode.IsNumber(c) && c != '/' && c != '-' {
				return errors.New("found an unsupported char: " + string(c))
			}
		}
	}

	return nil
}
