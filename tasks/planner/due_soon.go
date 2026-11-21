// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"cmp"
	"context"
	"slices"

	"github.com/kitsunium/todo/tasks/task"
)

// execute is DueSoon's body: the open tasks whose due date is ahead of now
// by Within at most, the soonest first. A task already overdue is not due
// soon: its reminder's time has passed.
func (h *DueSoonHandler) execute(ctx context.Context, in *DueSoonInput) (DueSoonOutput, error) {
	now, err := h.clock.Now(ctx)
	if err != nil {
		return DueSoonOutput{}, err
	}
	all, err := h.taskStore.List(ctx)
	if err != nil {
		return DueSoonOutput{}, err
	}
	out := make([]task.Task, 0, len(all))
	for _, t := range all {
		if t.State == task.LifecycleOpen && !t.Due.IsZero() && !t.Due.Before(now) && t.Due.Sub(now) <= in.Within {
			out = append(out, t)
		}
	}
	slices.SortFunc(out, func(a, b task.Task) int { return cmp.Or(a.Due.Compare(b.Due), cmp.Compare(a.ID, b.ID)) })
	return DueSoonOutput{Tasks: out}, nil
}
