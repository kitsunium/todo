// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"testing"
	"time"

	"github.com/kitsunium/todo/tasks/task"
)

// givenEditTaskChangesOnlyWhatIsGiven builds EditTask for "changes only what is given": the use case, its input, and
// what its output must be.
func givenEditTaskChangesOnlyWhatIsGiven(t *testing.T) (EditTask, EditTaskInput, func(*testing.T, EditTaskOutput)) {
	t.Helper()
	before := shared("task_a")
	before.Notes, before.AssigneeID = "the notes", cy
	journal := &memJournal{}
	later := epoch.Add(time.Hour)
	title, priority := "  A better title ", "high"
	in := EditTaskInput{User: bob, ID: "task_a", Title: &title, Priority: &priority}
	return NewEditTask(newMemTasks(before), crew(), contacts{}, fixedClock(later), journal), in, func(t *testing.T, out EditTaskOutput) {
		t.Helper()
		got := out.Task
		if got.Title != "A better title" || got.Priority != task.PriorityHigh || !got.UpdatedAt.Equal(later) {
			t.Errorf("what was given: %+v", got)
		}
		if got.Notes != "the notes" || got.AssigneeID != cy || got.OwnerID != ada || !got.Due.IsZero() {
			t.Errorf("what was not given changed: %+v", got)
		}
		if len(journal.kinds) != 1 || journal.kinds[0] != "edited" {
			t.Errorf("the journal was told %q, want edited once", journal.kinds)
		}
	}
}

// givenEditTaskAssignsAMemberOfTheGroup builds EditTask for "assigns a member of the group": the use case, its input, and
// what its output must be.
func givenEditTaskAssignsAMemberOfTheGroup(t *testing.T) (EditTask, EditTaskInput, func(*testing.T, EditTaskOutput)) {
	t.Helper()
	who := bob
	in := EditTaskInput{User: cy, ID: "task_a", AssigneeID: &who}
	return NewEditTask(newMemTasks(grouped("task_a")), crew(), contacts{}, fixedClock(epoch), &memJournal{}), in, func(t *testing.T, out EditTaskOutput) {
		t.Helper()
		if out.Task.AssigneeID != bob || out.Task.GroupID != core || out.Task.OwnerID != ada {
			t.Errorf("the task Cy assigned to Bob: %+v", out.Task)
		}
	}
}

// givenEditTaskClearsTheAssignee builds EditTask for "clears the assignee": the use case, its input, and
// what its output must be.
func givenEditTaskClearsTheAssignee(t *testing.T) (EditTask, EditTaskInput, func(*testing.T, EditTaskOutput)) {
	t.Helper()
	assigned := open("task_a")
	assigned.AssigneeID = dee
	none := " "
	in := EditTaskInput{User: dee, ID: "task_a", AssigneeID: &none} // the assignee sees it, and edits it
	return NewEditTask(newMemTasks(assigned), crew(), contacts{}, fixedClock(epoch), &memJournal{}), in, func(t *testing.T, out EditTaskOutput) {
		t.Helper()
		if out.Task.AssigneeID != "" {
			t.Errorf("the assignee is still %q", out.Task.AssigneeID)
		}
	}
}

// givenEditTaskRefusesAnEmptyTitle builds EditTask for "refuses an empty title": the use case and the input it
// refuses with task.ErrEmptyTitle.
func givenEditTaskRefusesAnEmptyTitle(t *testing.T) (EditTask, EditTaskInput) {
	t.Helper()
	blank := "   "
	return NewEditTask(newMemTasks(open("task_a")), crew(), contacts{}, fixedClock(epoch), &memJournal{}), EditTaskInput{User: ada, ID: "task_a", Title: &blank}
}

// givenEditTaskRefusesAnAssigneeOutsideTheGroup builds EditTask for "refuses an assignee outside the group": the use case and the input it
// refuses with task.ErrNotAssignable.
func givenEditTaskRefusesAnAssigneeOutsideTheGroup(t *testing.T) (EditTask, EditTaskInput) {
	t.Helper()
	who := dee
	return NewEditTask(newMemTasks(grouped("task_a")), crew(), contacts{ada + "|" + dee: true}, fixedClock(epoch), &memJournal{}), EditTaskInput{User: ada, ID: "task_a", AssigneeID: &who}
}
