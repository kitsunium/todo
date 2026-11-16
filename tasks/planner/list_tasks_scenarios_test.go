// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"slices"
	"testing"

	"github.com/kitsunium/todo/tasks/task"
)

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
	uc := NewListTasks(newMemTasks(low, soon, later, high, gone, other))
	return uc, ListTasksInput{Owner: ada}, func(t *testing.T, out ListTasksOutput) {
		t.Helper()
		var got []string
		for _, x := range out.Tasks {
			got = append(got, x.ID)
		}
		if want := []string{"task_high", "task_soon", "task_later", "task_low"}; !slices.Equal(got, want) {
			t.Errorf("the list: %q, want %q", got, want)
		}
		archived, err := uc.Execute(t.Context(), ListTasksInput{Owner: ada, State: "archived"})
		if err != nil || len(archived.Tasks) != 1 || archived.Tasks[0].ID != "task_gone" {
			t.Errorf("the archived tasks: %+v (%v), want task_gone alone", archived.Tasks, err)
		}
	}
}
