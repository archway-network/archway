//go:build precision_test

// NOTE: this test is run as a separate build tag as it modifies a global
// variable that is used in other tests. In a concurrent environment this
// might cause unforeseen issues, so we isolate it.
package upgrade052_test

import "testing"

func TestPrecisionBreakages(t *testing.T) {

}
