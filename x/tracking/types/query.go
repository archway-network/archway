package types

import "sigs.k8s.io/yaml"

// String implements the fmt.Stringer interface.
func (m TxTracking) String() string {
	bz, _ := yaml.Marshal(m)
	return string(bz)
}
