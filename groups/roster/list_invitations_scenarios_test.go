// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenListInvitationsListsThePendingInvitationsNewestFirst builds ListInvitations for "lists the pending invitations newest first": the use case, its input, and
// what its output must be.
func givenListInvitationsListsThePendingInvitationsNewestFirst(t *testing.T) (ListInvitations, ListInvitationsInput, func(*testing.T, ListInvitationsOutput)) {
	t.Helper()
	w := newWorld()
	g, h := w.core(), w.core()
	older := w.invited(g, "user_ada", "user_dee", team.InvitingPending)
	w.invited(g, "user_ada", "user_dee", team.InvitingDeclined)
	newer := w.invited(h, "user_bob", "user_dee", team.InvitingPending)
	w.invited(h, "user_ada", "user_eve", team.InvitingPending)
	gone := w.core()
	w.invited(gone, "user_ada", "user_dee", team.InvitingPending)
	delete(w.groups, gone.ID)
	return w.invitationsList(), ListInvitationsInput{User: "user_dee"}, func(t *testing.T, out ListInvitationsOutput) {
		t.Helper()
		if len(out.Invitations) != 2 || out.Invitations[0].ID != newer.ID || out.Invitations[1].ID != older.ID {
			t.Fatalf("the invitations: %+v, want the newer then the older one", out.Invitations)
		}
		if v := out.Invitations[0]; v.Group.ID != h.ID || v.Group.Name != "Core" || v.Inviter.Name != "bob" || v.Invitee.ID != "" || v.State != team.InvitingPending {
			t.Errorf("an invitation: %+v", v)
		}
	}
}
