// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenListGroupInvitationsListsThePendingInvitationsOfAGroupNewestFirst builds ListGroupInvitations for "lists the pending invitations of a group newest first": the use case, its input, and
// what its output must be.
func givenListGroupInvitationsListsThePendingInvitationsOfAGroupNewestFirst(t *testing.T) (ListGroupInvitations, ListGroupInvitationsInput, func(*testing.T, ListGroupInvitationsOutput)) {
	t.Helper()
	w := newWorld()
	g, other := w.core(), w.core()
	older := w.invited(g, "user_ada", "user_dee", team.InvitingPending)
	w.invited(g, "user_ada", "user_dee", team.InvitingDeclined)
	w.invited(g, "user_ada", "user_dee", team.InvitingWithdrawn)
	newer := w.invited(g, "user_bob", "user_eve", team.InvitingPending)
	w.invited(other, "user_ada", "user_dee", team.InvitingPending) // another group's
	return w.groupInvitations(), ListGroupInvitationsInput{User: "user_bob", ID: g.ID}, func(t *testing.T, out ListGroupInvitationsOutput) {
		t.Helper()
		if len(out.Invitations) != 2 || out.Invitations[0].ID != newer.ID || out.Invitations[1].ID != older.ID {
			t.Fatalf("the invitations: %+v, want the newer then the older one", out.Invitations)
		}
		if v := out.Invitations[0]; v.User.Name != "eve" || v.Inviter.Name != "bob" || !v.CreatedAt.Equal(newer.CreatedAt) {
			t.Errorf("an invitation: %+v", v)
		}
		raw, err := json.Marshal(out)
		if err != nil || strings.Contains(string(raw), "@") || strings.Contains(string(raw), "email") {
			t.Errorf("the answer %s (%v) must name people, not addresses", raw, err)
		}
	}
}

// givenListGroupInvitationsRefusesTheGroupOfOthers builds ListGroupInvitations for "refuses the group of others": the use case and the input it
// refuses with team.ErrNotFound.
func givenListGroupInvitationsRefusesTheGroupOfOthers(t *testing.T) (ListGroupInvitations, ListGroupInvitationsInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.groupInvitations(), ListGroupInvitationsInput{User: "user_dee", ID: g.ID}
}

// givenListGroupInvitationsRefusesAPlainMember builds ListGroupInvitations for "refuses a plain member": the use case and the input it
// refuses with team.ErrNotManager.
func givenListGroupInvitationsRefusesAPlainMember(t *testing.T) (ListGroupInvitations, ListGroupInvitationsInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.groupInvitations(), ListGroupInvitationsInput{User: "user_cy", ID: g.ID}
}
