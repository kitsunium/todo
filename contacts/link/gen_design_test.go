// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:f78b90bd39d042efd423450d5f8ea59f0f6ddf7d6ea3d6874b6a7bc17cd1eabd

package link

import (
	"flag"
	"hash/fnv"
	"os"
	"strconv"
	"testing"
)

// designDigest is the design the properties run under: with the property's
// path, it seeds its values, so a pull request replays the same ones.
const designDigest = "sha256:f78b90bd39d042efd423450d5f8ea59f0f6ddf7d6ea3d6874b6a7bc17cd1eabd"

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

// TestPropertyLinkOtherTheOtherEnd holds contacts/Link.Other by property TheOtherEnd.
func TestPropertyLinkOtherTheOtherEnd(t *testing.T) {
	seed := designSeed("contacts/Link.Other#TheOtherEnd")
	useSeed(t, seed)
	propertyLinkOtherTheOtherEnd(t, seed)
}

// TestPropertyPairOfSymmetric holds contacts/function/PairOf by property Symmetric.
func TestPropertyPairOfSymmetric(t *testing.T) {
	seed := designSeed("contacts/function/PairOf#Symmetric")
	useSeed(t, seed)
	propertyPairOfSymmetric(t, seed)
}
