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
	t, _, err := load(ctx, h.taskStore, h.groups, in.ID, in.User)
	if err != nil {
		return UnshareTaskOutput{}, err
	}
	if t.OwnerID != in.User && in.With != in.User {
		return UnshareTaskOutput{}, task.ErrNotOwner
	}
	t.SharedWith = slices.DeleteFunc(slices.Clone(t.SharedWith), func(u string) bool { return u == in.With })
	if t.UpdatedAt, err = h.clock.Now(ctx); err != nil {
		return UnshareTaskOutput{}, err
	}
	if err := h.taskStore.Save(ctx, t); err != nil {
		return UnshareTaskOutput{}, err
	}
	return UnshareTaskOutput{Task: t}, nil
}
