// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"

	"github.com/kitsunium/todo/tasks/task"
)

// kindDeleted is what the journal is told of a task its owner deleted.
const kindDeleted = "deleted"

// execute is DeleteTask's body: its owner deletes a task, and those it was
// shared with are told. A member it is shared with may not; anyone else
// finds none.
func (h *DeleteTaskHandler) execute(ctx context.Context, in *DeleteTaskInput) (DeleteTaskOutput, error) {
	t, err := h.taskStore.Get(ctx, in.ID)
	if err != nil {
		return DeleteTaskOutput{}, err
	}
	switch {
	case !t.SeenBy(in.User):
		return DeleteTaskOutput{}, task.ErrNotFound
	case t.OwnerID != in.User:
		return DeleteTaskOutput{}, task.ErrNotOwner
	}
	if err := h.taskStore.Delete(ctx, t.ID); err != nil {
		return DeleteTaskOutput{}, err
	}
	return DeleteTaskOutput{}, h.journal.Record(ctx, kindDeleted, t)
}
