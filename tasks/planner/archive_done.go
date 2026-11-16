// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"

	"github.com/kitsunium/todo/tasks/task"
)

// execute is ArchiveDone's body: every task done for longer than After is
// archived, and the journal told of each.
func (h *ArchiveDoneHandler) execute(ctx context.Context, in *ArchiveDoneInput) (ArchiveDoneOutput, error) {
	now, err := h.clock.Now(ctx)
	if err != nil {
		return ArchiveDoneOutput{}, err
	}
	all, err := h.taskStore.List(ctx)
	if err != nil {
		return ArchiveDoneOutput{}, err
	}
	var out ArchiveDoneOutput
	for _, t := range all {
		if t.State != task.LifecycleDone || now.Sub(t.DoneAt) < in.After {
			continue
		}
		if err := move(ctx, &t, task.LifecycleOnArchive, now); err != nil {
			return out, err
		}
		if err := h.taskStore.Save(ctx, t); err != nil {
			return out, err
		}
		if err := h.journal.Record(ctx, string(task.LifecycleOnArchive), t); err != nil {
			return out, err
		}
		out.Archived++
	}
	return out, nil
}
