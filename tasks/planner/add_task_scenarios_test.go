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
	return NewAddTask(fixedClock(epoch), crew(), store, journal), in, func(t *testing.T, out AddTaskOutput) {
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
	return NewAddTask(fixedClock(epoch), crew(), store, &memJournal{}), in, func(t *testing.T, out AddTaskOutput) {
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

// givenAddTaskRefusesAnEmptyTitle builds AddTask for "refuses an empty title": the use case and the input it
// refuses with task.ErrEmptyTitle.
func givenAddTaskRefusesAnEmptyTitle(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), crew(), newMemTasks(), &memJournal{}), AddTaskInput{Title: "   "}
}

// givenAddTaskRefusesALongTitle builds AddTask for "refuses a long title": the use case and the input it
// refuses with task.ErrTitleTooLong.
func givenAddTaskRefusesALongTitle(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), crew(), newMemTasks(), &memJournal{}), AddTaskInput{Title: strings.Repeat("é", task.MaxTitle+1)}
}

// givenAddTaskRefusesAnUnknownPriority builds AddTask for "refuses an unknown priority": the use case and the input it
// refuses with task.ErrUnknownPriority.
func givenAddTaskRefusesAnUnknownPriority(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), crew(), newMemTasks(), &memJournal{}), AddTaskInput{Title: "a task", Priority: "urgent"}
}

// givenAddTaskRefusesAGroupTheMemberIsNotIn builds AddTask for "refuses a group the member is not in": the use case and the input it
// refuses with task.ErrNotMember.
func givenAddTaskRefusesAGroupTheMemberIsNotIn(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), crew(), newMemTasks(), &memJournal{}), AddTaskInput{Owner: dee, Title: "a task", GroupID: core}
}

// givenAddTaskRefusesAnAssigneeWhoDoesNotSeeTheTask builds AddTask for "refuses an assignee who does not see the task": the use case and the input it
// refuses with task.ErrNotAssignable.
func givenAddTaskRefusesAnAssigneeWhoDoesNotSeeTheTask(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	// Ada is a contact of Dee's in nobody's books here, and would not be assignable if she were: a new personal task is seen by its owner alone.
	return NewAddTask(fixedClock(epoch), crew(), newMemTasks(), &memJournal{}), AddTaskInput{Owner: dee, Title: "a task", AssigneeID: ada}
}

// givenAddTaskRefusesAnAssigneeOutsideTheGroup builds AddTask for "refuses an assignee outside the group": the use case and the input it
// refuses with task.ErrNotAssignable.
func givenAddTaskRefusesAnAssigneeOutsideTheGroup(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	// Dee is a contact of Ada, which counts for a personal task and not for a group's.
	return NewAddTask(fixedClock(epoch), crew(), newMemTasks(), &memJournal{}), AddTaskInput{Owner: ada, Title: "a task", GroupID: core, AssigneeID: dee}
}

// givenAddTaskTellsItsAssigneeOfATaskAssigned builds AddTask for "tells its assignee of a task assigned": the use case, its input, and
// what its output must be.
func givenAddTaskTellsItsAssigneeOfATaskAssigned(t *testing.T) (AddTask, AddTaskInput, func(*testing.T, AddTaskOutput)) {
	t.Helper()
	journal := &memJournal{}
	uc := NewAddTask(fixedClock(epoch), crew(), newMemTasks(), journal)
	return uc, AddTaskInput{Owner: cy, Title: "plan the sprint", GroupID: core, AssigneeID: bob}, func(t *testing.T, out AddTaskOutput) {
		t.Helper()
		if len(journal.told) != 1 || journal.told[0] != "assigned to "+bob {
			t.Errorf("who was told: %q, want Bob alone", journal.told)
		}
		mine, err := uc.Execute(t.Context(), AddTaskInput{Owner: cy, Title: "mine to do", GroupID: core, AssigneeID: cy})
		if err != nil || mine.Task.AssigneeID != cy || len(journal.told) != 1 {
			t.Errorf("a task one assigns to oneself: %+v (%v), told %q: want nobody told", mine.Task, err, journal.told)
		}
	}
}
