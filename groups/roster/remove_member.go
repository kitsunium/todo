// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"
	"slices"

	"github.com/kitsunium/todo/groups/team"
)

// execute is RemoveMember's body. Anyone may leave but the owner, who deletes
// the group instead; the owner removes anyone, an admin removes members, a
// member nobody. Whoever leaves is not told, the others are; whoever is
// removed is told with them.
func (h *RemoveMemberHandler) execute(ctx context.Context, in *RemoveMemberInput) (RemoveMemberOutput, error) {
	now, err := h.clock.Now(ctx)
	if err != nil {
		return RemoveMemberOutput{}, err
	}
	leaving := in.Member == in.User
	g, err := h.groups.Change(ctx, in.ID, func(g *team.Group) error {
		role, err := roleIn(*g, true, in.User)
		if err != nil {
			return err
		}
		switch target := g.RoleOf(in.Member); {
		case target == "":
			return team.ErrNoMember
		case leaving && role == team.RoleOwner:
			return team.ErrOwnerStays
		case !leaving && (target == team.RoleOwner || role == team.RoleMember || (role == team.RoleAdmin && target == team.RoleAdmin)):
			return team.ErrCannotRemove
		}
		g.Members = slices.DeleteFunc(slices.Clone(g.Members), func(m team.Membership) bool { return m.UserID == in.Member })
		g.UpdatedAt = now
		return nil
	})
	if err != nil {
		return RemoveMemberOutput{}, err
	}
	actor, err := personOf(ctx, h.people, in.User)
	if err != nil {
		return RemoveMemberOutput{}, err
	}
	if leaving {
		return RemoveMemberOutput{}, h.notices.Tell(ctx, kindLeft, g, actor, g.MemberIDs())
	}
	return RemoveMemberOutput{}, h.notices.Tell(ctx, kindRemoved, g, actor, append(g.MemberIDs(), in.Member))
}
