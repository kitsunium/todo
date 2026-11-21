// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"

	"github.com/kitsunium/todo/groups/team"
)

// execute is DeleteGroup's body: only the owner deletes, the invitations into
// the group go with it, and the members, the owner excepted, are told.
func (h *DeleteGroupHandler) execute(ctx context.Context, in *DeleteGroupInput) (DeleteGroupOutput, error) {
	g, found, err := h.groups.Get(ctx, in.ID)
	if err != nil {
		return DeleteGroupOutput{}, err
	}
	role, err := roleIn(g, found, in.User)
	if err != nil {
		return DeleteGroupOutput{}, err
	}
	if role != team.RoleOwner {
		return DeleteGroupOutput{}, team.ErrNotOwner
	}
	invitations, err := h.invitations.OfGroup(ctx, g.ID)
	if err != nil {
		return DeleteGroupOutput{}, err
	}
	for _, inv := range invitations {
		if err := h.invitations.Delete(ctx, inv.ID); err != nil {
			return DeleteGroupOutput{}, err
		}
	}
	if err := h.groups.Delete(ctx, g.ID); err != nil {
		return DeleteGroupOutput{}, err
	}
	actor, err := personOf(ctx, h.people, in.User)
	if err != nil {
		return DeleteGroupOutput{}, err
	}
	return DeleteGroupOutput{}, h.notices.Tell(ctx, kindDeleted, g, actor, g.MemberIDs())
}
