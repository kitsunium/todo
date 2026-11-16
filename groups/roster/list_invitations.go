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

// execute is ListInvitations's body: the invitations still pending for the
// member, into a group that still exists, the newest first.
func (h *ListInvitationsHandler) execute(ctx context.Context, in *ListInvitationsInput) (ListInvitationsOutput, error) {
	received, err := h.invitations.Of(ctx, in.User)
	if err != nil {
		return ListInvitationsOutput{}, err
	}
	out := ListInvitationsOutput{Invitations: []InvitationView{}}
	for _, inv := range received {
		if inv.State != team.InvitingPending {
			continue
		}
		g, found, err := h.groups.Get(ctx, inv.GroupID)
		if err != nil {
			return ListInvitationsOutput{}, err
		}
		if !found {
			continue
		}
		inviter, err := personOf(ctx, h.people, inv.InviterID)
		if err != nil {
			return ListInvitationsOutput{}, err
		}
		out.Invitations = append(out.Invitations, InvitationView{ID: inv.ID, Group: refOf(g), Inviter: inviter, State: inv.State, CreatedAt: inv.CreatedAt})
	}
	slices.SortFunc(out.Invitations, func(a, b InvitationView) int {
		return cmp.Or(b.CreatedAt.Compare(a.CreatedAt), cmp.Compare(b.ID, a.ID))
	})
	return out, nil
}
