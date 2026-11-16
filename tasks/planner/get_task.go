// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"

	"github.com/kitsunium/todo/tasks/task"
)

// execute is GetTask's body: a task the member does not see is not found.
func (h *GetTaskHandler) execute(ctx context.Context, in *GetTaskInput) (GetTaskOutput, error) {
	t, err := h.taskStore.Get(ctx, in.ID)
	if err != nil {
		return GetTaskOutput{}, err
	}
	if !t.SeenBy(in.User) {
		return GetTaskOutput{}, task.ErrNotFound
	}
	return GetTaskOutput{Task: t}, nil
}
