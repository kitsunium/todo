// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"
	"slices"

	"github.com/kitsunium/todo/tasks/task"
)

// execute is ListTasks's body: the tasks the member sees, of the view asked
// for, or of the state asked for, or every one that is not archived; a group
// the member is in narrows the list. The most urgent first, the completed
// view the most recently finished first.
func (h *ListTasksHandler) execute(ctx context.Context, in *ListTasksInput) (ListTasksOutput, error) {
	switch in.View {
	case "", viewInbox, viewToday, viewUpcoming, viewShared, viewAssigned, viewCompleted:
	case viewGroup:
		if in.Group == "" {
			return ListTasksOutput{}, task.ErrGroupNeeded
		}
	default:
		return ListTasksOutput{}, task.ErrUnknownView
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return ListTasksOutput{}, err
	}
	tomorrow, _, err := dayAndWeek(now, in.TZ)
	if err != nil {
		return ListTasksOutput{}, err
	}
	groups, err := h.groups.Memberships(ctx, in.Owner)
	if err != nil {
		return ListTasksOutput{}, err
	}
	if in.Group != "" && groups[in.Group] == "" {
		return ListTasksOutput{}, task.ErrNoGroup
	}
	all, err := h.taskStore.List(ctx)
	if err != nil {
		return ListTasksOutput{}, err
	}
	out := make([]task.Task, 0, len(all))
	for _, t := range all {
		member := groups[t.GroupID] != ""
		if !t.SeenBy(in.Owner, member) || in.Group != "" && t.GroupID != in.Group {
			continue
		}
		if in.View != "" && inView(in.View, t, in.Owner, member, tomorrow) ||
			in.View == "" && (in.State == "" && t.State != task.LifecycleArchived || in.State != "" && string(t.State) == in.State) {
			out = append(out, t)
		}
	}
	if in.View == viewCompleted {
		slices.SortFunc(out, byCompletion)
	} else {
		slices.SortFunc(out, byUrgency)
	}
	return ListTasksOutput{Tasks: out}, nil
}
