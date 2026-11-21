// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenInviteInvitesAContact builds Invite for "invites a contact": the use case, its input, and
// what its output must be.
func givenInviteInvitesAContact(t *testing.T) (Invite, InviteInput, func(*testing.T, InviteOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	w.befriend("user_bob", "user_dee")
	return w.invite(), InviteInput{User: "user_bob", ID: g.ID, Invitee: "user_dee"}, func(t *testing.T, out InviteOutput) {
		t.Helper()
		v := out.Invitation
		if v.Group.ID != g.ID || v.Inviter.Name != "bob" || v.Invitee.Name != "dee" || v.State != team.InvitingPending {
			t.Errorf("the invitation: %+v", v)
		}
		kept := w.invitations[v.ID]
		if kept.GroupID != g.ID || kept.InviterID != "user_bob" || kept.InviteeID != "user_dee" || kept.State != team.InvitingPending || !kept.RespondedAt.IsZero() || !kept.CreatedAt.After(epoch) {
			t.Errorf("the store keeps %+v", kept)
		}
		want := "group-invited by user_bob on " + g.ID + " to user_dee"
		if len(w.notices) != 1 || w.notices[0] != want {
			t.Errorf("the notices: %q, want %q", w.notices, want)
		}
	}
}

// givenInviteRefusesTheGroupOfOthers builds Invite for "refuses the group of others": the use case and the input it
// refuses with team.ErrNotFound.
func givenInviteRefusesTheGroupOfOthers(t *testing.T) (Invite, InviteInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	w.befriend("user_dee", "user_eve")
	return w.invite(), InviteInput{User: "user_dee", ID: g.ID, Invitee: "user_eve"}
}

// givenInviteRefusesAPlainMember builds Invite for "refuses a plain member": the use case and the input it
// refuses with team.ErrNotManager.
func givenInviteRefusesAPlainMember(t *testing.T) (Invite, InviteInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	w.befriend("user_cy", "user_dee")
	return w.invite(), InviteInput{User: "user_cy", ID: g.ID, Invitee: "user_dee"}
}

// givenInviteRefusesSomeoneAlreadyIn builds Invite for "refuses someone already in": the use case and the input it
// refuses with team.ErrAlreadyMember.
func givenInviteRefusesSomeoneAlreadyIn(t *testing.T) (Invite, InviteInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	w.befriend("user_ada", "user_cy")
	return w.invite(), InviteInput{User: "user_ada", ID: g.ID, Invitee: "user_cy"}
}

// givenInviteRefusesWhoIsNoContact builds Invite for "refuses who is no contact": the use case and the input it
// refuses with team.ErrNotContact.
func givenInviteRefusesWhoIsNoContact(t *testing.T) (Invite, InviteInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.invite(), InviteInput{User: "user_ada", ID: g.ID, Invitee: "user_dee"}
}

// givenInviteRefusesAnInvitationAlreadyWaiting builds Invite for "refuses an invitation already waiting": the use case and the input it
// refuses with team.ErrAlreadyInvited.
func givenInviteRefusesAnInvitationAlreadyWaiting(t *testing.T) (Invite, InviteInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	w.befriend("user_ada", "user_dee")
	w.invited(g, "user_bob", "user_dee", team.InvitingPending)
	return w.invite(), InviteInput{User: "user_ada", ID: g.ID, Invitee: "user_dee"}
}
