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
	t, err := h.taskStore.Get(ctx, in.ID)
	if err != nil {
		return MoveTaskOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return MoveTaskOutput{}, err
	}
	if err := move(ctx, &t, task.LifecycleEvent(in.Event), now); err != nil {
		return MoveTaskOutput{}, err
	}
	if err := h.taskStore.Save(ctx, t); err != nil {
		return MoveTaskOutput{}, err
	}
	if err := h.journal.Record(ctx, in.Event, t); err != nil {
		return MoveTaskOutput{}, err
	}
	return MoveTaskOutput{Task: t}, nil
}
