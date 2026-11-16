// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:b86d0cd87931d5d99bbe4360b3b902a3632eea398c69f6b97c0556cdb2a00c3c

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

// TestCountTasksCountsEveryState is CountTasks's scenario "counts every state" (success).
func TestCountTasksCountsEveryState(t *testing.T) {
	uc, in, then := givenCountTasksCountsEveryState(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("CountTasks: %v, want a success", err)
	}
	then(t, out)
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

// TestDeleteTaskRefusesAMemberItIsSharedWith is DeleteTask's scenario "refuses a member it is shared with" (refusal).
func TestDeleteTaskRefusesAMemberItIsSharedWith(t *testing.T) {
	uc, in := givenDeleteTaskRefusesAMemberItIsSharedWith(t)
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

// TestEditTaskRefusesAnEmptyTitle is EditTask's scenario "refuses an empty title" (refusal).
func TestEditTaskRefusesAnEmptyTitle(t *testing.T) {
	uc, in := givenEditTaskRefusesAnEmptyTitle(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrEmptyTitle) {
		t.Fatalf("EditTask: %v, want task.ErrEmptyTitle", err)
	}
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

// TestGetTaskRefusesATaskTheMemberDoesNotSee is GetTask's scenario "refuses a task the member does not see" (refusal).
func TestGetTaskRefusesATaskTheMemberDoesNotSee(t *testing.T) {
	uc, in := givenGetTaskRefusesATaskTheMemberDoesNotSee(t)
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

// RecordPlannerTaskStoreV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordPlannerTaskStoreV1(next TaskStoreV1) TaskStoreV1 {
	return recordedplannerTaskStoreV1{next: next, calls: r.calls}
}

type recordedplannerTaskStoreV1 struct {
	next  TaskStoreV1
	calls *[]string
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
	want := []string{"tasks/port/TaskStore@v1.Get", "tasks/port/Clock@v1.Now", "tasks/port/TaskStore@v1.Save", "tasks/port/Journal@v1.Record"}
	if !slices.Equal(got, want) {
		t.Errorf("the calls: %q, want %q", got, want)
	}
}
