// What ListTasks and CountViews share: the member's day and week in their
// time zone, the views of the list, and the order of the completed one.

package planner

import (
	"cmp"
	"slices"
	"time"
	_ "time/tzdata" // a view reads the member's time zone by name, in any image

	"github.com/kitsunium/todo/tasks/task"
)

// The views of the list; the empty view is every task not archived, or those
// of a state.
const (
	viewInbox     = "inbox"
	viewToday     = "today"
	viewUpcoming  = "upcoming"
	viewShared    = "shared"
	viewAssigned  = "assigned"
	viewCompleted = "completed"
	viewGroup     = "group"
)

// dayAndWeek places now in the member's time zone, an IANA name, UTC when
// none is given, and returns the end of their day and the start of their
// week, a Monday at midnight.
func dayAndWeek(now time.Time, tz string) (time.Time, time.Time, error) {
	loc := time.UTC
	if tz != "" {
		l, err := time.LoadLocation(tz)
		if err != nil || tz == "Local" {
			return time.Time{}, time.Time{}, task.ErrBadZone
		}
		loc = l
	}
	local := now.In(loc)
	midnight := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	sinceMonday := (int(local.Weekday()) + 6) % 7
	return midnight.AddDate(0, 0, 1), midnight.AddDate(0, 0, -sinceMonday), nil
}

// inView reports whether a task belongs to a view for a member, the end of
// whose day is tomorrow, and who is or is not in the task's group. A view
// lists what is left to do, but two: the group view keeps the done tasks until
// they are archived, and completed lists the finished ones, done or archived.
// The inbox is the member's own tasks on no list of a group they are in: the
// tasks of a group deleted, which a message may not have reached yet, are
// back in it.
func inView(view string, t task.Task, user string, inGroup bool, tomorrow time.Time) bool {
	active := t.State == task.LifecycleOpen
	switch view {
	case viewInbox:
		return t.OwnerID == user && !inGroup && active
	case viewToday:
		return active && !t.Due.IsZero() && t.Due.Before(tomorrow)
	case viewUpcoming:
		return active && !t.Due.IsZero() && !t.Due.Before(tomorrow)
	case viewShared:
		return slices.Contains(t.SharedWith, user) && active
	case viewAssigned:
		return t.AssigneeID == user && active
	case viewCompleted:
		return t.State == task.LifecycleDone || t.State == task.LifecycleArchived
	}
	return t.State != task.LifecycleArchived
}

// byCompletion orders the completed view: the most recently finished first.
func byCompletion(a, b task.Task) int {
	finished := func(t task.Task) time.Time {
		if t.DoneAt.IsZero() {
			return t.UpdatedAt
		}
		return t.DoneAt
	}
	return cmp.Or(finished(b).Compare(finished(a)), cmp.Compare(a.ID, b.ID))
}
