// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"testing"

	"github.com/kitsunium/todo/contacts/link"
)

// givenCancelInviteWithdrawsAnInvitation builds CancelInvite for "withdraws an invitation": the use case, its input, and
// what its output must be.
func givenCancelInviteWithdrawsAnInvitation(t *testing.T) (CancelInvite, CancelInviteInput, func(*testing.T, CancelInviteOutput)) {
	t.Helper()
	w := newWorld()
	i := w.invited("user_ada", "zoe@example.com")
	keep := w.invited("user_ada", "kim@example.com")
	return NewCancelInvite(w.invitations()), CancelInviteInput{User: "user_ada", ID: i.ID}, func(t *testing.T, out CancelInviteOutput) {
		t.Helper()
		if out.Invitation.ID != i.ID || out.Invitation.Email != "zoe@example.com" {
			t.Errorf("the answer: %+v", out.Invitation)
		}
		if _, still := w.invites[i.ID]; still || len(w.invites) != 1 {
			t.Errorf("the invitations: %+v, want %s alone", w.invites, keep.ID)
		}
	}
}

// givenCancelInviteRefusesAnInvitationOfOthers builds CancelInvite for "refuses an invitation of others": the use case and the input it
// refuses with link.ErrNotFound.
func givenCancelInviteRefusesAnInvitationOfOthers(t *testing.T) (CancelInvite, CancelInviteInput) {
	t.Helper()
	w := newWorld()
	i := w.invited("user_ada", "zoe@example.com")
	return NewCancelInvite(w.invitations()), CancelInviteInput{User: "user_cy", ID: i.ID}
}

// givenCancelInviteRefusesAnInvitationAlreadyJoined builds CancelInvite for "refuses an invitation already joined": the use case and the input it
// refuses with link.ErrNotFound.
func givenCancelInviteRefusesAnInvitationAlreadyJoined(t *testing.T) (CancelInvite, CancelInviteInput) {
	t.Helper()
	w := newWorld()
	i := w.invited("user_ada", "zoe@example.com")
	i.Status = link.InviteJoined
	w.invites[i.ID] = i
	return NewCancelInvite(w.invitations()), CancelInviteInput{User: "user_ada", ID: i.ID}
}
