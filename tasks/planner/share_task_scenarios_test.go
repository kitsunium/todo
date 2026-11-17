// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"slices"
	"testing"
)

// givenShareTaskSharesWithAContact builds ShareTask for "shares with a contact": the use case, its input, and
// what its output must be.
func givenShareTaskSharesWithAContact(t *testing.T) (ShareTask, ShareTaskInput, func(*testing.T, ShareTaskOutput)) {
	t.Helper()
	store, journal := newMemTasks(open("task_a")), &memJournal{}
	uc := NewShareTask(store, crew(), contacts{ada + "|" + bob: true}, fixedClock(epoch), journal)
	in := ShareTaskInput{User: ada, ID: "task_a", With: bob}
	return uc, in, func(t *testing.T, out ShareTaskOutput) {
		t.Helper()
		if !slices.Equal(out.Task.SharedWith, []string{bob}) || !out.Task.SeenBy(bob, false) {
			t.Errorf("the task shared: %+v", out.Task)
		}
		again, err := uc.Execute(t.Context(), in)
		if want := "shared by " + ada + " to " + bob; len(journal.acts) != 1 || journal.acts[0] != want {
			t.Errorf("who did it: %q, want %q once", journal.acts, want)
		}
		if err != nil || !slices.Equal(again.Task.SharedWith, []string{bob}) || len(journal.kinds) != 1 || journal.kinds[0] != "shared" {
			t.Errorf("shared twice: %+v (%v), the journal told %q: want Bob once", again.Task.SharedWith, err, journal.kinds)
		}
	}
}

// givenShareTaskRefusesWhoIsNoContact builds ShareTask for "refuses who is no contact": the use case and the input it
// refuses with task.ErrNotContact.
func givenShareTaskRefusesWhoIsNoContact(t *testing.T) (ShareTask, ShareTaskInput) {
	t.Helper()
	uc := NewShareTask(newMemTasks(open("task_a")), crew(), contacts{ada + "|" + bob: true}, fixedClock(epoch), &memJournal{})
	return uc, ShareTaskInput{User: ada, ID: "task_a", With: cy}
}

// givenShareTaskRefusesWhoIsNotTheOwner builds ShareTask for "refuses who is not the owner": the use case and the input it
// refuses with task.ErrNotOwner.
func givenShareTaskRefusesWhoIsNotTheOwner(t *testing.T) (ShareTask, ShareTaskInput) {
	t.Helper()
	uc := NewShareTask(newMemTasks(shared("task_a")), crew(), contacts{bob + "|" + cy: true}, fixedClock(epoch), &memJournal{})
	return uc, ShareTaskInput{User: bob, ID: "task_a", With: cy}
}

// givenShareTaskRefusesAMemberOfTheGroupWhoIsNotTheOwner builds ShareTask for "refuses a member of the group who is not the owner": the use case and the input it
// refuses with task.ErrNotOwner.
func givenShareTaskRefusesAMemberOfTheGroupWhoIsNotTheOwner(t *testing.T) (ShareTask, ShareTaskInput) {
	t.Helper()
	uc := NewShareTask(newMemTasks(grouped("task_a")), crew(), contacts{bob + "|" + dee: true}, fixedClock(epoch), &memJournal{})
	return uc, ShareTaskInput{User: bob, ID: "task_a", With: dee} // an admin of the group runs its list, and shares none of its tasks
}
