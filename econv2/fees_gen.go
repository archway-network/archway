//go:build codegen

package main

import (
	"fmt"
	"html/template"
	"math"
	"os"
)

const tmpl = `//go:build !codegen

// Code generated. DO NOT EDIT.
package econv2
import sdk "github.com/cosmos/cosmos-sdk/types"
var GrowthFactorByDuration = map[int64]sdk.Dec{
	{{ range $key, $value := . }}
    {{$key}}: sdk.MustNewDecFromStr("{{$value}}"),{{ end }}
}
`

var phiPow2 = math.Pow(math.Phi, 2)

// f(t)=log𝝅(𝚽+t)*𝚽^2
func calcLogPP(t float64) string {
	// logb(x) = logk(x) / logk(b)
	// log𝝅(𝚽+t) = log10(𝚽+t) / log10(𝝅)
	s := math.Log10(math.Phi+t) / math.Log10(math.Pi)
	s = phiPow2 * s
	return fmt.Sprintf("%.60f", s)
}

// log4(2+t), precision is ms.
func calcLog4(t float64) string {
	// we use the base change property of logs which says that:
	// logb(x) = logk(x) / logk(b)
	// where b: 4
	// x: 2+t
	// and k is 10, because golang provides us the log10
	// log4(2+t) = log10(2+t) / log10(4)
	s := math.Log10(2+t) / math.Log10(4)
	return fmt.Sprintf("%.50f", s)
}

func main() {
	m := map[int64]string{}
	for i := float64(0); i <= 68; i += 0.001 {
		m[int64(i*1000)] = calcLogPP(i)
		//m[int64(i*1000)] = calcLog4(i)
	}

	tmpl, err := template.New("x").Parse(tmpl)
	if err != nil {
		panic(err)
	}

	//f, err := os.OpenFile("./econv2/econv2log4.go", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	f, err := os.OpenFile("./econv2/econv2logpiphi.go", os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	err = tmpl.Execute(f, m)
	if err != nil {
		panic(err)
	}
}
