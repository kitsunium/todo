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
	inv, found, err := h.invitations.Get(ctx, in.ID)
	if err != nil {
		return AnswerInvitationOutput{}, err
	}
	if !found || inv.InviteeID != in.User {
		return AnswerInvitationOutput{}, team.ErrNoInvitation
	}
	to, err := team.NewInviting().Fire(ctx, &inv, inv.State, team.InvitingEvent(in.Event))
	switch {
	case errors.Is(err, team.ErrInvitingNoTransition):
		return AnswerInvitationOutput{}, team.ErrNotPending
	case err != nil:
		return AnswerInvitationOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return AnswerInvitationOutput{}, err
	}
	var g team.Group
	if to == team.InvitingAccepted {
		var found bool
		if g, found, err = h.groups.Get(ctx, inv.GroupID); err != nil {
			return AnswerInvitationOutput{}, err
		}
		if !found {
			return AnswerInvitationOutput{}, team.ErrNotFound
		}
		if g.RoleOf(in.User) == "" {
			g.Members = append(g.Members[:len(g.Members):len(g.Members)], team.Membership{UserID: in.User, Role: team.RoleMember, JoinedAt: now})
			g.UpdatedAt = now
			if err := h.groups.Save(ctx, g); err != nil {
				return AnswerInvitationOutput{}, err
			}
		}
	}
	inv.State, inv.RespondedAt = to, now
	if err := h.invitations.Save(ctx, inv); err != nil {
		return AnswerInvitationOutput{}, err
	}
	if to != team.InvitingAccepted {
		return AnswerInvitationOutput{}, nil
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
