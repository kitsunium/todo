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

// execute is DeleteTask's body: its owner deletes a task, and so does an
// owner or an admin of its group; everyone who saw it is told. A member who
// sees it and is none of them may not; anyone else finds none.
func (h *DeleteTaskHandler) execute(ctx context.Context, in *DeleteTaskInput) (DeleteTaskOutput, error) {
	t, role, err := load(ctx, h.taskStore, h.groups, in.ID, in.User)
	if err != nil {
		return DeleteTaskOutput{}, err
	}
	if !manages(t, in.User, role) {
		return DeleteTaskOutput{}, task.ErrNotOwner
	}
	if err := h.taskStore.Delete(ctx, t.ID); err != nil {
		return DeleteTaskOutput{}, err
	}
	return DeleteTaskOutput{}, h.journal.Record(ctx, kindDeleted, t)
}
