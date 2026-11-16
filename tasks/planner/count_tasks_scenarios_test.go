// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"maps"
	"testing"
	"time"

	"github.com/kitsunium/todo/tasks/task"
)

// givenCountTasksCountsEveryState builds CountTasks for "counts every state": the use case, its input, and
// what its output must be.
func givenCountTasksCountsEveryState(t *testing.T) (CountTasks, CountTasksInput, func(*testing.T, CountTasksOutput)) {
	t.Helper()
	late, onTime := open("task_late"), open("task_on_time")
	late.Due, onTime.Due = epoch.Add(-time.Hour), epoch.Add(time.Hour)
	lateDone := done("task_done", epoch)
	lateDone.Due = epoch.Add(-time.Hour)
	uc := NewCountTasks(fixedClock(epoch), newMemTasks(late, onTime, lateDone))
	return uc, CountTasksInput{}, func(t *testing.T, out CountTasksOutput) {
		t.Helper()
		want := map[string]int{string(task.LifecycleOpen): 2, string(task.LifecycleDone): 1, string(task.LifecycleArchived): 0}
		if !maps.Equal(out.Counts, want) || out.Total != 3 || out.Overdue != 1 {
			t.Errorf("the count: %+v, want %v, 3 in all, 1 overdue", out, want)
		}
	}
}
