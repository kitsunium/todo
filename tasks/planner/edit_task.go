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
// it. Its owner alone moves it to another group they are in, or to none, and
// an assignee who is not in the new group is cleared. A member newly
// assigned by someone else is told. A task the member does not see is not
// found.
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
	// What needs the groups is asked before the change, of the task as it is
	// now: the group it will be on, whether its owner may put it there, and
	// whether its assignee, the new one or the one it keeps, is in that group.
	// The change holds each answer to the task it then sees.
	seen, _, err := load(ctx, h.taskStore, h.groups, in.ID, in.User)
	if err != nil {
		return EditTaskOutput{}, err
	}
	group, moving := seen.GroupID, false
	if in.GroupID != nil {
		if target := strings.TrimSpace(*in.GroupID); target != seen.GroupID {
			if seen.OwnerID != in.User {
				return EditTaskOutput{}, task.ErrNotOwner
			}
			if target != "" {
				role, err := h.groups.RoleOf(ctx, target, in.User)
				if err != nil {
					return EditTaskOutput{}, err
				}
				if role == "" {
					return EditTaskOutput{}, task.ErrNotMember
				}
			}
			group, moving = target, true
		}
	}
	assignee, assigning := seen.AssigneeID, false
	if in.AssigneeID != nil {
		assignee = strings.TrimSpace(*in.AssigneeID)
		assigning = assignee != seen.AssigneeID
	}
	inGroup := false
	if assignee != "" && group != "" {
		role, err := h.groups.RoleOf(ctx, group, assignee)
		if err != nil {
			return EditTaskOutput{}, err
		}
		inGroup = role != ""
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return EditTaskOutput{}, err
	}
	before := seen.AssigneeID
	t, err := change(ctx, h.taskStore, h.groups, in.ID, in.User, func(t *task.Task, _ string) error {
		if moving {
			if t.OwnerID != in.User {
				return task.ErrNotOwner
			}
			t.GroupID = group
		}
		if in.Title != nil {
			t.Title = title
		}
		if in.Priority != nil {
			t.Priority = priority
		}
		if in.Notes != nil {
			t.Notes = strings.TrimSpace(*in.Notes)
		}
		before = t.AssigneeID
		if in.AssigneeID != nil {
			t.AssigneeID = assignee
		}
		// The assignee is judged again only when it changes or the task
		// moves: an edit of the title leaves who does the task alone.
		if (assigning || moving) && t.AssigneeID != "" && !assignable(*t, t.AssigneeID, inGroup && t.GroupID == group) {
			if assigning {
				return task.ErrNotAssignable
			}
			t.AssigneeID = "" // the task left the group its assignee did it for
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
	if t.AssigneeID != "" && t.AssigneeID != before && t.AssigneeID != in.User {
		if err := h.journal.Tell(ctx, kindAssigned, t, in.User, t.AssigneeID); err != nil {
			return EditTaskOutput{}, err
		}
	}
	return EditTaskOutput{Task: t}, nil
}
