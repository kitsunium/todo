// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"

	"github.com/kitsunium/todo/tasks/task"
)

// execute is CountTasks's body: a count per state, every state named even
// at zero, and the open tasks past their due date.
func (h *CountTasksHandler) execute(ctx context.Context, in *CountTasksInput) (CountTasksOutput, error) {
	now, err := h.clock.Now(ctx)
	if err != nil {
		return CountTasksOutput{}, err
	}
	all, err := h.taskStore.List(ctx)
	if err != nil {
		return CountTasksOutput{}, err
	}
	out := CountTasksOutput{Total: len(all), Counts: map[string]int{
		string(task.LifecycleOpen): 0, string(task.LifecycleDone): 0, string(task.LifecycleArchived): 0,
	}}
	for _, t := range all {
		out.Counts[string(t.State)]++
		if t.Overdue(now) {
			out.Overdue++
		}
	}
	return out, nil
}
