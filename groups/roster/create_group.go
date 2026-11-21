// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"

	"github.com/kitsunium/todo/groups/team"
)

// execute is CreateGroup's body: the name is cleaned and the color checked,
// or picked from the name, and the creator is the group's owner, alone in it.
func (h *CreateGroupHandler) execute(ctx context.Context, in *CreateGroupInput) (CreateGroupOutput, error) {
	name, err := team.CleanName(in.Name)
	if err != nil {
		return CreateGroupOutput{}, err
	}
	color, err := team.ColorFor(in.Color, name)
	if err != nil {
		return CreateGroupOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return CreateGroupOutput{}, err
	}
	id, err := h.groups.NewID(ctx)
	if err != nil {
		return CreateGroupOutput{}, err
	}
	g := team.Group{
		ID: id, Name: name, Color: color, CreatedAt: now, UpdatedAt: now,
		Members: []team.Membership{{UserID: in.User, Role: team.RoleOwner, JoinedAt: now}},
	}
	if err := h.groups.Save(ctx, g); err != nil {
		return CreateGroupOutput{}, err
	}
	v, err := present(ctx, h.people, in.User, g)
	return CreateGroupOutput{Group: v}, err
}
