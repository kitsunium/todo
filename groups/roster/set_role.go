// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"

	"github.com/kitsunium/todo/groups/team"
)

// execute is SetRole's body: the role is read first, then only the owner
// changes one, of another member, never the owner's own. The member whose
// role changed is told.
func (h *SetRoleHandler) execute(ctx context.Context, in *SetRoleInput) (SetRoleOutput, error) {
	role, err := team.ParseRole(in.Role)
	if err != nil {
		return SetRoleOutput{}, err
	}
	g, found, err := h.groups.Get(ctx, in.ID)
	if err != nil {
		return SetRoleOutput{}, err
	}
	caller, err := roleIn(g, found, in.User)
	if err != nil {
		return SetRoleOutput{}, err
	}
	switch target := g.RoleOf(in.Member); {
	case caller != team.RoleOwner:
		return SetRoleOutput{}, team.ErrNotOwner
	case target == "":
		return SetRoleOutput{}, team.ErrNoMember
	case target == team.RoleOwner:
		return SetRoleOutput{}, team.ErrOwnerRole
	}
	if g.UpdatedAt, err = h.clock.Now(ctx); err != nil {
		return SetRoleOutput{}, err
	}
	members := make([]team.Membership, len(g.Members))
	for i, m := range g.Members {
		if m.UserID == in.Member {
			m.Role = role
		}
		members[i] = m
	}
	g.Members = members
	if err := h.groups.Save(ctx, g); err != nil {
		return SetRoleOutput{}, err
	}
	actor, err := personOf(ctx, h.people, in.User)
	if err != nil {
		return SetRoleOutput{}, err
	}
	if err := h.notices.Tell(ctx, kindRole, g, actor, []string{in.Member}); err != nil {
		return SetRoleOutput{}, err
	}
	v, err := present(ctx, h.people, in.User, g)
	return SetRoleOutput{Group: v}, err
}
