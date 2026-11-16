// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"reflect"
	"strings"
	"testing"

	"github.com/kitsunium/todo/tasks/task"
)

// givenAddTaskAddsAnOpenTask builds AddTask for "adds an open task": the use case, its input, and
// what its output must be.
func givenAddTaskAddsAnOpenTask(t *testing.T) (AddTask, AddTaskInput, func(*testing.T, AddTaskOutput)) {
	t.Helper()
	store, journal := newMemTasks(), &memJournal{}
	due := epoch.Add(48 * 3600e9)
	in := AddTaskInput{Owner: ada, Title: "  write the design  ", Priority: "High", AssigneeID: ada, Due: due}
	return NewAddTask(fixedClock(epoch), crew(), contacts{}, store, journal), in, func(t *testing.T, out AddTaskOutput) {
		t.Helper()
		got := out.Task
		if got.ID == "" || got.OwnerID != ada || got.Title != "write the design" || got.Priority != task.PriorityHigh || got.State != task.LifecycleOpen {
			t.Errorf("the task added: %+v", got)
		}
		if got.AssigneeID != ada || got.GroupID != "" {
			t.Errorf("its assignee and its group: %+v", got)
		}
		if !got.CreatedAt.Equal(epoch) || !got.UpdatedAt.Equal(epoch) || !got.Due.Equal(due) || !got.DoneAt.IsZero() {
			t.Errorf("its dates: %+v", got)
		}
		if kept, err := store.Get(t.Context(), got.ID); err != nil || !reflect.DeepEqual(kept, got) {
			t.Errorf("the store keeps %+v (%v), want %+v", kept, err, got)
		}
		if len(journal.kinds) != 1 || journal.kinds[0] != "added" {
			t.Errorf("the journal was told %q, want added once", journal.kinds)
		}
	}
}

// givenAddTaskAddsATaskToAGroupList builds AddTask for "adds a task to a group list": the use case, its input, and
// what its output must be.
func givenAddTaskAddsATaskToAGroupList(t *testing.T) (AddTask, AddTaskInput, func(*testing.T, AddTaskOutput)) {
	t.Helper()
	store := newMemTasks()
	in := AddTaskInput{Owner: cy, Title: "plan the sprint", GroupID: core, AssigneeID: bob}
	return NewAddTask(fixedClock(epoch), crew(), contacts{}, store, &memJournal{}), in, func(t *testing.T, out AddTaskOutput) {
		t.Helper()
		got := out.Task
		if got.OwnerID != cy || got.GroupID != core || got.AssigneeID != bob {
			t.Errorf("the task added: %+v", got)
		}
		if kept, _ := store.Get(t.Context(), got.ID); kept.GroupID != core || !got.SeenBy(ada, true) || got.SeenBy(dee, false) {
			t.Errorf("the group task is kept %+v: Ada, in the group, sees it and Dee does not", kept)
		}
	}
}

// givenAddTaskAssignsATaskToAContact builds AddTask for "assigns a task to a contact": the use case, its input, and
// what its output must be.
func givenAddTaskAssignsATaskToAContact(t *testing.T) (AddTask, AddTaskInput, func(*testing.T, AddTaskOutput)) {
	t.Helper()
	in := AddTaskInput{Owner: dee, Title: "water the plants", AssigneeID: ada}
	return NewAddTask(fixedClock(epoch), crew(), contacts{dee + "|" + ada: true}, newMemTasks(), &memJournal{}), in, func(t *testing.T, out AddTaskOutput) {
		t.Helper()
		if out.Task.AssigneeID != ada || out.Task.OwnerID != dee || !out.Task.SeenBy(ada, false) {
			t.Errorf("the task assigned to a contact: %+v, want it seen by its assignee", out.Task)
		}
	}
}

// givenAddTaskRefusesAnEmptyTitle builds AddTask for "refuses an empty title": the use case and the input it
// refuses with task.ErrEmptyTitle.
func givenAddTaskRefusesAnEmptyTitle(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), crew(), contacts{}, newMemTasks(), &memJournal{}), AddTaskInput{Title: "   "}
}

// givenAddTaskRefusesALongTitle builds AddTask for "refuses a long title": the use case and the input it
// refuses with task.ErrTitleTooLong.
func givenAddTaskRefusesALongTitle(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), crew(), contacts{}, newMemTasks(), &memJournal{}), AddTaskInput{Title: strings.Repeat("é", task.MaxTitle+1)}
}

// givenAddTaskRefusesAnUnknownPriority builds AddTask for "refuses an unknown priority": the use case and the input it
// refuses with task.ErrUnknownPriority.
func givenAddTaskRefusesAnUnknownPriority(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), crew(), contacts{}, newMemTasks(), &memJournal{}), AddTaskInput{Title: "a task", Priority: "urgent"}
}

// givenAddTaskRefusesAGroupTheMemberIsNotIn builds AddTask for "refuses a group the member is not in": the use case and the input it
// refuses with task.ErrNotMember.
func givenAddTaskRefusesAGroupTheMemberIsNotIn(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), crew(), contacts{}, newMemTasks(), &memJournal{}), AddTaskInput{Owner: dee, Title: "a task", GroupID: core}
}

// givenAddTaskRefusesAnAssigneeWhoIsNoContact builds AddTask for "refuses an assignee who is no contact": the use case and the input it
// refuses with task.ErrNotAssignable.
func givenAddTaskRefusesAnAssigneeWhoIsNoContact(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), crew(), contacts{}, newMemTasks(), &memJournal{}), AddTaskInput{Owner: dee, Title: "a task", AssigneeID: ada}
}

// givenAddTaskRefusesAnAssigneeOutsideTheGroup builds AddTask for "refuses an assignee outside the group": the use case and the input it
// refuses with task.ErrNotAssignable.
func givenAddTaskRefusesAnAssigneeOutsideTheGroup(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	// Dee is a contact of Ada, which counts for a personal task and not for a group's.
	return NewAddTask(fixedClock(epoch), crew(), contacts{ada + "|" + dee: true}, newMemTasks(), &memJournal{}), AddTaskInput{Owner: ada, Title: "a task", GroupID: core, AssigneeID: dee}
}
