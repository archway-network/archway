//go:build codegen

package main

import (
	"fmt"
	"html/template"
	"math"
	"os"
	"time"
)

const maxElapsedBlockTimes = 68 * time.Second
const maxPrecision = 1 * time.Millisecond
const instances = maxElapsedBlockTimes / maxPrecision

const tmpl = `//go:build !codegen

package econv2
import sdk "github.com/cosmos/cosmos-sdk/types"
var GrowthFactorByDuration = map[int64]sdk.Dec{
	{{ range $key, $value := . }}
    {{$key}}: sdk.MustNewDecFromStr("{{$value}}"),{{ end }}
}
`

var phiPow2 = math.Pow(math.Phi, 2)

// f(t)=log𝝅(𝚽+t)*𝚽^2
func calcLogPP(t time.Duration) string {
	// logb(x) = logk(x) / logk(b)
	// log𝝅(𝚽+t) = log10(𝚽+t) / log10(𝝅)
	s := math.Log10(math.Phi+float64(t)) / math.Log10(math.Pi)
	res := phiPow2 * s
	return fmt.Sprintf("%.50f", res)
}

// log4(2+t), precision is ms.
func calcLog4(t time.Duration) string {
	// we use the base change property of logs which says that:
	// logb(x) = logk(x) / logk(b)
	// where b: 4
	// x: 2+t
	// and k is 10, because golang provides us the log10
	// log4(2+t) = log10(2+t) / log10(4)
	s := math.Log10(2+float64(t)) / math.Log10(4)
	return fmt.Sprintf("%.50f", s)
}

func main() {
	m := map[int64]string{}
	for i := time.Duration(0); i < instances; i++ {
		m[int64(i)] = calcLogPP(i)
	}

	tmpl, err := template.New("x").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	f, err := os.OpenFile("./econv2/econv2phipi.go", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = tmpl.Execute(f, m)
	if err != nil {
		panic(err)
	}
}
