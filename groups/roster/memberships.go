// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"github.com/kitsunium/todo/groups/team"
)

// execute is Memberships's body: the groups the member is in, by name, then
// by id, each with the role they have in it.
func (h *MembershipsHandler) execute(ctx context.Context, in *MembershipsInput) (MembershipsOutput, error) {
	gs, err := h.groups.Of(ctx, in.User)
	if err != nil {
		return MembershipsOutput{}, err
	}
	slices.SortFunc(gs, func(a, b team.Group) int {
		return cmp.Or(cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), cmp.Compare(a.ID, b.ID))
	})
	out := MembershipsOutput{Memberships: make([]Belonging, 0, len(gs))}
	for _, g := range gs {
		out.Memberships = append(out.Memberships, Belonging{Group: refOf(g), Role: g.RoleOf(in.User)})
	}
	return out, nil
}
