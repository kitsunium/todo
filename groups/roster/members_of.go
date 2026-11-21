// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"
)

// execute is MembersOf's body: the ids of the group's members, in the order
// they joined; none for a group that does not exist.
func (h *MembersOfHandler) execute(ctx context.Context, in *MembersOfInput) (MembersOfOutput, error) {
	g, found, err := h.groups.Get(ctx, in.Group)
	if err != nil || !found {
		return MembersOfOutput{}, err
	}
	return MembersOfOutput{Users: g.MemberIDs()}, nil
}
