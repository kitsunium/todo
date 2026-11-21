// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"cmp"
	"context"
	"slices"

	"github.com/kitsunium/todo/groups/team"
)

// execute is ListGroupInvitations's body: the invitations of the group still
// pending, the newest first, each naming who is invited and who invited, never
// an address. It is the rule of Invite: a group the member is not in is not
// found, and a plain member is refused.
func (h *ListGroupInvitationsHandler) execute(ctx context.Context, in *ListGroupInvitationsInput) (ListGroupInvitationsOutput, error) {
	g, found, err := h.groups.Get(ctx, in.ID)
	if err != nil {
		return ListGroupInvitationsOutput{}, err
	}
	role, err := roleIn(g, found, in.User)
	if err != nil {
		return ListGroupInvitationsOutput{}, err
	}
	if !team.Manages(role) {
		return ListGroupInvitationsOutput{}, team.ErrNotManager
	}
	waiting, err := h.invitations.OfGroup(ctx, g.ID)
	if err != nil {
		return ListGroupInvitationsOutput{}, err
	}
	out := ListGroupInvitationsOutput{Invitations: []PendingInvitation{}}
	for _, inv := range waiting {
		if inv.State != team.InvitingPending {
			continue
		}
		invitee, err := personOf(ctx, h.people, inv.InviteeID)
		if err != nil {
			return ListGroupInvitationsOutput{}, err
		}
		inviter, err := personOf(ctx, h.people, inv.InviterID)
		if err != nil {
			return ListGroupInvitationsOutput{}, err
		}
		out.Invitations = append(out.Invitations, PendingInvitation{ID: inv.ID, User: invitee, Inviter: inviter, CreatedAt: inv.CreatedAt})
	}
	slices.SortFunc(out.Invitations, func(a, b PendingInvitation) int {
		return cmp.Or(b.CreatedAt.Compare(a.CreatedAt), cmp.Compare(b.ID, a.ID))
	})
	return out, nil
}
