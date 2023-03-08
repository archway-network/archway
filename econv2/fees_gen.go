package main

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"math"
	"os"
	"time"
)

const maxElapsedBlockTimes = 68 * time.Second
const maxPrecision = 1 * time.Millisecond
const instances = maxElapsedBlockTimes / maxPrecision

// log4(2+t)*200m, precision is ms.
func calcLog4(t time.Duration) sdk.Dec {
	// we use the base change formula which says that:
	// logb(x) = logk(x) / logk(b)
	// where b: 4
	// x: 2+t
	// and k is 10, because golang provides us the log10
	// log4(2+t) = log10(2+t) / log10(4)

	s := math.Log10(2+float64(t)) / math.Log10(4)
	fmt.Fprintf(os.Stderr, "%.50f\n", s)
	return sdk.ZeroDec()
}

func main() {
	for i := time.Duration(0); i < instances; i++ {
		calcLog4(i)
	}
}
