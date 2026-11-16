// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"
)

// execute is GetTask's body: a task the member does not see is not found.
func (h *GetTaskHandler) execute(ctx context.Context, in *GetTaskInput) (GetTaskOutput, error) {
	t, _, err := load(ctx, h.taskStore, h.groups, in.ID, in.User)
	if err != nil {
		return GetTaskOutput{}, err
	}
	return GetTaskOutput{Task: t}, nil
}
