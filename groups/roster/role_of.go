// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"
)

// execute is RoleOf's body: the member's role, empty for a stranger and for
// a group that does not exist.
func (h *RoleOfHandler) execute(ctx context.Context, in *RoleOfInput) (RoleOfOutput, error) {
	g, found, err := h.groups.Get(ctx, in.Group)
	if err != nil || !found {
		return RoleOfOutput{}, err
	}
	return RoleOfOutput{Role: g.RoleOf(in.User)}, nil
}
