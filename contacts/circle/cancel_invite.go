// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"context"

	"github.com/kitsunium/todo/contacts/link"
)

// execute is CancelInvite's body: an invitation of the member's own, still
// waiting, is deleted. Another member's, or one already joined, is not found.
func (h *CancelInviteHandler) execute(ctx context.Context, in *CancelInviteInput) (CancelInviteOutput, error) {
	i, found, err := h.invites.Get(ctx, in.ID)
	if err != nil {
		return CancelInviteOutput{}, err
	}
	if !found || i.InviterID != in.User || i.Status != link.InvitePending {
		return CancelInviteOutput{}, link.ErrNotFound
	}
	if err := h.invites.Delete(ctx, i.ID); err != nil {
		return CancelInviteOutput{}, err
	}
	return CancelInviteOutput{Invitation: invitationOf(i)}, nil
}
