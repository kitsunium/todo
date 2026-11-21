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

// errUnchanged ends a change that has nothing to write: the store keeps the
// task as it is.
var errUnchanged = errors.New("unchanged")

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
		// Who is still in the group is asked before the change; the change
		// holds it to the task as it is then.
		assignee, stays := t.AssigneeID, false
		if assignee != "" && !in.Deleted {
			role, err := h.groups.RoleOf(ctx, in.Group, assignee)
			if err != nil {
				return FollowGroupOutput{}, err
			}
			stays = role != ""
		}
		_, err = h.taskStore.Change(ctx, t.ID, func(t *task.Task) error {
			if t.GroupID != in.Group {
				return errUnchanged // moved since it was listed
			}
			changed := false
			if in.Deleted {
				t.GroupID, changed = "", true
			}
			if t.AssigneeID != "" && t.AssigneeID != t.OwnerID && !slices.Contains(t.SharedWith, t.AssigneeID) && !(stays && t.AssigneeID == assignee) {
				t.AssigneeID, changed = "", true
			}
			if !changed {
				return errUnchanged
			}
			t.UpdatedAt = now
			return nil
		})
		if err != nil && !errors.Is(err, errUnchanged) && !errors.Is(err, task.ErrNotFound) {
			return FollowGroupOutput{}, err
		}
	}
	return FollowGroupOutput{}, nil
}
