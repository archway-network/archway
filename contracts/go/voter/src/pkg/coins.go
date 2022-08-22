package pkg

import (
	"errors"
	"strings"
	"unicode"

	"github.com/CosmWasm/cosmwasm-go/std/math"
	stdTypes "github.com/CosmWasm/cosmwasm-go/std/types"
)

// CoinsContainMinAmount checks that coins have a target coin which amount is GTE to target.
func CoinsContainMinAmount(coins []stdTypes.Coin, coinExpected stdTypes.Coin) error {
	for _, coin := range coins {
		if coin.Denom != coinExpected.Denom {
			continue
		}

		if coin.Amount.LT(coinExpected.Amount) {
			break
		}

		return nil
	}

	return errors.New("expected coin amount (" + coinExpected.String() + "): not found")
}

// AddCoins merges coin slices.
// CONTRACT: doesn't handle duplicated denoms within coins slice.
func AddCoins(srcCoins []stdTypes.Coin, dstCoins ...stdTypes.Coin) []stdTypes.Coin {
	addCoin := func(coin stdTypes.Coin) {
		for i := 0; i < len(srcCoins); i++ {
			if srcCoins[i].Denom != coin.Denom {
				continue
			}

			srcCoins[i].Amount = srcCoins[i].Amount.Add(coin.Amount)
			return
		}

		srcCoins = append(srcCoins, coin)
	}

	for _, dstCoin := range dstCoins {
		addCoin(dstCoin)
	}

	return srcCoins
}

// ParseCoinFromString parses and validates stdTypes.Coin.
func ParseCoinFromString(coinStr string) (retCoin stdTypes.Coin, retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("parsing coin (" + coinStr + "): " + retErr.Error())
		}
	}()

	coinStr = strings.TrimSpace(coinStr)

	denomStartIdx := -1
	for i, c := range coinStr {
		if c > unicode.MaxASCII {
			retErr = errors.New("not ASCII char found")
			return
		}

		if !unicode.IsDigit(c) {
			denomStartIdx = i
			break
		}
	}
	if denomStartIdx == -1 {
		retErr = errors.New("denom start: not found")
		return
	}

	var amount math.Uint128
	if err := amount.FromString(coinStr[:denomStartIdx]); err != nil {
		retErr = errors.New("amount parse: " + err.Error())
		return
	}

	denom := coinStr[denomStartIdx:]
	if err := ValidateDenom(denom); err != nil {
		retErr = errors.New("denom validation: " + err.Error())
		return
	}

	retCoin.Denom = denom
	retCoin.Amount = amount

	return
}

// ParseCoinsFromString parses and validates []stdTypes.Coin.
func ParseCoinsFromString(coinsStr string) (retCoins []stdTypes.Coin, retErr error) {
	defer func() {
		if retErr != nil {
			retErr = errors.New("parsing coins: " + retErr.Error())
		}
	}()

	coinsStr = strings.TrimSpace(coinsStr)
	if coinsStr == "" {
		return
	}

	coinStrs := strings.Split(coinsStr, ",")
	retCoins = make([]stdTypes.Coin, 0, len(coinStrs))
	for _, coinStr := range coinStrs {
		coin, err := ParseCoinFromString(coinStr)
		if err != nil {
			retErr = err
			return
		}
		retCoins = append(retCoins, coin)
	}

	return
}
