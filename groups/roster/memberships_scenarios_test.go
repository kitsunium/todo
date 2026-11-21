// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenMembershipsListsTheGroupsAMemberBelongsTo builds Memberships for "lists the groups a member belongs to": the use case, its input, and
// what its output must be.
func givenMembershipsListsTheGroupsAMemberBelongsTo(t *testing.T) (Memberships, MembershipsInput, func(*testing.T, MembershipsOutput)) {
	t.Helper()
	w := newWorld()
	g, h := w.core(), w.core()
	h.Name = "Alpha"
	h.Members = h.Members[:2] // cy and bob, ada is not in it
	w.groups[h.ID] = h
	return w.memberships(), MembershipsInput{User: "user_cy"}, func(t *testing.T, out MembershipsOutput) {
		t.Helper()
		if len(out.Memberships) != 2 {
			t.Fatalf("the memberships: %+v", out.Memberships)
		}
		first, second := out.Memberships[0], out.Memberships[1]
		if first.Group.ID != h.ID || first.Group.Name != "Alpha" || first.Role != team.RoleMember || second.Group.ID != g.ID || second.Group.Color != "teal" || second.Role != team.RoleMember {
			t.Errorf("the memberships: %+v", out.Memberships)
		}
	}
}
