// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:32913ff1f1c2e4c8f3bc85573da565471d0c6a3bf2b70f14abb926ad227fee32

package team

import (
	"flag"
	"hash/fnv"
	"os"
	"strconv"
	"testing"
)

// designDigest is the design the properties run under: with the property's
// path, it seeds its values, so a pull request replays the same ones.
const designDigest = "sha256:32913ff1f1c2e4c8f3bc85573da565471d0c6a3bf2b70f14abb926ad227fee32"

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

// TestPropertyGroupMemberIDsOneIDAMember holds groups/Group.MemberIDs by property OneIDAMember.
func TestPropertyGroupMemberIDsOneIDAMember(t *testing.T) {
	seed := designSeed("groups/Group.MemberIDs#OneIDAMember")
	useSeed(t, seed)
	propertyGroupMemberIDsOneIDAMember(t, seed)
}

// TestPropertyGroupRoleOfEachMemberHasItsRole holds groups/Group.RoleOf by property EachMemberHasItsRole.
func TestPropertyGroupRoleOfEachMemberHasItsRole(t *testing.T) {
	seed := designSeed("groups/Group.RoleOf#EachMemberHasItsRole")
	useSeed(t, seed)
	propertyGroupRoleOfEachMemberHasItsRole(t, seed)
}

// TestPropertyCleanNameTrimmedOrRefused holds groups/function/CleanName by property TrimmedOrRefused.
func TestPropertyCleanNameTrimmedOrRefused(t *testing.T) {
	seed := designSeed("groups/function/CleanName#TrimmedOrRefused")
	useSeed(t, seed)
	propertyCleanNameTrimmedOrRefused(t, seed)
}

// TestPropertyColorForKnownOrRefused holds groups/function/ColorFor by property KnownOrRefused.
func TestPropertyColorForKnownOrRefused(t *testing.T) {
	seed := designSeed("groups/function/ColorFor#KnownOrRefused")
	useSeed(t, seed)
	propertyColorForKnownOrRefused(t, seed)
}

// TestPropertyManagesOwnerAndAdminOnly holds groups/function/Manages by property OwnerAndAdminOnly.
func TestPropertyManagesOwnerAndAdminOnly(t *testing.T) {
	seed := designSeed("groups/function/Manages#OwnerAndAdminOnly")
	useSeed(t, seed)
	propertyManagesOwnerAndAdminOnly(t, seed)
}

// TestPropertyParseRoleAssignableOnly holds groups/function/ParseRole by property AssignableOnly.
func TestPropertyParseRoleAssignableOnly(t *testing.T) {
	seed := designSeed("groups/function/ParseRole#AssignableOnly")
	useSeed(t, seed)
	propertyParseRoleAssignableOnly(t, seed)
}

// TestPropertyRankedOwnerFirst holds groups/function/Ranked by property OwnerFirst.
func TestPropertyRankedOwnerFirst(t *testing.T) {
	seed := designSeed("groups/function/Ranked#OwnerFirst")
	useSeed(t, seed)
	propertyRankedOwnerFirst(t, seed)
}
