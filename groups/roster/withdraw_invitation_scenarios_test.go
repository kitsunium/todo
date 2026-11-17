// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenWithdrawInvitationWithdrawsAnInvitation builds WithdrawInvitation for "withdraws an invitation": the use case, its input, and
// what its output must be.
func givenWithdrawInvitationWithdrawsAnInvitation(t *testing.T) (WithdrawInvitation, WithdrawInvitationInput, func(*testing.T, WithdrawInvitationOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	inv := w.invited(g, "user_ada", "user_dee", team.InvitingPending)
	keep := w.invited(g, "user_ada", "user_eve", team.InvitingPending)
	return w.withdraw(), WithdrawInvitationInput{User: "user_bob", ID: g.ID, InvitationID: inv.ID}, func(t *testing.T, _ WithdrawInvitationOutput) {
		t.Helper()
		if got := w.invitations[inv.ID]; got.State != team.InvitingWithdrawn {
			t.Errorf("the invitation: %+v, want it withdrawn", got)
		}
		if got := w.invitations[keep.ID]; got.State != team.InvitingPending {
			t.Errorf("the other invitation: %+v, want it left alone", got)
		}
		if len(w.notices) != 0 {
			t.Errorf("the notices: %q, want none", w.notices)
		}
	}
}

// givenWithdrawInvitationRefusesTheGroupOfOthers builds WithdrawInvitation for "refuses the group of others": the use case and the input it
// refuses with team.ErrNotFound.
func givenWithdrawInvitationRefusesTheGroupOfOthers(t *testing.T) (WithdrawInvitation, WithdrawInvitationInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	inv := w.invited(g, "user_ada", "user_dee", team.InvitingPending)
	return w.withdraw(), WithdrawInvitationInput{User: "user_eve", ID: g.ID, InvitationID: inv.ID}
}

// givenWithdrawInvitationRefusesAPlainMemberWhoWithdraws builds WithdrawInvitation for "refuses a plain member who withdraws": the use case and the input it
// refuses with team.ErrNotManager.
func givenWithdrawInvitationRefusesAPlainMemberWhoWithdraws(t *testing.T) (WithdrawInvitation, WithdrawInvitationInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	inv := w.invited(g, "user_ada", "user_dee", team.InvitingPending)
	return w.withdraw(), WithdrawInvitationInput{User: "user_cy", ID: g.ID, InvitationID: inv.ID}
}

// givenWithdrawInvitationRefusesAnInvitationOfAnotherGroup builds WithdrawInvitation for "refuses an invitation of another group": the use case and the input it
// refuses with team.ErrNoInvitation.
func givenWithdrawInvitationRefusesAnInvitationOfAnotherGroup(t *testing.T) (WithdrawInvitation, WithdrawInvitationInput) {
	t.Helper()
	w := newWorld()
	g, other := w.core(), w.core()
	inv := w.invited(other, "user_ada", "user_dee", team.InvitingPending)
	return w.withdraw(), WithdrawInvitationInput{User: "user_ada", ID: g.ID, InvitationID: inv.ID}
}

// givenWithdrawInvitationRefusesAnInvitationAlreadyAnswered builds WithdrawInvitation for "refuses an invitation already answered": the use case and the input it
// refuses with team.ErrNoInvitation.
func givenWithdrawInvitationRefusesAnInvitationAlreadyAnswered(t *testing.T) (WithdrawInvitation, WithdrawInvitationInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	inv := w.invited(g, "user_ada", "user_dee", team.InvitingAccepted)
	return w.withdraw(), WithdrawInvitationInput{User: "user_ada", ID: g.ID, InvitationID: inv.ID}
}

// givenWithdrawInvitationRefusesAnUnknownInvitation builds WithdrawInvitation for "refuses an unknown invitation": the use case and the input it
// refuses with team.ErrNoInvitation.
func givenWithdrawInvitationRefusesAnUnknownInvitation(t *testing.T) (WithdrawInvitation, WithdrawInvitationInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.withdraw(), WithdrawInvitationInput{User: "user_ada", ID: g.ID, InvitationID: "invitation_none"}
}
