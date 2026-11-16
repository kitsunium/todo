// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:35933b95eb37968572e6d46dc1379b8281e3503fec0e8451a525bac3fb04f4af

package account

import (
	"flag"
	"hash/fnv"
	"os"
	"strconv"
	"testing"
)

// designDigest is the design the properties run under: with the property's
// path, it seeds its values, so a pull request replays the same ones.
const designDigest = "sha256:35933b95eb37968572e6d46dc1379b8281e3503fec0e8451a525bac3fb04f4af"

// designSeed is the seed of a property: the design's digest and its path —
// and DESIGN_SEED when it is set.
func designSeed(path string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(designDigest + "|" + path + "|" + os.Getenv("DESIGN_SEED")))
	return h.Sum64()
}

// useSeed logs a property's seed and hands it to pgregory.net/rapid, whose
// only public way in is its -rapid.seed flag: rapid.Check then draws the same
// values in every run of the same design and the same DESIGN_SEED.
func useSeed(t *testing.T, seed uint64) {
	t.Helper()
	t.Logf("property seed %d (go test -rapid.seed=%d replays it)", seed, seed)
	if f := flag.Lookup("rapid.seed"); f != nil {
		_ = f.Value.Set(strconv.FormatUint(seed, 10))
	}
}

// TestPropertyAccountLockedOnlyBeforeItsEnd holds identity/Account.Locked by property OnlyBeforeItsEnd.
func TestPropertyAccountLockedOnlyBeforeItsEnd(t *testing.T) {
	seed := designSeed("identity/Account.Locked#OnlyBeforeItsEnd")
	useSeed(t, seed)
	propertyAccountLockedOnlyBeforeItsEnd(t, seed)
}

// TestPropertyNormalizeEmailLowerOrRefused holds identity/function/NormalizeEmail by property LowerOrRefused.
func TestPropertyNormalizeEmailLowerOrRefused(t *testing.T) {
	seed := designSeed("identity/function/NormalizeEmail#LowerOrRefused")
	useSeed(t, seed)
	propertyNormalizeEmailLowerOrRefused(t, seed)
}
