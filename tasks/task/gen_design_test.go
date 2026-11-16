// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:8d40d0873f391460466a0aa214adea3d028e9293117fc8bd920c20c5af44af3e

package task

import (
	"flag"
	"hash/fnv"
	"os"
	"strconv"
	"testing"
)

// designDigest is the design the properties run under: with the property's
// path, it seeds its values, so a pull request replays the same ones.
const designDigest = "sha256:8d40d0873f391460466a0aa214adea3d028e9293117fc8bd920c20c5af44af3e"

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

// TestPropertyTaskOverdueNeverWhenDone holds tasks/Task.Overdue by property NeverWhenDone.
func TestPropertyTaskOverdueNeverWhenDone(t *testing.T) {
	seed := designSeed("tasks/Task.Overdue#NeverWhenDone")
	useSeed(t, seed)
	propertyTaskOverdueNeverWhenDone(t, seed)
}

// TestPropertyParsePriorityKnownOrRefused holds tasks/function/ParsePriority by property KnownOrRefused.
func TestPropertyParsePriorityKnownOrRefused(t *testing.T) {
	seed := designSeed("tasks/function/ParsePriority#KnownOrRefused")
	useSeed(t, seed)
	propertyParsePriorityKnownOrRefused(t, seed)
}
