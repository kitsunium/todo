// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"slices"
	"testing"
)

// givenMembersOfListsTheMembersOfAGroup builds MembersOf for "lists the members of a group": the use case, its input, and
// what its output must be.
func givenMembersOfListsTheMembersOfAGroup(t *testing.T) (MembersOf, MembersOfInput, func(*testing.T, MembersOfOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	uc := NewMembersOf(w.groupsPort())
	return uc, MembersOfInput{Group: g.ID}, func(t *testing.T, out MembersOfOutput) {
		t.Helper()
		want := g.MemberIDs()
		if len(want) == 0 || !slices.Equal(out.Users, want) {
			t.Errorf("the members: %q, want %q", out.Users, want)
		}
		if got, err := uc.Execute(t.Context(), MembersOfInput{Group: "group_nobody"}); err != nil || len(got.Users) != 0 {
			t.Errorf("a group that does not exist has members %q (%v)", got.Users, err)
		}
	}
}
