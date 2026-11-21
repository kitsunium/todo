// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenDeleteGroupDeletesAGroup builds DeleteGroup for "deletes a group": the use case, its input, and
// what its output must be.
func givenDeleteGroupDeletesAGroup(t *testing.T) (DeleteGroup, DeleteGroupInput, func(*testing.T, DeleteGroupOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	other := w.core()
	w.invited(g, "user_ada", "user_dee", team.InvitingPending)
	w.invited(g, "user_ada", "user_eve", team.InvitingDeclined)
	kept := w.invited(other, "user_ada", "user_dee", team.InvitingPending)
	return w.remove(), DeleteGroupInput{User: "user_ada", ID: g.ID}, func(t *testing.T, _ DeleteGroupOutput) {
		t.Helper()
		if _, still := w.groups[g.ID]; still {
			t.Error("the group is still there")
		}
		if _, ok := w.groups[other.ID]; !ok {
			t.Error("another group went with it")
		}
		if len(w.invitations) != 1 || w.invitations[kept.ID].ID == "" {
			t.Errorf("the invitations left: %+v, want the other group's alone", w.invitations)
		}
		want := "group-deleted by user_ada on " + g.ID + " to user_cy,user_bob,user_ada"
		if len(w.notices) != 1 || w.notices[0] != want {
			t.Errorf("the notices: %q, want %q", w.notices, want)
		}
	}
}

// givenDeleteGroupRefusesTheGroupOfOthers builds DeleteGroup for "refuses the group of others": the use case and the input it
// refuses with team.ErrNotFound.
func givenDeleteGroupRefusesTheGroupOfOthers(t *testing.T) (DeleteGroup, DeleteGroupInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.remove(), DeleteGroupInput{User: "user_dee", ID: g.ID}
}

// givenDeleteGroupRefusesAnAdmin builds DeleteGroup for "refuses an admin": the use case and the input it
// refuses with team.ErrNotOwner.
func givenDeleteGroupRefusesAnAdmin(t *testing.T) (DeleteGroup, DeleteGroupInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.remove(), DeleteGroupInput{User: "user_bob", ID: g.ID}
}
