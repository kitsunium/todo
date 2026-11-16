// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"
	"slices"

	"github.com/kitsunium/todo/tasks/task"
)

// execute is ListTasks's body: the owner's tasks of the state asked for, or
// every one that is not archived, the most urgent first.
func (h *ListTasksHandler) execute(ctx context.Context, in *ListTasksInput) (ListTasksOutput, error) {
	all, err := h.taskStore.List(ctx)
	if err != nil {
		return ListTasksOutput{}, err
	}
	out := make([]task.Task, 0, len(all))
	for _, t := range all {
		if t.OwnerID != in.Owner {
			continue
		}
		if in.State == "" && t.State != task.LifecycleArchived || in.State != "" && string(t.State) == in.State {
			out = append(out, t)
		}
	}
	slices.SortFunc(out, byUrgency)
	return ListTasksOutput{Tasks: out}, nil
}
