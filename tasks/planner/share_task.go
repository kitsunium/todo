// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"
	"slices"

	"github.com/kitsunium/todo/tasks/task"
)

// kindShared is what the journal is told of a task shared with one more
// member.
const kindShared = "shared"

// execute is ShareTask's body: the owner shares a task with one of their
// contacts. Sharing it again with the same member changes nothing.
func (h *ShareTaskHandler) execute(ctx context.Context, in *ShareTaskInput) (ShareTaskOutput, error) {
	t, _, err := load(ctx, h.taskStore, h.groups, in.ID, in.User)
	if err != nil {
		return ShareTaskOutput{}, err
	}
	if t.OwnerID != in.User {
		return ShareTaskOutput{}, task.ErrNotOwner
	}
	if in.With == t.OwnerID || slices.Contains(t.SharedWith, in.With) {
		return ShareTaskOutput{Task: t}, nil
	}
	friends, err := h.circle.AreContacts(ctx, in.User, in.With)
	if err != nil {
		return ShareTaskOutput{}, err
	}
	if !friends {
		return ShareTaskOutput{}, task.ErrNotContact
	}
	t.SharedWith = append(slices.Clone(t.SharedWith), in.With)
	if t.UpdatedAt, err = h.clock.Now(ctx); err != nil {
		return ShareTaskOutput{}, err
	}
	if err := h.taskStore.Save(ctx, t); err != nil {
		return ShareTaskOutput{}, err
	}
	if err := h.journal.Record(ctx, kindShared, t); err != nil {
		return ShareTaskOutput{}, err
	}
	return ShareTaskOutput{Task: t}, nil
}
