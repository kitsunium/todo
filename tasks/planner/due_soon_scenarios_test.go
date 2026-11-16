// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"slices"
	"testing"
	"time"
)

// givenDueSoonListsTheOpenTasksDueSoon builds DueSoon for "lists the open tasks due soon": the use case, its input, and
// what its output must be.
func givenDueSoonListsTheOpenTasksDueSoon(t *testing.T) (DueSoon, DueSoonInput, func(*testing.T, DueSoonOutput)) {
	t.Helper()
	tomorrow, tonight, nextWeek, late, none := open("task_tomorrow"), open("task_tonight"), open("task_next_week"), open("task_late"), open("task_none")
	tomorrow.Due, tonight.Due = epoch.Add(20*time.Hour), epoch.Add(8*time.Hour)
	nextWeek.Due, late.Due = epoch.Add(7*24*time.Hour), epoch.Add(-time.Hour)
	finished := done("task_done", epoch)
	finished.Due = epoch.Add(time.Hour)
	uc := NewDueSoon(fixedClock(epoch), newMemTasks(tomorrow, tonight, nextWeek, late, none, finished))
	return uc, DueSoonInput{Within: 24 * time.Hour}, func(t *testing.T, out DueSoonOutput) {
		t.Helper()
		var got []string
		for _, x := range out.Tasks {
			got = append(got, x.ID)
		}
		if want := []string{"task_tonight", "task_tomorrow"}; !slices.Equal(got, want) {
			t.Errorf("due soon: %q, want %q", got, want)
		}
	}
}
