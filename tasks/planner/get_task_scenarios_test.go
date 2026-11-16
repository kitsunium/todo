// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"testing"
)

// givenGetTaskReadsATaskSharedWithTheMember builds GetTask for "reads a task shared with the member": the use case, its input, and
// what its output must be.
func givenGetTaskReadsATaskSharedWithTheMember(t *testing.T) (GetTask, GetTaskInput, func(*testing.T, GetTaskOutput)) {
	t.Helper()
	return NewGetTask(newMemTasks(shared("task_a")), crew()), GetTaskInput{User: bob, ID: "task_a"}, func(t *testing.T, out GetTaskOutput) {
		t.Helper()
		if out.Task.ID != "task_a" || out.Task.OwnerID != ada {
			t.Errorf("the task read: %+v", out.Task)
		}
	}
}

// givenGetTaskReadsATaskOfAGroupOfTheMember builds GetTask for "reads a task of a group of the member": the use case, its input, and
// what its output must be.
func givenGetTaskReadsATaskOfAGroupOfTheMember(t *testing.T) (GetTask, GetTaskInput, func(*testing.T, GetTaskOutput)) {
	t.Helper()
	return NewGetTask(newMemTasks(grouped("task_a")), crew()), GetTaskInput{User: cy, ID: "task_a"}, func(t *testing.T, out GetTaskOutput) {
		t.Helper()
		if out.Task.ID != "task_a" || out.Task.GroupID != core {
			t.Errorf("the task read: %+v", out.Task)
		}
	}
}

// givenGetTaskReadsATaskAssignedToTheMember builds GetTask for "reads a task assigned to the member": the use case, its input, and
// what its output must be.
func givenGetTaskReadsATaskAssignedToTheMember(t *testing.T) (GetTask, GetTaskInput, func(*testing.T, GetTaskOutput)) {
	t.Helper()
	assigned := open("task_a")
	assigned.AssigneeID = dee
	return NewGetTask(newMemTasks(assigned), crew()), GetTaskInput{User: dee, ID: "task_a"}, func(t *testing.T, out GetTaskOutput) {
		t.Helper()
		if out.Task.ID != "task_a" || out.Task.AssigneeID != dee {
			t.Errorf("the task read: %+v", out.Task)
		}
	}
}

// givenGetTaskRefusesATaskTheMemberDoesNotSee builds GetTask for "refuses a task the member does not see": the use case and the input it
// refuses with task.ErrNotFound.
func givenGetTaskRefusesATaskTheMemberDoesNotSee(t *testing.T) (GetTask, GetTaskInput) {
	t.Helper()
	return NewGetTask(newMemTasks(shared("task_a")), crew()), GetTaskInput{User: cy, ID: "task_a"}
}

// givenGetTaskRefusesAGroupTaskToWhoLeftTheGroup builds GetTask for "refuses a group task to who left the group": the use case and the input it
// refuses with task.ErrNotFound.
func givenGetTaskRefusesAGroupTaskToWhoLeftTheGroup(t *testing.T) (GetTask, GetTaskInput) {
	t.Helper()
	return NewGetTask(newMemTasks(grouped("task_a")), crew()), GetTaskInput{User: dee, ID: "task_a"}
}
