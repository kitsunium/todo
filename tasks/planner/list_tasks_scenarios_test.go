// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"slices"
	"testing"
	"time"

	"github.com/kitsunium/todo/tasks/task"
)

// ids are the ids of tasks, in order.
func ids(ts []task.Task) []string {
	var out []string
	for _, x := range ts {
		out = append(out, x.ID)
	}
	return out
}

// givenListTasksListsTheMostUrgentFirst builds ListTasks for "lists the most urgent first": the use case, its input, and
// what its output must be.
func givenListTasksListsTheMostUrgentFirst(t *testing.T) (ListTasks, ListTasksInput, func(*testing.T, ListTasksOutput)) {
	t.Helper()
	low, soon, later, high, gone := open("task_low"), open("task_soon"), open("task_later"), open("task_high"), done("task_gone", epoch)
	low.Priority, high.Priority = task.PriorityLow, task.PriorityHigh
	soon.Due, later.Due = epoch.Add(3600e9), epoch.Add(2*3600e9)
	gone.State = task.LifecycleArchived
	other := open("task_of_bob")
	other.OwnerID = bob
	uc := NewListTasks(crew(), fixedClock(epoch), newMemTasks(low, soon, later, high, gone, other))
	return uc, ListTasksInput{Owner: ada}, func(t *testing.T, out ListTasksOutput) {
		t.Helper()
		if got, want := ids(out.Tasks), []string{"task_high", "task_soon", "task_later", "task_low"}; !slices.Equal(got, want) {
			t.Errorf("the list: %q, want %q", got, want)
		}
		archived, err := uc.Execute(t.Context(), ListTasksInput{Owner: ada, State: "archived"})
		if err != nil || len(archived.Tasks) != 1 || archived.Tasks[0].ID != "task_gone" {
			t.Errorf("the archived tasks: %+v (%v), want task_gone alone", archived.Tasks, err)
		}
	}
}

// givenListTasksListsTheTasksOfAGroup builds ListTasks for "lists the tasks of a group": the use case, its input, and
// what its output must be.
func givenListTasksListsTheTasksOfAGroup(t *testing.T) (ListTasks, ListTasksInput, func(*testing.T, ListTasksOutput)) {
	t.Helper()
	onGroup, mine, assigned, sharedOne, stranger := grouped("task_group"), open("task_mine"), open("task_assigned"), shared("task_shared"), grouped("task_unseen")
	mine.OwnerID, assigned.AssigneeID, sharedOne.SharedWith = cy, cy, []string{cy}
	stranger.OwnerID, stranger.GroupID = dee, "group_other" // a group Cy is not in
	uc := NewListTasks(crew(), fixedClock(epoch), newMemTasks(onGroup, mine, assigned, sharedOne, stranger))
	return uc, ListTasksInput{Owner: cy}, func(t *testing.T, out ListTasksOutput) {
		t.Helper()
		got := ids(out.Tasks)
		slices.Sort(got)
		if want := []string{"task_assigned", "task_group", "task_mine", "task_shared"}; !slices.Equal(got, want) {
			t.Errorf("what Cy sees: %q, want %q", got, want)
		}
		narrowed, err := uc.Execute(t.Context(), ListTasksInput{Owner: cy, Group: core})
		if err != nil || !slices.Equal(ids(narrowed.Tasks), []string{"task_group"}) {
			t.Errorf("the tasks of the group: %q (%v), want task_group alone", ids(narrowed.Tasks), err)
		}
		both, err := uc.Execute(t.Context(), ListTasksInput{Owner: cy, View: "group", Group: core})
		if err != nil || !slices.Equal(ids(both.Tasks), []string{"task_group"}) {
			t.Errorf("the group view: %q (%v), want task_group alone", ids(both.Tasks), err)
		}
	}
}

// givenListTasksListsAView builds ListTasks for "lists a view": the use case, its input, and
// what its output must be.
func givenListTasksListsAView(t *testing.T) (ListTasks, ListTasksInput, func(*testing.T, ListTasksOutput)) {
	t.Helper()
	// The day of a member in Paris ends at 2026-11-16 23:00 UTC: Paris is an hour ahead.
	inbox, today, tonight, upcoming := open("task_inbox"), open("task_today"), open("task_tonight"), open("task_upcoming")
	late, finished, assigned, onGroup := open("task_late"), done("task_finished", epoch), open("task_assigned"), grouped("task_group")
	today.Due, tonight.Due = epoch.Add(2*time.Hour), time.Date(2026, 11, 16, 22, 30, 0, 0, time.UTC)
	upcoming.Due = time.Date(2026, 11, 16, 23, 30, 0, 0, time.UTC) // after midnight in Paris
	late.Due = epoch.Add(-48 * time.Hour)
	assigned.OwnerID, assigned.AssigneeID = bob, ada
	older := done("task_older", epoch.Add(-72*time.Hour))
	uc := NewListTasks(crew(), fixedClock(epoch), newMemTasks(inbox, today, tonight, upcoming, late, finished, older, assigned, onGroup))
	return uc, ListTasksInput{Owner: ada, View: "today", TZ: "Europe/Paris"}, func(t *testing.T, out ListTasksOutput) {
		t.Helper()
		ask := func(view string) []string {
			t.Helper()
			res, err := uc.Execute(t.Context(), ListTasksInput{Owner: ada, View: view, TZ: "Europe/Paris"})
			if err != nil {
				t.Fatalf("view %s: %v", view, err)
			}
			got := ids(res.Tasks)
			slices.Sort(got)
			return got
		}
		if got, want := ids(out.Tasks), []string{"task_late", "task_today", "task_tonight"}; !slices.Equal(got, want) {
			t.Errorf("today in Paris: %q, want %q (the overdue one first)", got, want)
		}
		for view, want := range map[string][]string{
			"inbox":    {"task_inbox", "task_late", "task_today", "task_tonight", "task_upcoming"},
			"upcoming": {"task_upcoming"},
			"assigned": {"task_assigned"},
			"shared":   nil,
		} {
			if got := ask(view); !slices.Equal(got, want) {
				t.Errorf("view %s: %q, want %q", view, got, want)
			}
		}
		completed, err := uc.Execute(t.Context(), ListTasksInput{Owner: ada, View: "completed"})
		if err != nil || !slices.Equal(ids(completed.Tasks), []string{"task_finished", "task_older"}) {
			t.Errorf("completed, the latest first: %q (%v)", ids(completed.Tasks), err)
		}
	}
}

// givenListTasksRefusesAnUnknownView builds ListTasks for "refuses an unknown view": the use case and the input it
// refuses with task.ErrUnknownView.
func givenListTasksRefusesAnUnknownView(t *testing.T) (ListTasks, ListTasksInput) {
	t.Helper()
	return NewListTasks(crew(), fixedClock(epoch), newMemTasks()), ListTasksInput{Owner: ada, View: "someday"}
}

// givenListTasksRefusesTheGroupViewWithoutAGroup builds ListTasks for "refuses the group view without a group": the use case and the input it
// refuses with task.ErrGroupNeeded.
func givenListTasksRefusesTheGroupViewWithoutAGroup(t *testing.T) (ListTasks, ListTasksInput) {
	t.Helper()
	return NewListTasks(crew(), fixedClock(epoch), newMemTasks()), ListTasksInput{Owner: ada, View: "group"}
}

// givenListTasksRefusesAnUnknownTimeZone builds ListTasks for "refuses an unknown time zone": the use case and the input it
// refuses with task.ErrBadZone.
func givenListTasksRefusesAnUnknownTimeZone(t *testing.T) (ListTasks, ListTasksInput) {
	t.Helper()
	return NewListTasks(crew(), fixedClock(epoch), newMemTasks()), ListTasksInput{Owner: ada, View: "today", TZ: "Mars/Olympus"}
}

// givenListTasksRefusesAGroupTheMemberIsNotIn builds ListTasks for "refuses a group the member is not in": the use case and the input it
// refuses with task.ErrNoGroup.
func givenListTasksRefusesAGroupTheMemberIsNotIn(t *testing.T) (ListTasks, ListTasksInput) {
	t.Helper()
	return NewListTasks(crew(), fixedClock(epoch), newMemTasks(grouped("task_group"))), ListTasksInput{Owner: dee, Group: core}
}
