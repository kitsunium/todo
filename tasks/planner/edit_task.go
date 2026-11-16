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
// as when a task is added; a field left out stays, an empty assignee clears
// it. A task the member does not see is not found.
func (h *EditTaskHandler) execute(ctx context.Context, in *EditTaskInput) (EditTaskOutput, error) {
	var title string
	if in.Title != nil {
		title = strings.TrimSpace(*in.Title)
		switch {
		case title == "":
			return EditTaskOutput{}, task.ErrEmptyTitle
		case utf8.RuneCountInString(title) > task.MaxTitle:
			return EditTaskOutput{}, task.ErrTitleTooLong
		}
	}
	var priority task.Priority
	if in.Priority != nil {
		var err error
		if priority, err = task.ParsePriority(*in.Priority); err != nil {
			return EditTaskOutput{}, err
		}
	}
	// Whether the new assignee is in the task's group is asked before the
	// change, of the group the task is on now; the change holds it to that.
	var next, group string
	inGroup := false
	if in.AssigneeID != nil {
		if next = strings.TrimSpace(*in.AssigneeID); next != "" {
			seen, _, err := load(ctx, h.taskStore, h.groups, in.ID, in.User)
			if err != nil {
				return EditTaskOutput{}, err
			}
			if group = seen.GroupID; group != "" {
				role, err := h.groups.RoleOf(ctx, group, next)
				if err != nil {
					return EditTaskOutput{}, err
				}
				inGroup = role != ""
			}
		}
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return EditTaskOutput{}, err
	}
	t, err := change(ctx, h.taskStore, h.groups, in.ID, in.User, func(t *task.Task, _ string) error {
		if in.Title != nil {
			t.Title = title
		}
		if in.Priority != nil {
			t.Priority = priority
		}
		if in.Notes != nil {
			t.Notes = strings.TrimSpace(*in.Notes)
		}
		if in.AssigneeID != nil && next != t.AssigneeID {
			if next != "" && !assignable(*t, next, inGroup && t.GroupID == group) {
				return task.ErrNotAssignable
			}
			t.AssigneeID = next
		}
		if in.Due != nil {
			t.Due = *in.Due
		}
		t.UpdatedAt = now
		return nil
	})
	if err != nil {
		return EditTaskOutput{}, err
	}
	if err := h.journal.Record(ctx, kindEdited, t); err != nil {
		return EditTaskOutput{}, err
	}
	return EditTaskOutput{Task: t}, nil
}
