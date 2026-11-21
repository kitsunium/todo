// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"testing"
)

// givenUnshareTaskLetsAMemberLeaveATask builds UnshareTask for "lets a member leave a task": the use case, its input, and
// what its output must be.
func givenUnshareTaskLetsAMemberLeaveATask(t *testing.T) (UnshareTask, UnshareTaskInput, func(*testing.T, UnshareTaskOutput)) {
	t.Helper()
	store := newMemTasks(shared("task_a"))
	return NewUnshareTask(store, crew(), fixedClock(epoch)), UnshareTaskInput{User: bob, ID: "task_a", With: bob}, func(t *testing.T, out UnshareTaskOutput) {
		t.Helper()
		kept := store.byID["task_a"]
		if len(out.Task.SharedWith) != 0 || out.Task.SeenBy(bob, false) || kept.SeenBy(bob, false) {
			t.Errorf("the task Bob left: %+v", out.Task)
		}
	}
}

// givenUnshareTaskRefusesAMemberWhoRemovesAnother builds UnshareTask for "refuses a member who removes another": the use case and the input it
// refuses with task.ErrNotOwner.
func givenUnshareTaskRefusesAMemberWhoRemovesAnother(t *testing.T) (UnshareTask, UnshareTaskInput) {
	t.Helper()
	both := shared("task_a")
	both.SharedWith = append(both.SharedWith, cy)
	return NewUnshareTask(newMemTasks(both), crew(), fixedClock(epoch)), UnshareTaskInput{User: bob, ID: "task_a", With: cy}
}
