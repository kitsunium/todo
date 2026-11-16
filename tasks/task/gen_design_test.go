// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:0c616a2ddc211ada253896609000dcaab6758455e2991084b86e140dd79690fe

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
const designDigest = "sha256:0c616a2ddc211ada253896609000dcaab6758455e2991084b86e140dd79690fe"

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

// TestPropertyTaskSeenByOwnerAndSharedOnly holds tasks/Task.SeenBy by property OwnerAndSharedOnly.
func TestPropertyTaskSeenByOwnerAndSharedOnly(t *testing.T) {
	seed := designSeed("tasks/Task.SeenBy#OwnerAndSharedOnly")
	useSeed(t, seed)
	propertyTaskSeenByOwnerAndSharedOnly(t, seed)
}

// TestPropertyParsePriorityKnownOrRefused holds tasks/function/ParsePriority by property KnownOrRefused.
func TestPropertyParsePriorityKnownOrRefused(t *testing.T) {
	seed := designSeed("tasks/function/ParsePriority#KnownOrRefused")
	useSeed(t, seed)
	propertyParsePriorityKnownOrRefused(t, seed)
}
