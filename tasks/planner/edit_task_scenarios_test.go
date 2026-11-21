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
	return NewEditTask(newMemTasks(before), crew(), fixedClock(later), journal), in, func(t *testing.T, out EditTaskOutput) {
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
	journal := &memJournal{}
	return NewEditTask(newMemTasks(grouped("task_a")), crew(), fixedClock(epoch), journal), in, func(t *testing.T, out EditTaskOutput) {
		t.Helper()
		if out.Task.AssigneeID != bob || out.Task.GroupID != core || out.Task.OwnerID != ada {
			t.Errorf("the task Cy assigned to Bob: %+v", out.Task)
		}
		if len(journal.told) != 1 || journal.told[0] != "assigned to "+bob {
			t.Errorf("who was told of the assignment: %q, want Bob alone", journal.told)
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
	return NewEditTask(newMemTasks(assigned), crew(), fixedClock(epoch), &memJournal{}), in, func(t *testing.T, out EditTaskOutput) {
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
	return NewEditTask(newMemTasks(open("task_a")), crew(), fixedClock(epoch), &memJournal{}), EditTaskInput{User: ada, ID: "task_a", Title: &blank}
}

// givenEditTaskRefusesAnAssigneeOutsideTheGroup builds EditTask for "refuses an assignee outside the group": the use case and the input it
// refuses with task.ErrNotAssignable.
func givenEditTaskRefusesAnAssigneeOutsideTheGroup(t *testing.T) (EditTask, EditTaskInput) {
	t.Helper()
	who := dee
	return NewEditTask(newMemTasks(grouped("task_a")), crew(), fixedClock(epoch), &memJournal{}), EditTaskInput{User: ada, ID: "task_a", AssigneeID: &who}
}

// givenEditTaskAssignsAMemberTheTaskIsSharedWith builds EditTask for "assigns a member the task is shared with": the use case, its input, and
// what its output must be.
func givenEditTaskAssignsAMemberTheTaskIsSharedWith(t *testing.T) (EditTask, EditTaskInput, func(*testing.T, EditTaskOutput)) {
	t.Helper()
	who := bob
	in := EditTaskInput{User: ada, ID: "task_a", AssigneeID: &who}
	return NewEditTask(newMemTasks(shared("task_a")), crew(), fixedClock(epoch), &memJournal{}), in, func(t *testing.T, out EditTaskOutput) {
		t.Helper()
		if out.Task.AssigneeID != bob {
			t.Errorf("the task assigned to who it is shared with: %+v", out.Task)
		}
	}
}

// givenEditTaskRefusesAContactTheTaskIsNotSharedWith builds EditTask for "refuses a contact the task is not shared with": the use case and the input it
// refuses with task.ErrNotAssignable.
func givenEditTaskRefusesAContactTheTaskIsNotSharedWith(t *testing.T) (EditTask, EditTaskInput) {
	t.Helper()
	// Bob, with whom the task is shared, asks for it to be assigned to Dee, who
	// sees nothing of it: only Ada, its owner, lets someone else see it.
	who := dee
	return NewEditTask(newMemTasks(shared("task_a")), crew(), fixedClock(epoch), &memJournal{}), EditTaskInput{User: bob, ID: "task_a", AssigneeID: &who}
}

// givenEditTaskMovesATaskToAnotherGroupOfItsOwner builds EditTask for "moves a task to another group of its owner": the use case, its input, and
// what its output must be.
func givenEditTaskMovesATaskToAnotherGroupOfItsOwner(t *testing.T) (EditTask, EditTaskInput, func(*testing.T, EditTaskOutput)) {
	t.Helper()
	groups := crew()
	groups["group_side"] = map[string]string{ada: roleOwner, dee: "member"}
	moved := grouped("task_a")
	moved.AssigneeID = ada // its owner, who is in both groups
	side, journal := "group_side", &memJournal{}
	uc := NewEditTask(newMemTasks(moved), groups, fixedClock(epoch), journal)
	return uc, EditTaskInput{User: ada, ID: "task_a", GroupID: &side}, func(t *testing.T, out EditTaskOutput) {
		t.Helper()
		if out.Task.GroupID != side || out.Task.AssigneeID != ada {
			t.Errorf("the task moved: %+v, want it on the other group, still its owner's to do", out.Task)
		}
		none := ""
		back, err := uc.Execute(t.Context(), EditTaskInput{User: ada, ID: "task_a", GroupID: &none})
		if err != nil || back.Task.GroupID != "" {
			t.Errorf("moved to no group: %+v (%v)", back.Task, err)
		}
		if len(journal.told) != 0 {
			t.Errorf("nobody is newly assigned, and %q was told", journal.told)
		}
	}
}

// givenEditTaskClearsAnAssigneeWhoIsNotInTheNewGroup builds EditTask for "clears an assignee who is not in the new group": the use case, its input, and
// what its output must be.
func givenEditTaskClearsAnAssigneeWhoIsNotInTheNewGroup(t *testing.T) (EditTask, EditTaskInput, func(*testing.T, EditTaskOutput)) {
	t.Helper()
	groups := crew()
	groups["group_side"] = map[string]string{ada: roleOwner}
	moved := grouped("task_a")
	moved.AssigneeID = cy // in the first group, not in the other
	side := "group_side"
	uc := NewEditTask(newMemTasks(moved), groups, fixedClock(epoch), &memJournal{})
	return uc, EditTaskInput{User: ada, ID: "task_a", GroupID: &side}, func(t *testing.T, out EditTaskOutput) {
		t.Helper()
		if out.Task.GroupID != side || out.Task.AssigneeID != "" {
			t.Errorf("the task moved: %+v, want it on the other group and unassigned", out.Task)
		}
	}
}

// givenEditTaskRefusesAMemberWhoMovesTheTask builds EditTask for "refuses a member who moves the task": the use case and the input it
// refuses with task.ErrNotOwner.
func givenEditTaskRefusesAMemberWhoMovesTheTask(t *testing.T) (EditTask, EditTaskInput) {
	t.Helper()
	none := ""
	return NewEditTask(newMemTasks(grouped("task_a")), crew(), fixedClock(epoch), &memJournal{}), EditTaskInput{User: bob, ID: "task_a", GroupID: &none}
}

// givenEditTaskRefusesAGroupItsOwnerIsNotIn builds EditTask for "refuses a group its owner is not in": the use case and the input it
// refuses with task.ErrNotMember.
func givenEditTaskRefusesAGroupItsOwnerIsNotIn(t *testing.T) (EditTask, EditTaskInput) {
	t.Helper()
	other := "group_other"
	return NewEditTask(newMemTasks(open("task_a")), crew(), fixedClock(epoch), &memJournal{}), EditTaskInput{User: ada, ID: "task_a", GroupID: &other}
}
