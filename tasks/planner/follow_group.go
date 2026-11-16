// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"
	"errors"
	"slices"

	"github.com/kitsunium/todo/tasks/task"
)

// execute is FollowGroup's body: of the tasks on the group's list, each read
// again before it is changed, a deleted group's go back to their owners'
// lists, and one assigned to a member who is no longer in the group, and no
// more its owner's or shared with, loses its assignee. Whoever is in the
// group when the message is heard decides, so hearing it twice changes
// nothing the first time did not.
func (h *FollowGroupHandler) execute(ctx context.Context, in *FollowGroupInput) (FollowGroupOutput, error) {
	all, err := h.taskStore.List(ctx)
	if err != nil {
		return FollowGroupOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return FollowGroupOutput{}, err
	}
	for _, listed := range all {
		if listed.GroupID != in.Group {
			continue
		}
		t, err := h.taskStore.Get(ctx, listed.ID)
		if errors.Is(err, task.ErrNotFound) {
			continue // deleted since it was listed
		}
		if err != nil {
			return FollowGroupOutput{}, err
		}
		if t.GroupID != in.Group {
			continue // moved since it was listed
		}
		changed := false
		if in.Deleted {
			t.GroupID, changed = "", true
		}
		if t.AssigneeID != "" && t.AssigneeID != t.OwnerID && !slices.Contains(t.SharedWith, t.AssigneeID) {
			stays := false
			if !in.Deleted {
				role, err := h.groups.RoleOf(ctx, in.Group, t.AssigneeID)
				if err != nil {
					return FollowGroupOutput{}, err
				}
				stays = role != ""
			}
			if !stays {
				t.AssigneeID, changed = "", true
			}
		}
		if !changed {
			continue
		}
		t.UpdatedAt = now
		if err := h.taskStore.Save(ctx, t); err != nil {
			return FollowGroupOutput{}, err
		}
	}
	return FollowGroupOutput{}, nil
}
