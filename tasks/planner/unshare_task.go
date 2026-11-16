// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"
	"slices"

	"github.com/kitsunium/todo/tasks/task"
)

// execute is UnshareTask's body: the owner ends the sharing with a member,
// or that member leaves the task; a member it is shared with removes
// nobody else.
func (h *UnshareTaskHandler) execute(ctx context.Context, in *UnshareTaskInput) (UnshareTaskOutput, error) {
	now, err := h.clock.Now(ctx)
	if err != nil {
		return UnshareTaskOutput{}, err
	}
	t, err := change(ctx, h.taskStore, h.groups, in.ID, in.User, func(t *task.Task, _ string) error {
		if t.OwnerID != in.User && in.With != in.User {
			return task.ErrNotOwner
		}
		t.SharedWith = slices.DeleteFunc(slices.Clone(t.SharedWith), func(u string) bool { return u == in.With })
		t.UpdatedAt = now
		return nil
	})
	return UnshareTaskOutput{Task: t}, err
}
