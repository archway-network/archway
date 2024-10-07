// Package omap defines a generic-based type for creating ordered maps. It
// exports a "Sorter" interface, allowing the creation of ordered maps with
// custom key and value types.
//
// Specifically, omap supports ordered maps with keys of type string or
// asset.Pair and values of any type. See impl.go for examples.
//
// ## Motivation
//
// Ensuring deterministic behavior is crucial in blockchain systems, as all
// nodes must reach a consensus on the state of the blockchain. Every action,
// given the same input, should consistently yield the same result. A
// divergence in state could impede the ability of nodes to validate a block,
// prohibiting the addition of the block to the chain, which could lead to
// chain halts.
package omap

import (
	"sort"
)

// OrderedMap is a wrapper struct around the built-in map that has guarantees
// about order because it sorts its keys with a custom sorter. It has a public
// API that mirrors that functionality of `map`. OrderedMap is built with
// generics, so it can hold various combinations of key-value types.
type OrderedMap[K comparable, V any] struct {
	Data        map[K]V
	orderedKeys []K
	keyIndexMap map[K]int // useful for delete operation
	isOrdered   bool
	sorter      Sorter[K]
}

// Sorter is an interface used for ordering the keys in the OrderedMap.
type Sorter[K any] interface {
	// Returns true if 'a' is less than 'b' Less needs to be defined for the
	// key type, K, to provide a comparison operation.
	Less(a K, b K) bool
}

// ensureOrder is a method on the OrderedMap that sorts the keys in the map
// and rebuilds the index map.
func (om *OrderedMap[K, V]) ensureOrder() {
	keys := make([]K, 0, len(om.Data))
	for key := range om.Data {
		keys = append(keys, key)
	}

	// Sort the keys using the Sort function
	lessFunc := func(i, j int) bool {
		return om.sorter.Less(keys[i], keys[j])
	}
	sort.Slice(keys, lessFunc)

	om.orderedKeys = keys
	om.keyIndexMap = make(map[K]int)
	for idx, key := range om.orderedKeys {
		om.keyIndexMap[key] = idx
	}
	om.isOrdered = true
}

// BuildFrom is a method that builds an OrderedMap from a given map and a
// sorter for the keys. This function is useful for creating new OrderedMap
// types with typed keys.
func (om OrderedMap[K, V]) BuildFrom(
	data map[K]V, sorter Sorter[K],
) OrderedMap[K, V] {
	om.Data = data
	om.sorter = sorter
	om.ensureOrder()
	return om
}

// Range returns a channel of keys in their sorted order. This allows you
// to iterate over the map in a deterministic order. Using a channel here
// makes it so that the iteration is done lazily rather loading the entire
// map (OrderedMap.data) into memory and then iterating.
func (om OrderedMap[K, V]) Range() <-chan (K) {
	iterChan := make(chan K)
	go func() {
		defer close(iterChan)
		// Generate or compute values on-demand
		for _, key := range om.orderedKeys {
			iterChan <- key
		}
	}()
	return iterChan
}

// Has checks whether a key exists in the map.
func (om OrderedMap[K, V]) Has(key K) bool {
	_, exists := om.Data[key]
	return exists
}

// Len returns the number of items in the map.
func (om OrderedMap[K, V]) Len() int {
	return len(om.Data)
}

// Keys returns a slice of the keys in their sorted order.
func (om *OrderedMap[K, V]) Keys() []K {
	if !om.isOrdered {
		om.ensureOrder()
	}
	return om.orderedKeys
}

// Set adds a key-value pair to the map, or updates the value if the key
// already exists. It ensures the keys are ordered after the operation.
func (om *OrderedMap[K, V]) Set(key K, val V) {
	om.Data[key] = val
	om.ensureOrder() // TODO perf: make this more efficient with a clever insert.
}

// Delete removes a key-value pair from the map if the key exists.
func (om *OrderedMap[K, V]) Delete(key K) {
	idx, keyExists := om.keyIndexMap[key]
	if keyExists {
		delete(om.Data, key)

		orderedKeys := om.orderedKeys
		orderedKeys = append(orderedKeys[:idx], orderedKeys[idx+1:]...)
		om.orderedKeys = orderedKeys
	}
}
