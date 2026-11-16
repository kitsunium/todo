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

// execute is AddTask's body: the title and the priority are checked, then the
// group, which the owner must be in, and the assignee, who must be someone the
// task may be assigned to; the task is kept open on its owner's list, or the
// group's, and the journal told.
func (h *AddTaskHandler) execute(ctx context.Context, in *AddTaskInput) (AddTaskOutput, error) {
	title := strings.TrimSpace(in.Title)
	switch {
	case title == "":
		return AddTaskOutput{}, task.ErrEmptyTitle
	case utf8.RuneCountInString(title) > task.MaxTitle:
		return AddTaskOutput{}, task.ErrTitleTooLong
	}
	priority, err := task.ParsePriority(in.Priority)
	if err != nil {
		return AddTaskOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return AddTaskOutput{}, err
	}
	t := task.Task{
		OwnerID: in.Owner, GroupID: strings.TrimSpace(in.GroupID), AssigneeID: strings.TrimSpace(in.AssigneeID),
	}
	role, err := roleIn(ctx, h.groups, t, in.Owner)
	if err != nil {
		return AddTaskOutput{}, err
	}
	if t.GroupID != "" && role == "" {
		return AddTaskOutput{}, task.ErrNotMember
	}
	if t.AssigneeID != "" {
		inGroup := false
		if t.GroupID != "" {
			assignee, err := h.groups.RoleOf(ctx, t.GroupID, t.AssigneeID)
			if err != nil {
				return AddTaskOutput{}, err
			}
			inGroup = assignee != ""
		}
		if !assignable(t, t.AssigneeID, inGroup) {
			return AddTaskOutput{}, task.ErrNotAssignable
		}
	}
	id, err := h.taskStore.NewID(ctx)
	if err != nil {
		return AddTaskOutput{}, err
	}
	t.ID, t.Title, t.Notes, t.Priority, t.Due = id, title, strings.TrimSpace(in.Notes), priority, in.Due
	t.State, t.CreatedAt, t.UpdatedAt = task.LifecycleInitial, now, now
	if err := h.taskStore.Save(ctx, t); err != nil {
		return AddTaskOutput{}, err
	}
	if err := h.journal.Record(ctx, kindAdded, t); err != nil {
		return AddTaskOutput{}, err
	}
	if t.AssigneeID != "" && t.AssigneeID != in.Owner {
		if err := h.journal.Tell(ctx, kindAssigned, t, t.AssigneeID); err != nil {
			return AddTaskOutput{}, err
		}
	}
	return AddTaskOutput{Task: t}, nil
}

// kindAdded is what the journal is told of a task just added; a move is
// told by its lifecycle event.
const kindAdded = "added"

// kindAssigned is what a member is told of a task someone else assigned to
// them.
const kindAssigned = "assigned"
