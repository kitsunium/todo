// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
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
	in := AddTaskInput{Owner: ada, Title: "  write the design  ", Priority: "High", Assignee: "ada@example.com", Due: due}
	return NewAddTask(fixedClock(epoch), store, journal), in, func(t *testing.T, out AddTaskOutput) {
		t.Helper()
		got := out.Task
		if got.ID == "" || got.OwnerID != ada || got.Title != "write the design" || got.Priority != task.PriorityHigh || got.State != task.LifecycleOpen {
			t.Errorf("the task added: %+v", got)
		}
		if !got.CreatedAt.Equal(epoch) || !got.UpdatedAt.Equal(epoch) || !got.Due.Equal(due) || !got.DoneAt.IsZero() {
			t.Errorf("its dates: %+v", got)
		}
		if kept, err := store.Get(t.Context(), got.ID); err != nil || kept != got {
			t.Errorf("the store keeps %+v (%v), want %+v", kept, err, got)
		}
		if len(journal.kinds) != 1 || journal.kinds[0] != "added" {
			t.Errorf("the journal was told %q, want added once", journal.kinds)
		}
	}
}

// givenAddTaskRefusesAnEmptyTitle builds AddTask for "refuses an empty title": the use case and the input it
// refuses with task.ErrEmptyTitle.
func givenAddTaskRefusesAnEmptyTitle(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), newMemTasks(), &memJournal{}), AddTaskInput{Title: "   "}
}

// givenAddTaskRefusesALongTitle builds AddTask for "refuses a long title": the use case and the input it
// refuses with task.ErrTitleTooLong.
func givenAddTaskRefusesALongTitle(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), newMemTasks(), &memJournal{}), AddTaskInput{Title: strings.Repeat("é", task.MaxTitle+1)}
}

// givenAddTaskRefusesAnUnknownPriority builds AddTask for "refuses an unknown priority": the use case and the input it
// refuses with task.ErrUnknownPriority.
func givenAddTaskRefusesAnUnknownPriority(t *testing.T) (AddTask, AddTaskInput) {
	t.Helper()
	return NewAddTask(fixedClock(epoch), newMemTasks(), &memJournal{}), AddTaskInput{Title: "a task", Priority: "urgent"}
}
