// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"
)

// execute is GetGroup's body: a group the member is not in is not found.
func (h *GetGroupHandler) execute(ctx context.Context, in *GetGroupInput) (GetGroupOutput, error) {
	g, found, err := h.groups.Get(ctx, in.ID)
	if err != nil {
		return GetGroupOutput{}, err
	}
	if _, err := roleIn(g, found, in.User); err != nil {
		return GetGroupOutput{}, err
	}
	v, err := present(ctx, h.people, in.User, g)
	return GetGroupOutput{Group: v}, err
}
