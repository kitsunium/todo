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
	var name string
	if in.Name != nil {
		var err error
		if name, err = team.CleanName(*in.Name); err != nil {
			return EditGroupOutput{}, err
		}
	}
	if in.Color != nil {
		if *in.Color == "" {
			return EditGroupOutput{}, team.ErrBadColor
		}
		if _, err := team.ColorFor(*in.Color, ""); err != nil {
			return EditGroupOutput{}, err
		}
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return EditGroupOutput{}, err
	}
	g, err := h.groups.Change(ctx, in.ID, func(g *team.Group) error {
		role, err := roleIn(*g, true, in.User)
		if err != nil {
			return err
		}
		if !team.Manages(role) {
			return team.ErrNotManager
		}
		if in.Name != nil {
			g.Name = name
		}
		if in.Color != nil {
			g.Color = *in.Color
		}
		g.UpdatedAt = now
		return nil
	})
	if err != nil {
		return EditGroupOutput{}, err
	}
	v, err := present(ctx, h.people, in.User, g)
	return EditGroupOutput{Group: v}, err
}
