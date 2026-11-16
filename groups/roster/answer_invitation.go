// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"
	"errors"

	"github.com/kitsunium/todo/groups/team"
)

// execute is AnswerInvitation's body. An invitation of someone else's is not
// found; one no longer pending (answered, or expired) cannot be answered.
// Accepting adds the invitee to the group as a member, unless it was deleted
// meanwhile, and tells the members; declining only keeps the answer.
func (h *AnswerInvitationHandler) execute(ctx context.Context, in *AnswerInvitationInput) (AnswerInvitationOutput, error) {
	now, err := h.clock.Now(ctx)
	if err != nil {
		return AnswerInvitationOutput{}, err
	}
	// The invitation is the gate: it is answered once, whoever asks twice at
	// the same instant, and only then does the group change.
	inv, err := h.invitations.Change(ctx, in.ID, func(inv *team.Invitation) error {
		if inv.InviteeID != in.User {
			return team.ErrNoInvitation
		}
		to, err := team.NewInviting().Fire(ctx, inv, inv.State, team.InvitingEvent(in.Event))
		switch {
		case errors.Is(err, team.ErrInvitingNoTransition):
			return team.ErrNotPending
		case err != nil:
			return err
		}
		inv.State, inv.RespondedAt = to, now
		return nil
	})
	if err != nil {
		return AnswerInvitationOutput{}, err
	}
	if inv.State != team.InvitingAccepted {
		return AnswerInvitationOutput{}, nil
	}
	// The member is added to the group as it is at that instant: two invitees
	// who accept at once both join. A group deleted meanwhile is not found.
	g, err := h.groups.Change(ctx, inv.GroupID, func(g *team.Group) error {
		if g.RoleOf(in.User) == "" {
			g.Members = append(g.Members[:len(g.Members):len(g.Members)], team.Membership{UserID: in.User, Role: team.RoleMember, JoinedAt: now})
			g.UpdatedAt = now
		}
		return nil
	})
	if err != nil {
		return AnswerInvitationOutput{}, err
	}
	actor, err := personOf(ctx, h.people, in.User)
	if err != nil {
		return AnswerInvitationOutput{}, err
	}
	if err := h.notices.Tell(ctx, kindJoined, g, actor, g.MemberIDs()); err != nil {
		return AnswerInvitationOutput{}, err
	}
	v, err := present(ctx, h.people, in.User, g)
	return AnswerInvitationOutput{Group: v}, err
}
