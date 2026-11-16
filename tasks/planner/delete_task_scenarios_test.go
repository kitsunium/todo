// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"testing"
)

// givenDeleteTaskDeletesATaskOfTheOwner builds DeleteTask for "deletes a task of the owner": the use case, its input, and
// what its output must be.
func givenDeleteTaskDeletesATaskOfTheOwner(t *testing.T) (DeleteTask, DeleteTaskInput, func(*testing.T, DeleteTaskOutput)) {
	t.Helper()
	store, journal := newMemTasks(shared("task_a")), &memJournal{}
	return NewDeleteTask(store, crew(), journal), DeleteTaskInput{User: ada, ID: "task_a"}, func(t *testing.T, _ DeleteTaskOutput) {
		t.Helper()
		if len(store.byID) != 0 || len(journal.kinds) != 1 || journal.kinds[0] != "deleted" {
			t.Errorf("%d task(s) left, the journal told %q: want none, and deleted once", len(store.byID), journal.kinds)
		}
	}
}

// givenDeleteTaskDeletesATaskOfTheGroupAsItsAdmin builds DeleteTask for "deletes a task of the group as its admin": the use case, its input, and
// what its output must be.
func givenDeleteTaskDeletesATaskOfTheGroupAsItsAdmin(t *testing.T) (DeleteTask, DeleteTaskInput, func(*testing.T, DeleteTaskOutput)) {
	t.Helper()
	mine := grouped("task_a")
	mine.OwnerID = cy // a member's task, deleted by the admin
	store := newMemTasks(mine)
	return NewDeleteTask(store, crew(), &memJournal{}), DeleteTaskInput{User: bob, ID: "task_a"}, func(t *testing.T, _ DeleteTaskOutput) {
		t.Helper()
		if len(store.byID) != 0 {
			t.Errorf("%d task(s) left, want none", len(store.byID))
		}
	}
}

// givenDeleteTaskRefusesAMemberItIsSharedWith builds DeleteTask for "refuses a member it is shared with": the use case and the input it
// refuses with task.ErrNotOwner.
func givenDeleteTaskRefusesAMemberItIsSharedWith(t *testing.T) (DeleteTask, DeleteTaskInput) {
	t.Helper()
	return NewDeleteTask(newMemTasks(shared("task_a")), crew(), &memJournal{}), DeleteTaskInput{User: bob, ID: "task_a"}
}

// givenDeleteTaskRefusesAPlainMemberOfTheGroup builds DeleteTask for "refuses a plain member of the group": the use case and the input it
// refuses with task.ErrNotOwner.
func givenDeleteTaskRefusesAPlainMemberOfTheGroup(t *testing.T) (DeleteTask, DeleteTaskInput) {
	t.Helper()
	return NewDeleteTask(newMemTasks(grouped("task_a")), crew(), &memJournal{}), DeleteTaskInput{User: cy, ID: "task_a"}
}
