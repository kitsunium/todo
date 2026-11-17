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
	seen, _, err := load(ctx, h.taskStore, h.groups, in.ID, in.User)
	if err != nil {
		return ShareTaskOutput{}, err
	}
	if seen.OwnerID != in.User {
		return ShareTaskOutput{}, task.ErrNotOwner
	}
	if in.With == seen.OwnerID || slices.Contains(seen.SharedWith, in.With) {
		return ShareTaskOutput{Task: seen}, nil
	}
	friends, err := h.circle.AreContacts(ctx, in.User, in.With)
	if err != nil {
		return ShareTaskOutput{}, err
	}
	if !friends {
		return ShareTaskOutput{}, task.ErrNotContact
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return ShareTaskOutput{}, err
	}
	added := false
	t, err := change(ctx, h.taskStore, h.groups, in.ID, in.User, func(t *task.Task, _ string) error {
		if t.OwnerID != in.User {
			return task.ErrNotOwner
		}
		if added = !slices.Contains(t.SharedWith, in.With); added {
			t.SharedWith, t.UpdatedAt = append(slices.Clone(t.SharedWith), in.With), now
		}
		return nil
	})
	if err != nil || !added {
		return ShareTaskOutput{Task: t}, err
	}
	if err := h.journal.Involve(ctx, kindShared, t, in.User, in.With); err != nil {
		return ShareTaskOutput{}, err
	}
	return ShareTaskOutput{Task: t}, nil
}
