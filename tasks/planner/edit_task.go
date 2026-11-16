// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/kitsunium/todo/tasks/task"
)

// kindEdited is what the journal is told of a task whose words changed.
const kindEdited = "edited"

// execute is EditTask's body: each field given replaces the task's, checked
// as when a task is added; a field left out stays. A task the member does
// not see is not found.
func (h *EditTaskHandler) execute(ctx context.Context, in *EditTaskInput) (EditTaskOutput, error) {
	t, err := h.taskStore.Get(ctx, in.ID)
	if err != nil {
		return EditTaskOutput{}, err
	}
	if !t.SeenBy(in.User) {
		return EditTaskOutput{}, task.ErrNotFound
	}
	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		switch {
		case title == "":
			return EditTaskOutput{}, task.ErrEmptyTitle
		case utf8.RuneCountInString(title) > task.MaxTitle:
			return EditTaskOutput{}, task.ErrTitleTooLong
		}
		t.Title = title
	}
	if in.Priority != nil {
		if t.Priority, err = task.ParsePriority(*in.Priority); err != nil {
			return EditTaskOutput{}, err
		}
	}
	if in.Notes != nil {
		t.Notes = strings.TrimSpace(*in.Notes)
	}
	if in.Assignee != nil {
		t.Assignee = strings.TrimSpace(*in.Assignee)
	}
	if in.Due != nil {
		t.Due = *in.Due
	}
	if t.UpdatedAt, err = h.clock.Now(ctx); err != nil {
		return EditTaskOutput{}, err
	}
	if err := h.taskStore.Save(ctx, t); err != nil {
		return EditTaskOutput{}, err
	}
	if err := h.journal.Record(ctx, kindEdited, t); err != nil {
		return EditTaskOutput{}, err
	}
	return EditTaskOutput{Task: t}, nil
}
