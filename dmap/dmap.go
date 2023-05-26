// Package dmap provides functions and mechanisms to ensure deterministic
// map behavior.
package dmap

import (
	"sort"

	"golang.org/x/exp/constraints"
)

func SortedKeys[K constraints.Ordered, V any](in map[K]V) []K {
	// Performance: using a known size yields faster allocation
	// than using append;
	out := make([]K, len(in))

	i := 0
	for k := range in {
		out[i] = k
		i++
	}

	SortSlice(out)

	return out
}

func SortSlice[T constraints.Ordered](s []T) {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
}
