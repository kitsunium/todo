// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"

	"github.com/kitsunium/todo/tasks/task"
)

// execute is MoveTask's body: the task is read, moved by its lifecycle,
// kept, and the journal told of the event. An event the task's state does
// not answer, or one the lifecycle does not know, is ErrNotAllowed.
func (h *MoveTaskHandler) execute(ctx context.Context, in *MoveTaskInput) (MoveTaskOutput, error) {
	now, err := h.clock.Now(ctx)
	if err != nil {
		return MoveTaskOutput{}, err
	}
	// A task they do not see: as if there were none.
	t, err := change(ctx, h.taskStore, h.groups, in.ID, in.Owner, func(t *task.Task, _ string) error {
		return move(ctx, t, task.LifecycleEvent(in.Event), now)
	})
	if err != nil {
		return MoveTaskOutput{}, err
	}
	if err := h.journal.Record(ctx, in.Event, t); err != nil {
		return MoveTaskOutput{}, err
	}
	return MoveTaskOutput{Task: t}, nil
}
