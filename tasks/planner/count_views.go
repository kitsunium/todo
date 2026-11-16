// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"

	"github.com/kitsunium/todo/tasks/task"
)

// execute is CountViews's body: for the tasks the member sees, how many are
// left to do in each view, how many are overdue, how many were finished since
// Monday, and how many are open on each of their groups' lists, every group
// named even at zero.
func (h *CountViewsHandler) execute(ctx context.Context, in *CountViewsInput) (CountViewsOutput, error) {
	now, err := h.clock.Now(ctx)
	if err != nil {
		return CountViewsOutput{}, err
	}
	tomorrow, weekStart, err := dayAndWeek(now, in.TZ)
	if err != nil {
		return CountViewsOutput{}, err
	}
	groups, err := h.groups.Memberships(ctx, in.User)
	if err != nil {
		return CountViewsOutput{}, err
	}
	all, err := h.taskStore.List(ctx)
	if err != nil {
		return CountViewsOutput{}, err
	}
	out := CountViewsOutput{Groups: map[string]int{}}
	for id := range groups {
		out.Groups[id] = 0
	}
	counted := []struct {
		view string
		n    *int
	}{
		{viewInbox, &out.Inbox}, {viewToday, &out.Today}, {viewUpcoming, &out.Upcoming},
		{viewShared, &out.Shared}, {viewAssigned, &out.Assigned},
	}
	for _, t := range all {
		_, member := groups[t.GroupID]
		if !t.SeenBy(in.User, member) {
			continue
		}
		for _, c := range counted {
			if inView(c.view, t, in.User, member, tomorrow) {
				*c.n++
			}
		}
		if t.Overdue(now) {
			out.Overdue++
		}
		if !t.DoneAt.IsZero() && !t.DoneAt.Before(weekStart) {
			out.CompletedThisWeek++
		}
		if member && t.State == task.LifecycleOpen {
			out.Groups[t.GroupID]++
		}
	}
	return out, nil
}
