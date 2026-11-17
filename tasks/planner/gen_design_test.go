// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:6a4c0cc7ef82f45cbf05f7471fb6e2644c468fdaa00e78e6aa4e3dd63779cfee

package planner

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/kitsunium/todo/tasks/task"
)

// TestAddTaskAddsAnOpenTask is AddTask's scenario "adds an open task" (success).
func TestAddTaskAddsAnOpenTask(t *testing.T) {
	uc, in, then := givenAddTaskAddsAnOpenTask(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("AddTask: %v, want a success", err)
	}
	then(t, out)
}

// TestAddTaskTellsItsAssigneeOfATaskAssigned is AddTask's scenario "tells its assignee of a task assigned" (success).
func TestAddTaskTellsItsAssigneeOfATaskAssigned(t *testing.T) {
	uc, in, then := givenAddTaskTellsItsAssigneeOfATaskAssigned(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("AddTask: %v, want a success", err)
	}
	then(t, out)
}

// TestAddTaskAddsATaskToAGroupList is AddTask's scenario "adds a task to a group list" (success).
func TestAddTaskAddsATaskToAGroupList(t *testing.T) {
	uc, in, then := givenAddTaskAddsATaskToAGroupList(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("AddTask: %v, want a success", err)
	}
	then(t, out)
}

// TestAddTaskRefusesAnEmptyTitle is AddTask's scenario "refuses an empty title" (refusal).
func TestAddTaskRefusesAnEmptyTitle(t *testing.T) {
	uc, in := givenAddTaskRefusesAnEmptyTitle(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrEmptyTitle) {
		t.Fatalf("AddTask: %v, want task.ErrEmptyTitle", err)
	}
}

// TestAddTaskRefusesALongTitle is AddTask's scenario "refuses a long title" (refusal).
func TestAddTaskRefusesALongTitle(t *testing.T) {
	uc, in := givenAddTaskRefusesALongTitle(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrTitleTooLong) {
		t.Fatalf("AddTask: %v, want task.ErrTitleTooLong", err)
	}
}

// TestAddTaskRefusesAnUnknownPriority is AddTask's scenario "refuses an unknown priority" (refusal).
func TestAddTaskRefusesAnUnknownPriority(t *testing.T) {
	uc, in := givenAddTaskRefusesAnUnknownPriority(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrUnknownPriority) {
		t.Fatalf("AddTask: %v, want task.ErrUnknownPriority", err)
	}
}

// TestAddTaskRefusesAGroupTheMemberIsNotIn is AddTask's scenario "refuses a group the member is not in" (refusal).
func TestAddTaskRefusesAGroupTheMemberIsNotIn(t *testing.T) {
	uc, in := givenAddTaskRefusesAGroupTheMemberIsNotIn(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotMember) {
		t.Fatalf("AddTask: %v, want task.ErrNotMember", err)
	}
}

// TestAddTaskRefusesAnAssigneeWhoDoesNotSeeTheTask is AddTask's scenario "refuses an assignee who does not see the task" (refusal).
func TestAddTaskRefusesAnAssigneeWhoDoesNotSeeTheTask(t *testing.T) {
	uc, in := givenAddTaskRefusesAnAssigneeWhoDoesNotSeeTheTask(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotAssignable) {
		t.Fatalf("AddTask: %v, want task.ErrNotAssignable", err)
	}
}

// TestAddTaskRefusesAnAssigneeOutsideTheGroup is AddTask's scenario "refuses an assignee outside the group" (refusal).
func TestAddTaskRefusesAnAssigneeOutsideTheGroup(t *testing.T) {
	uc, in := givenAddTaskRefusesAnAssigneeOutsideTheGroup(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotAssignable) {
		t.Fatalf("AddTask: %v, want task.ErrNotAssignable", err)
	}
}

// TestCountTasksCountsEveryState is CountTasks's scenario "counts every state" (success).
func TestCountTasksCountsEveryState(t *testing.T) {
	uc, in, then := givenCountTasksCountsEveryState(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("CountTasks: %v, want a success", err)
	}
	then(t, out)
}

// TestCountViewsCountsEveryView is CountViews's scenario "counts every view" (success).
func TestCountViewsCountsEveryView(t *testing.T) {
	uc, in, then := givenCountViewsCountsEveryView(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("CountViews: %v, want a success", err)
	}
	then(t, out)
}

// TestCountViewsRefusesAnUnknownTimeZone is CountViews's scenario "refuses an unknown time zone" (refusal).
func TestCountViewsRefusesAnUnknownTimeZone(t *testing.T) {
	uc, in := givenCountViewsRefusesAnUnknownTimeZone(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrBadZone) {
		t.Fatalf("CountViews: %v, want task.ErrBadZone", err)
	}
}

// TestDeleteTaskDeletesATaskOfTheOwner is DeleteTask's scenario "deletes a task of the owner" (success).
func TestDeleteTaskDeletesATaskOfTheOwner(t *testing.T) {
	uc, in, then := givenDeleteTaskDeletesATaskOfTheOwner(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("DeleteTask: %v, want a success", err)
	}
	then(t, out)
}

// TestDeleteTaskDeletesATaskOfTheGroupAsItsAdmin is DeleteTask's scenario "deletes a task of the group as its admin" (success).
func TestDeleteTaskDeletesATaskOfTheGroupAsItsAdmin(t *testing.T) {
	uc, in, then := givenDeleteTaskDeletesATaskOfTheGroupAsItsAdmin(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("DeleteTask: %v, want a success", err)
	}
	then(t, out)
}

// TestDeleteTaskRefusesAMemberItIsSharedWith is DeleteTask's scenario "refuses a member it is shared with" (refusal).
func TestDeleteTaskRefusesAMemberItIsSharedWith(t *testing.T) {
	uc, in := givenDeleteTaskRefusesAMemberItIsSharedWith(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotOwner) {
		t.Fatalf("DeleteTask: %v, want task.ErrNotOwner", err)
	}
}

// TestDeleteTaskRefusesAPlainMemberOfTheGroup is DeleteTask's scenario "refuses a plain member of the group" (refusal).
func TestDeleteTaskRefusesAPlainMemberOfTheGroup(t *testing.T) {
	uc, in := givenDeleteTaskRefusesAPlainMemberOfTheGroup(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotOwner) {
		t.Fatalf("DeleteTask: %v, want task.ErrNotOwner", err)
	}
}

// TestDueSoonListsTheOpenTasksDueSoon is DueSoon's scenario "lists the open tasks due soon" (success).
func TestDueSoonListsTheOpenTasksDueSoon(t *testing.T) {
	uc, in, then := givenDueSoonListsTheOpenTasksDueSoon(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("DueSoon: %v, want a success", err)
	}
	then(t, out)
}

// TestEditTaskChangesOnlyWhatIsGiven is EditTask's scenario "changes only what is given" (success).
func TestEditTaskChangesOnlyWhatIsGiven(t *testing.T) {
	uc, in, then := givenEditTaskChangesOnlyWhatIsGiven(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("EditTask: %v, want a success", err)
	}
	then(t, out)
}

// TestEditTaskMovesATaskToAnotherGroupOfItsOwner is EditTask's scenario "moves a task to another group of its owner" (success).
func TestEditTaskMovesATaskToAnotherGroupOfItsOwner(t *testing.T) {
	uc, in, then := givenEditTaskMovesATaskToAnotherGroupOfItsOwner(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("EditTask: %v, want a success", err)
	}
	then(t, out)
}

// TestEditTaskClearsAnAssigneeWhoIsNotInTheNewGroup is EditTask's scenario "clears an assignee who is not in the new group" (success).
func TestEditTaskClearsAnAssigneeWhoIsNotInTheNewGroup(t *testing.T) {
	uc, in, then := givenEditTaskClearsAnAssigneeWhoIsNotInTheNewGroup(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("EditTask: %v, want a success", err)
	}
	then(t, out)
}

// TestEditTaskRefusesAMemberWhoMovesTheTask is EditTask's scenario "refuses a member who moves the task" (refusal).
func TestEditTaskRefusesAMemberWhoMovesTheTask(t *testing.T) {
	uc, in := givenEditTaskRefusesAMemberWhoMovesTheTask(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotOwner) {
		t.Fatalf("EditTask: %v, want task.ErrNotOwner", err)
	}
}

// TestEditTaskRefusesAGroupItsOwnerIsNotIn is EditTask's scenario "refuses a group its owner is not in" (refusal).
func TestEditTaskRefusesAGroupItsOwnerIsNotIn(t *testing.T) {
	uc, in := givenEditTaskRefusesAGroupItsOwnerIsNotIn(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotMember) {
		t.Fatalf("EditTask: %v, want task.ErrNotMember", err)
	}
}

// TestEditTaskAssignsAMemberOfTheGroup is EditTask's scenario "assigns a member of the group" (success).
func TestEditTaskAssignsAMemberOfTheGroup(t *testing.T) {
	uc, in, then := givenEditTaskAssignsAMemberOfTheGroup(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("EditTask: %v, want a success", err)
	}
	then(t, out)
}

// TestEditTaskAssignsAMemberTheTaskIsSharedWith is EditTask's scenario "assigns a member the task is shared with" (success).
func TestEditTaskAssignsAMemberTheTaskIsSharedWith(t *testing.T) {
	uc, in, then := givenEditTaskAssignsAMemberTheTaskIsSharedWith(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("EditTask: %v, want a success", err)
	}
	then(t, out)
}

// TestEditTaskRefusesAContactTheTaskIsNotSharedWith is EditTask's scenario "refuses a contact the task is not shared with" (refusal).
func TestEditTaskRefusesAContactTheTaskIsNotSharedWith(t *testing.T) {
	uc, in := givenEditTaskRefusesAContactTheTaskIsNotSharedWith(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotAssignable) {
		t.Fatalf("EditTask: %v, want task.ErrNotAssignable", err)
	}
}

// TestEditTaskClearsTheAssignee is EditTask's scenario "clears the assignee" (success).
func TestEditTaskClearsTheAssignee(t *testing.T) {
	uc, in, then := givenEditTaskClearsTheAssignee(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("EditTask: %v, want a success", err)
	}
	then(t, out)
}

// TestEditTaskRefusesAnEmptyTitle is EditTask's scenario "refuses an empty title" (refusal).
func TestEditTaskRefusesAnEmptyTitle(t *testing.T) {
	uc, in := givenEditTaskRefusesAnEmptyTitle(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrEmptyTitle) {
		t.Fatalf("EditTask: %v, want task.ErrEmptyTitle", err)
	}
}

// TestEditTaskRefusesAnAssigneeOutsideTheGroup is EditTask's scenario "refuses an assignee outside the group" (refusal).
func TestEditTaskRefusesAnAssigneeOutsideTheGroup(t *testing.T) {
	uc, in := givenEditTaskRefusesAnAssigneeOutsideTheGroup(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotAssignable) {
		t.Fatalf("EditTask: %v, want task.ErrNotAssignable", err)
	}
}

// TestFollowGroupSendsTheTasksOfADeletedGroupBackToTheirOwners is FollowGroup's scenario "sends the tasks of a deleted group back to their owners" (success).
func TestFollowGroupSendsTheTasksOfADeletedGroupBackToTheirOwners(t *testing.T) {
	uc, in, then := givenFollowGroupSendsTheTasksOfADeletedGroupBackToTheirOwners(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("FollowGroup: %v, want a success", err)
	}
	then(t, out)
}

// TestFollowGroupUnassignsWhoLeftTheGroup is FollowGroup's scenario "unassigns who left the group" (success).
func TestFollowGroupUnassignsWhoLeftTheGroup(t *testing.T) {
	uc, in, then := givenFollowGroupUnassignsWhoLeftTheGroup(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("FollowGroup: %v, want a success", err)
	}
	then(t, out)
}

// TestGetTaskReadsATaskSharedWithTheMember is GetTask's scenario "reads a task shared with the member" (success).
func TestGetTaskReadsATaskSharedWithTheMember(t *testing.T) {
	uc, in, then := givenGetTaskReadsATaskSharedWithTheMember(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("GetTask: %v, want a success", err)
	}
	then(t, out)
}

// TestGetTaskReadsATaskOfAGroupOfTheMember is GetTask's scenario "reads a task of a group of the member" (success).
func TestGetTaskReadsATaskOfAGroupOfTheMember(t *testing.T) {
	uc, in, then := givenGetTaskReadsATaskOfAGroupOfTheMember(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("GetTask: %v, want a success", err)
	}
	then(t, out)
}

// TestGetTaskReadsATaskAssignedToTheMember is GetTask's scenario "reads a task assigned to the member" (success).
func TestGetTaskReadsATaskAssignedToTheMember(t *testing.T) {
	uc, in, then := givenGetTaskReadsATaskAssignedToTheMember(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("GetTask: %v, want a success", err)
	}
	then(t, out)
}

// TestGetTaskRefusesATaskTheMemberDoesNotSee is GetTask's scenario "refuses a task the member does not see" (refusal).
func TestGetTaskRefusesATaskTheMemberDoesNotSee(t *testing.T) {
	uc, in := givenGetTaskRefusesATaskTheMemberDoesNotSee(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotFound) {
		t.Fatalf("GetTask: %v, want task.ErrNotFound", err)
	}
}

// TestGetTaskRefusesAGroupTaskToWhoLeftTheGroup is GetTask's scenario "refuses a group task to who left the group" (refusal).
func TestGetTaskRefusesAGroupTaskToWhoLeftTheGroup(t *testing.T) {
	uc, in := givenGetTaskRefusesAGroupTaskToWhoLeftTheGroup(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotFound) {
		t.Fatalf("GetTask: %v, want task.ErrNotFound", err)
	}
}

// TestListTasksListsTheMostUrgentFirst is ListTasks's scenario "lists the most urgent first" (success).
func TestListTasksListsTheMostUrgentFirst(t *testing.T) {
	uc, in, then := givenListTasksListsTheMostUrgentFirst(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ListTasks: %v, want a success", err)
	}
	then(t, out)
}

// TestListTasksListsTheTasksOfAGroup is ListTasks's scenario "lists the tasks of a group" (success).
func TestListTasksListsTheTasksOfAGroup(t *testing.T) {
	uc, in, then := givenListTasksListsTheTasksOfAGroup(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ListTasks: %v, want a success", err)
	}
	then(t, out)
}

// TestListTasksListsAView is ListTasks's scenario "lists a view" (success).
func TestListTasksListsAView(t *testing.T) {
	uc, in, then := givenListTasksListsAView(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ListTasks: %v, want a success", err)
	}
	then(t, out)
}

// TestListTasksRefusesAnUnknownView is ListTasks's scenario "refuses an unknown view" (refusal).
func TestListTasksRefusesAnUnknownView(t *testing.T) {
	uc, in := givenListTasksRefusesAnUnknownView(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrUnknownView) {
		t.Fatalf("ListTasks: %v, want task.ErrUnknownView", err)
	}
}

// TestListTasksRefusesTheGroupViewWithoutAGroup is ListTasks's scenario "refuses the group view without a group" (refusal).
func TestListTasksRefusesTheGroupViewWithoutAGroup(t *testing.T) {
	uc, in := givenListTasksRefusesTheGroupViewWithoutAGroup(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrGroupNeeded) {
		t.Fatalf("ListTasks: %v, want task.ErrGroupNeeded", err)
	}
}

// TestListTasksRefusesAnUnknownTimeZone is ListTasks's scenario "refuses an unknown time zone" (refusal).
func TestListTasksRefusesAnUnknownTimeZone(t *testing.T) {
	uc, in := givenListTasksRefusesAnUnknownTimeZone(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrBadZone) {
		t.Fatalf("ListTasks: %v, want task.ErrBadZone", err)
	}
}

// TestListTasksRefusesAGroupTheMemberIsNotIn is ListTasks's scenario "refuses a group the member is not in" (refusal).
func TestListTasksRefusesAGroupTheMemberIsNotIn(t *testing.T) {
	uc, in := givenListTasksRefusesAGroupTheMemberIsNotIn(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNoGroup) {
		t.Fatalf("ListTasks: %v, want task.ErrNoGroup", err)
	}
}

// TestMoveTaskCompletesAnOpenTask is MoveTask's scenario "completes an open task" (success).
func TestMoveTaskCompletesAnOpenTask(t *testing.T) {
	uc, in, then := givenMoveTaskCompletesAnOpenTask(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("MoveTask: %v, want a success", err)
	}
	then(t, out)
}

// TestMoveTaskRefusesAnUnknownTask is MoveTask's scenario "refuses an unknown task" (refusal).
func TestMoveTaskRefusesAnUnknownTask(t *testing.T) {
	uc, in := givenMoveTaskRefusesAnUnknownTask(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotFound) {
		t.Fatalf("MoveTask: %v, want task.ErrNotFound", err)
	}
}

// TestMoveTaskRefusesTheTaskOfAnotherMember is MoveTask's scenario "refuses the task of another member" (refusal).
func TestMoveTaskRefusesTheTaskOfAnotherMember(t *testing.T) {
	uc, in := givenMoveTaskRefusesTheTaskOfAnotherMember(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotFound) {
		t.Fatalf("MoveTask: %v, want task.ErrNotFound", err)
	}
}

// TestMoveTaskMovesATaskSharedWithTheMember is MoveTask's scenario "moves a task shared with the member" (success).
func TestMoveTaskMovesATaskSharedWithTheMember(t *testing.T) {
	uc, in, then := givenMoveTaskMovesATaskSharedWithTheMember(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("MoveTask: %v, want a success", err)
	}
	then(t, out)
}

// TestMoveTaskMovesATaskOfAGroupOfTheMember is MoveTask's scenario "moves a task of a group of the member" (success).
func TestMoveTaskMovesATaskOfAGroupOfTheMember(t *testing.T) {
	uc, in, then := givenMoveTaskMovesATaskOfAGroupOfTheMember(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("MoveTask: %v, want a success", err)
	}
	then(t, out)
}

// TestMoveTaskMovesATaskAssignedToTheMember is MoveTask's scenario "moves a task assigned to the member" (success).
func TestMoveTaskMovesATaskAssignedToTheMember(t *testing.T) {
	uc, in, then := givenMoveTaskMovesATaskAssignedToTheMember(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("MoveTask: %v, want a success", err)
	}
	then(t, out)
}

// TestMoveTaskRefusesToArchiveAnOpenTask is MoveTask's scenario "refuses to archive an open task" (refusal).
func TestMoveTaskRefusesToArchiveAnOpenTask(t *testing.T) {
	uc, in := givenMoveTaskRefusesToArchiveAnOpenTask(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotAllowed) {
		t.Fatalf("MoveTask: %v, want task.ErrNotAllowed", err)
	}
}

// TestShareTaskSharesWithAContact is ShareTask's scenario "shares with a contact" (success).
func TestShareTaskSharesWithAContact(t *testing.T) {
	uc, in, then := givenShareTaskSharesWithAContact(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ShareTask: %v, want a success", err)
	}
	then(t, out)
}

// TestShareTaskRefusesWhoIsNoContact is ShareTask's scenario "refuses who is no contact" (refusal).
func TestShareTaskRefusesWhoIsNoContact(t *testing.T) {
	uc, in := givenShareTaskRefusesWhoIsNoContact(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotContact) {
		t.Fatalf("ShareTask: %v, want task.ErrNotContact", err)
	}
}

// TestShareTaskRefusesWhoIsNotTheOwner is ShareTask's scenario "refuses who is not the owner" (refusal).
func TestShareTaskRefusesWhoIsNotTheOwner(t *testing.T) {
	uc, in := givenShareTaskRefusesWhoIsNotTheOwner(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotOwner) {
		t.Fatalf("ShareTask: %v, want task.ErrNotOwner", err)
	}
}

// TestShareTaskRefusesAMemberOfTheGroupWhoIsNotTheOwner is ShareTask's scenario "refuses a member of the group who is not the owner" (refusal).
func TestShareTaskRefusesAMemberOfTheGroupWhoIsNotTheOwner(t *testing.T) {
	uc, in := givenShareTaskRefusesAMemberOfTheGroupWhoIsNotTheOwner(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotOwner) {
		t.Fatalf("ShareTask: %v, want task.ErrNotOwner", err)
	}
}

// TestUnshareTaskLetsAMemberLeaveATask is UnshareTask's scenario "lets a member leave a task" (success).
func TestUnshareTaskLetsAMemberLeaveATask(t *testing.T) {
	uc, in, then := givenUnshareTaskLetsAMemberLeaveATask(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("UnshareTask: %v, want a success", err)
	}
	then(t, out)
}

// TestUnshareTaskRefusesAMemberWhoRemovesAnother is UnshareTask's scenario "refuses a member who removes another" (refusal).
func TestUnshareTaskRefusesAMemberWhoRemovesAnother(t *testing.T) {
	uc, in := givenUnshareTaskRefusesAMemberWhoRemovesAnother(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotOwner) {
		t.Fatalf("UnshareTask: %v, want task.ErrNotOwner", err)
	}
}

// sequenceRecorder wraps the ports a sequence goes through, and records each
// call in order.
type sequenceRecorder struct {
	calls *[]string
}

// RecordPlannerClockV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordPlannerClockV1(next ClockV1) ClockV1 {
	return recordedplannerClockV1{next: next, calls: r.calls}
}

type recordedplannerClockV1 struct {
	next  ClockV1
	calls *[]string
}

func (x recordedplannerClockV1) Now(ctx context.Context) (time.Time, error) {
	*x.calls = append(*x.calls, "tasks/port/Clock@v1.Now")
	return x.next.Now(ctx)
}

// RecordPlannerGroupsV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordPlannerGroupsV1(next GroupsV1) GroupsV1 {
	return recordedplannerGroupsV1{next: next, calls: r.calls}
}

type recordedplannerGroupsV1 struct {
	next  GroupsV1
	calls *[]string
}

func (x recordedplannerGroupsV1) Memberships(ctx context.Context, user string) (map[string]string, error) {
	*x.calls = append(*x.calls, "tasks/port/Groups@v1.Memberships")
	return x.next.Memberships(ctx, user)
}

func (x recordedplannerGroupsV1) RoleOf(ctx context.Context, group string, user string) (string, error) {
	*x.calls = append(*x.calls, "tasks/port/Groups@v1.RoleOf")
	return x.next.RoleOf(ctx, group, user)
}

// RecordPlannerJournalV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordPlannerJournalV1(next JournalV1) JournalV1 {
	return recordedplannerJournalV1{next: next, calls: r.calls}
}

type recordedplannerJournalV1 struct {
	next  JournalV1
	calls *[]string
}

func (x recordedplannerJournalV1) Record(ctx context.Context, kind string, t task.Task) error {
	*x.calls = append(*x.calls, "tasks/port/Journal@v1.Record")
	return x.next.Record(ctx, kind, t)
}

func (x recordedplannerJournalV1) Tell(ctx context.Context, kind string, t task.Task, to string) error {
	*x.calls = append(*x.calls, "tasks/port/Journal@v1.Tell")
	return x.next.Tell(ctx, kind, t, to)
}

// RecordPlannerTaskStoreV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordPlannerTaskStoreV1(next TaskStoreV1) TaskStoreV1 {
	return recordedplannerTaskStoreV1{next: next, calls: r.calls}
}

type recordedplannerTaskStoreV1 struct {
	next  TaskStoreV1
	calls *[]string
}

func (x recordedplannerTaskStoreV1) Change(ctx context.Context, id string, change func(*task.Task) error) (task.Task, error) {
	*x.calls = append(*x.calls, "tasks/port/TaskStore@v1.Change")
	return x.next.Change(ctx, id, change)
}

func (x recordedplannerTaskStoreV1) Delete(ctx context.Context, id string) error {
	*x.calls = append(*x.calls, "tasks/port/TaskStore@v1.Delete")
	return x.next.Delete(ctx, id)
}

func (x recordedplannerTaskStoreV1) Get(ctx context.Context, id string) (task.Task, error) {
	*x.calls = append(*x.calls, "tasks/port/TaskStore@v1.Get")
	return x.next.Get(ctx, id)
}

func (x recordedplannerTaskStoreV1) List(ctx context.Context) ([]task.Task, error) {
	*x.calls = append(*x.calls, "tasks/port/TaskStore@v1.List")
	return x.next.List(ctx)
}

func (x recordedplannerTaskStoreV1) NewID(ctx context.Context) (string, error) {
	*x.calls = append(*x.calls, "tasks/port/TaskStore@v1.NewID")
	return x.next.NewID(ctx)
}

func (x recordedplannerTaskStoreV1) Save(ctx context.Context, t task.Task) error {
	*x.calls = append(*x.calls, "tasks/port/TaskStore@v1.Save")
	return x.next.Save(ctx, t)
}

// TestSequenceCompleteTask is the order of sequence complete-task: the diagram's links, in the
// diagram's order (D21).
func TestSequenceCompleteTask(t *testing.T) {
	var got []string
	run := givenSequenceCompleteTask(t, sequenceRecorder{calls: &got})
	if err := run(t.Context()); err != nil {
		t.Fatalf("the sequence failed: %v", err)
	}
	want := []string{"tasks/port/Clock@v1.Now", "tasks/port/TaskStore@v1.Get", "tasks/port/Groups@v1.RoleOf", "tasks/port/TaskStore@v1.Change", "tasks/port/Journal@v1.Record"}
	if !slices.Equal(got, want) {
		t.Errorf("the calls: %q, want %q", got, want)
	}
}
