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

// execute is ListGroups's body: the member's groups by name, then by id.
func (h *ListGroupsHandler) execute(ctx context.Context, in *ListGroupsInput) (ListGroupsOutput, error) {
	gs, err := h.groups.Of(ctx, in.User)
	if err != nil {
		return ListGroupsOutput{}, err
	}
	slices.SortFunc(gs, func(a, b team.Group) int {
		return cmp.Or(cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), cmp.Compare(a.ID, b.ID))
	})
	out := ListGroupsOutput{Groups: make([]GroupView, 0, len(gs))}
	for _, g := range gs {
		v, err := present(ctx, h.people, in.User, g)
		if err != nil {
			return ListGroupsOutput{}, err
		}
		out.Groups = append(out.Groups, v)
	}
	return out, nil
}
