// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenAnswerInvitationAcceptsAnInvitation builds AnswerInvitation for "accepts an invitation": the use case, its input, and
// what its output must be.
func givenAnswerInvitationAcceptsAnInvitation(t *testing.T) (AnswerInvitation, AnswerInvitationInput, func(*testing.T, AnswerInvitationOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	inv := w.invited(g, "user_bob", "user_dee", team.InvitingPending)
	return w.answer(), AnswerInvitationInput{User: "user_dee", ID: inv.ID, Event: "accept"}, func(t *testing.T, out AnswerInvitationOutput) {
		t.Helper()
		if out.Group.ID != g.ID || out.Group.Role != team.RoleMember || len(out.Group.Members) != 4 {
			t.Errorf("the group joined: %+v", out.Group)
		}
		if role := w.role(g.ID, "user_dee"); role != team.RoleMember {
			t.Errorf("dee is %q in the store", role)
		}
		if kept := w.invitations[inv.ID]; kept.State != team.InvitingAccepted || !kept.RespondedAt.After(epoch) {
			t.Errorf("the invitation kept: %+v", kept)
		}
		want := "group-joined by user_dee on " + g.ID + " to user_cy,user_bob,user_ada,user_dee"
		if len(w.notices) != 1 || w.notices[0] != want {
			t.Errorf("the notices: %q, want %q", w.notices, want)
		}
	}
}

// givenAnswerInvitationDeclinesAnInvitation builds AnswerInvitation for "declines an invitation": the use case, its input, and
// what its output must be.
func givenAnswerInvitationDeclinesAnInvitation(t *testing.T) (AnswerInvitation, AnswerInvitationInput, func(*testing.T, AnswerInvitationOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	inv := w.invited(g, "user_bob", "user_dee", team.InvitingPending)
	return w.answer(), AnswerInvitationInput{User: "user_dee", ID: inv.ID, Event: "decline"}, func(t *testing.T, out AnswerInvitationOutput) {
		t.Helper()
		if out.Group.ID != "" {
			t.Errorf("a declined invitation shows no group: %+v", out.Group)
		}
		if kept := w.invitations[inv.ID]; kept.State != team.InvitingDeclined || !kept.RespondedAt.After(epoch) {
			t.Errorf("the invitation kept: %+v", kept)
		}
		if w.role(g.ID, "user_dee") != "" || len(w.notices) != 0 {
			t.Errorf("dee is %q, notices %q", w.role(g.ID, "user_dee"), w.notices)
		}
	}
}

// givenAnswerInvitationRefusesTheInvitationOfOthers builds AnswerInvitation for "refuses the invitation of others": the use case and the input it
// refuses with team.ErrNoInvitation.
func givenAnswerInvitationRefusesTheInvitationOfOthers(t *testing.T) (AnswerInvitation, AnswerInvitationInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	inv := w.invited(g, "user_bob", "user_dee", team.InvitingPending)
	return w.answer(), AnswerInvitationInput{User: "user_eve", ID: inv.ID, Event: "accept"}
}

// givenAnswerInvitationRefusesAnInvitationAlreadyAnswered builds AnswerInvitation for "refuses an invitation already answered": the use case and the input it
// refuses with team.ErrNotPending.
func givenAnswerInvitationRefusesAnInvitationAlreadyAnswered(t *testing.T) (AnswerInvitation, AnswerInvitationInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	inv := w.invited(g, "user_bob", "user_dee", team.InvitingExpired)
	return w.answer(), AnswerInvitationInput{User: "user_dee", ID: inv.ID, Event: "accept"}
}

// givenAnswerInvitationRefusesAnInvitationIntoADeletedGroup builds AnswerInvitation for "refuses an invitation into a deleted group": the use case and the input it
// refuses with team.ErrNotFound.
func givenAnswerInvitationRefusesAnInvitationIntoADeletedGroup(t *testing.T) (AnswerInvitation, AnswerInvitationInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	inv := w.invited(g, "user_bob", "user_dee", team.InvitingPending)
	delete(w.groups, g.ID)
	return w.answer(), AnswerInvitationInput{User: "user_dee", ID: inv.ID, Event: "accept"}
}
