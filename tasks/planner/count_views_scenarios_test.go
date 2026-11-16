// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"maps"
	"testing"
	"time"
)

// givenCountViewsCountsEveryView builds CountViews for "counts every view": the use case, its input, and
// what its output must be.
func givenCountViewsCountsEveryView(t *testing.T) (CountViews, CountViewsInput, func(*testing.T, CountViewsOutput)) {
	t.Helper()
	inbox, today, upcoming, late := open("task_inbox"), open("task_today"), open("task_upcoming"), open("task_late")
	today.Due, upcoming.Due, late.Due = epoch.Add(2*time.Hour), epoch.Add(48*time.Hour), epoch.Add(-time.Hour)
	sharedWithAda, assigned, onGroup := shared("task_shared"), open("task_assigned"), grouped("task_group")
	sharedWithAda.OwnerID, sharedWithAda.SharedWith = bob, []string{ada}
	assigned.OwnerID, assigned.AssigneeID = bob, ada
	thisWeek, lastWeek := done("task_this_week", epoch), done("task_last_week", epoch.Add(-24*time.Hour)) // epoch is a Monday
	unseen := open("task_unseen")
	unseen.OwnerID = dee
	uc := NewCountViews(crew(), fixedClock(epoch), newMemTasks(inbox, today, upcoming, late, sharedWithAda, assigned, onGroup, thisWeek, lastWeek, unseen))
	return uc, CountViewsInput{User: ada}, func(t *testing.T, out CountViewsOutput) {
		t.Helper()
		got := map[string]int{
			"inbox": out.Inbox, "today": out.Today, "upcoming": out.Upcoming, "shared": out.Shared, "assigned": out.Assigned,
			"overdue": out.Overdue, "completedThisWeek": out.CompletedThisWeek,
		}
		want := map[string]int{"inbox": 4, "today": 2, "upcoming": 1, "shared": 1, "assigned": 1, "overdue": 1, "completedThisWeek": 1}
		if !maps.Equal(got, want) || !maps.Equal(out.Groups, map[string]int{core: 1}) {
			t.Errorf("the counts: %v and the groups %v, want %v and {core: 1}", got, out.Groups, want)
		}
		empty, err := uc.Execute(t.Context(), CountViewsInput{User: dee})
		if err != nil || empty.Inbox != 1 || empty.Shared != 0 || len(empty.Groups) != 0 {
			t.Errorf("Dee's counts: %+v (%v), want her one task in the inbox and no group", empty, err)
		}
	}
}

// givenCountViewsRefusesAnUnknownTimeZone builds CountViews for "refuses an unknown time zone": the use case and the input it
// refuses with task.ErrBadZone.
func givenCountViewsRefusesAnUnknownTimeZone(t *testing.T) (CountViews, CountViewsInput) {
	t.Helper()
	return NewCountViews(crew(), fixedClock(epoch), newMemTasks()), CountViewsInput{User: ada, TZ: "Local"}
}
