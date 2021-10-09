package gastracker

import "math/big"

// This function is mainly used to safely convert the uint64 number to big.Int and
// then to sdk.Dec/sdk.Int
func ConvertUint64ToBigInt(n uint64) *big.Int {
	return big.NewInt(0).SetUint64(n)
}
