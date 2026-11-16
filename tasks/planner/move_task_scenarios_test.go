// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"reflect"
	"testing"

	"github.com/kitsunium/todo/tasks/task"
)

// givenMoveTaskCompletesAnOpenTask builds MoveTask for "completes an open task": the use case, its input, and
// what its output must be.
func givenMoveTaskCompletesAnOpenTask(t *testing.T) (MoveTask, MoveTaskInput, func(*testing.T, MoveTaskOutput)) {
	t.Helper()
	store, journal := newMemTasks(open("task_a")), &memJournal{}
	later := epoch.Add(3600e9)
	return NewMoveTask(store, fixedClock(later), journal), MoveTaskInput{Owner: ada, ID: "task_a", Event: "complete"}, func(t *testing.T, out MoveTaskOutput) {
		t.Helper()
		if out.Task.State != task.LifecycleDone || !out.Task.DoneAt.Equal(later) || !out.Task.UpdatedAt.Equal(later) {
			t.Errorf("the task completed: %+v", out.Task)
		}
		if kept, _ := store.Get(t.Context(), "task_a"); !reflect.DeepEqual(kept, out.Task) {
			t.Errorf("the store keeps %+v, want %+v", kept, out.Task)
		}
		if len(journal.kinds) != 1 || journal.kinds[0] != "complete" {
			t.Errorf("the journal was told %q, want complete once", journal.kinds)
		}
	}
}

// givenMoveTaskRefusesAnUnknownTask builds MoveTask for "refuses an unknown task": the use case and the input it
// refuses with task.ErrNotFound.
func givenMoveTaskRefusesAnUnknownTask(t *testing.T) (MoveTask, MoveTaskInput) {
	t.Helper()
	return NewMoveTask(newMemTasks(), fixedClock(epoch), &memJournal{}), MoveTaskInput{Owner: ada, ID: "task_nobody", Event: "complete"}
}

// givenMoveTaskRefusesToArchiveAnOpenTask builds MoveTask for "refuses to archive an open task": the use case and the input it
// refuses with task.ErrNotAllowed.
func givenMoveTaskRefusesToArchiveAnOpenTask(t *testing.T) (MoveTask, MoveTaskInput) {
	t.Helper()
	return NewMoveTask(newMemTasks(open("task_a")), fixedClock(epoch), &memJournal{}), MoveTaskInput{Owner: ada, ID: "task_a", Event: "archive"}
}

// givenMoveTaskRefusesTheTaskOfAnotherMember builds MoveTask for "refuses the task of another member": the use case and the input it
// refuses with task.ErrNotFound.
func givenMoveTaskRefusesTheTaskOfAnotherMember(t *testing.T) (MoveTask, MoveTaskInput) {
	t.Helper()
	return NewMoveTask(newMemTasks(open("task_a")), fixedClock(epoch), &memJournal{}), MoveTaskInput{Owner: bob, ID: "task_a", Event: "complete"}
}

// givenMoveTaskMovesATaskSharedWithTheMember builds MoveTask for "moves a task shared with the member": the use case, its input, and
// what its output must be.
func givenMoveTaskMovesATaskSharedWithTheMember(t *testing.T) (MoveTask, MoveTaskInput, func(*testing.T, MoveTaskOutput)) {
	t.Helper()
	uc := NewMoveTask(newMemTasks(shared("task_a")), fixedClock(epoch), &memJournal{})
	return uc, MoveTaskInput{Owner: bob, ID: "task_a", Event: "complete"}, func(t *testing.T, out MoveTaskOutput) {
		t.Helper()
		if out.Task.State != task.LifecycleDone || out.Task.OwnerID != ada {
			t.Errorf("the task Bob completed: %+v", out.Task)
		}
	}
}
