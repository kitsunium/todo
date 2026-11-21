// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"

	"github.com/kitsunium/todo/groups/team"
)

// execute is Invite's body: an owner or an admin asks one of their contacts,
// who is not in the group yet and has no invitation waiting into it. The
// invitee is told who asks.
func (h *InviteHandler) execute(ctx context.Context, in *InviteInput) (InviteOutput, error) {
	g, found, err := h.groups.Get(ctx, in.ID)
	if err != nil {
		return InviteOutput{}, err
	}
	role, err := roleIn(g, found, in.User)
	if err != nil {
		return InviteOutput{}, err
	}
	if !team.Manages(role) {
		return InviteOutput{}, team.ErrNotManager
	}
	if in.Invitee == in.User || g.RoleOf(in.Invitee) != "" {
		return InviteOutput{}, team.ErrAlreadyMember
	}
	contact, err := h.circle.AreContacts(ctx, in.User, in.Invitee)
	if err != nil {
		return InviteOutput{}, err
	}
	if !contact {
		return InviteOutput{}, team.ErrNotContact
	}
	waiting, err := h.invitations.OfGroup(ctx, g.ID)
	if err != nil {
		return InviteOutput{}, err
	}
	for _, inv := range waiting {
		if inv.InviteeID == in.Invitee && inv.State == team.InvitingPending {
			return InviteOutput{}, team.ErrAlreadyInvited
		}
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return InviteOutput{}, err
	}
	id, err := h.invitations.NewID(ctx)
	if err != nil {
		return InviteOutput{}, err
	}
	inv := team.Invitation{ID: id, GroupID: g.ID, InviterID: in.User, InviteeID: in.Invitee, State: team.InvitingInitial, CreatedAt: now}
	if err := h.invitations.Save(ctx, inv); err != nil {
		return InviteOutput{}, err
	}
	inviter, err := personOf(ctx, h.people, in.User)
	if err != nil {
		return InviteOutput{}, err
	}
	invitee, err := personOf(ctx, h.people, in.Invitee)
	if err != nil {
		return InviteOutput{}, err
	}
	if err := h.notices.Tell(ctx, kindInvited, g, inviter, []string{in.Invitee}); err != nil {
		return InviteOutput{}, err
	}
	return InviteOutput{Invitation: InvitationView{ID: inv.ID, Group: refOf(g), Inviter: inviter, Invitee: invitee, State: inv.State, CreatedAt: inv.CreatedAt}}, nil
}
