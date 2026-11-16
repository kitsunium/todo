// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"
)

// givenListGroupsListsTheGroupsOfAMemberByName builds ListGroups for "lists the groups of a member by name": the use case, its input, and
// what its output must be.
func givenListGroupsListsTheGroupsOfAMemberByName(t *testing.T) (ListGroups, ListGroupsInput, func(*testing.T, ListGroupsOutput)) {
	t.Helper()
	w := newWorld()
	b, a, c := w.core(), w.core(), w.core()
	b.Name, a.Name, c.Name = "beta", "Alpha", "alpha"
	w.groups[a.ID], w.groups[b.ID], w.groups[c.ID] = a, b, c
	elsewhere := w.core()
	elsewhere.Members = elsewhere.Members[:1] // cy alone, the first member of the fixture
	w.groups[elsewhere.ID] = elsewhere
	return w.list(), ListGroupsInput{User: "user_ada"}, func(t *testing.T, out ListGroupsOutput) {
		t.Helper()
		var got []string
		for _, v := range out.Groups {
			got = append(got, v.ID)
		}
		want := []string{a.ID, c.ID, b.ID} // and not the one cy is alone in
		if len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
			t.Errorf("the groups: %v, want %v (by name, ignoring case, then by id)", got, want)
		}
		if out.Groups[0].Role != "owner" {
			t.Errorf("the role of the member who looks: %q", out.Groups[0].Role)
		}
	}
}
