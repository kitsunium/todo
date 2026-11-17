// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"
	"errors"

	"github.com/kitsunium/todo/groups/team"
)

// execute is WithdrawInvitation's body: an owner or an admin of the group
// takes back one of its invitations still pending. The invitation is the
// gate, as it is for AnswerInvitation: it moves once, under the store's
// Change, so a withdrawal and an acceptance at the same instant leave one
// winner. Whoever comes second finds the invitation no longer pending, which
// is no invitation to withdraw (not found), as it is for one of another group,
// one already answered and one nobody made.
func (h *WithdrawInvitationHandler) execute(ctx context.Context, in *WithdrawInvitationInput) (WithdrawInvitationOutput, error) {
	g, found, err := h.groups.Get(ctx, in.ID)
	if err != nil {
		return WithdrawInvitationOutput{}, err
	}
	role, err := roleIn(g, found, in.User)
	if err != nil {
		return WithdrawInvitationOutput{}, err
	}
	if !team.Manages(role) {
		return WithdrawInvitationOutput{}, team.ErrNotManager
	}
	_, err = h.invitations.Change(ctx, in.InvitationID, func(inv *team.Invitation) error {
		if inv.GroupID != g.ID {
			return team.ErrNoInvitation
		}
		to, err := team.NewInviting().Fire(ctx, inv, inv.State, team.InvitingOnWithdraw)
		switch {
		case errors.Is(err, team.ErrInvitingNoTransition):
			return team.ErrNoInvitation
		case err != nil:
			return err
		}
		inv.State = to
		return nil
	})
	return WithdrawInvitationOutput{}, err
}
