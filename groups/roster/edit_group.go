// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"

	"github.com/kitsunium/todo/groups/team"
)

// execute is EditGroup's body: its owner and admins change what is sent, a
// name cleaned as at creation, a color of the palette (an empty one is not
// one: only creation picks).
func (h *EditGroupHandler) execute(ctx context.Context, in *EditGroupInput) (EditGroupOutput, error) {
	g, found, err := h.groups.Get(ctx, in.ID)
	if err != nil {
		return EditGroupOutput{}, err
	}
	role, err := roleIn(g, found, in.User)
	if err != nil {
		return EditGroupOutput{}, err
	}
	if !team.Manages(role) {
		return EditGroupOutput{}, team.ErrNotManager
	}
	if in.Name != nil {
		if g.Name, err = team.CleanName(*in.Name); err != nil {
			return EditGroupOutput{}, err
		}
	}
	if in.Color != nil {
		if *in.Color == "" {
			return EditGroupOutput{}, team.ErrBadColor
		}
		if g.Color, err = team.ColorFor(*in.Color, g.Name); err != nil {
			return EditGroupOutput{}, err
		}
	}
	if g.UpdatedAt, err = h.clock.Now(ctx); err != nil {
		return EditGroupOutput{}, err
	}
	if err := h.groups.Save(ctx, g); err != nil {
		return EditGroupOutput{}, err
	}
	v, err := present(ctx, h.people, in.User, g)
	return EditGroupOutput{Group: v}, err
}
