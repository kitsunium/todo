// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:e049376e6f3e39ba5615be3b676d6ca97a5b1068d722cafef1679ea63a05c50d

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

// TestArchiveDoneArchivesWhatWasDoneLongAgo is ArchiveDone's scenario "archives what was done long ago" (success).
func TestArchiveDoneArchivesWhatWasDoneLongAgo(t *testing.T) {
	uc, in, then := givenArchiveDoneArchivesWhatWasDoneLongAgo(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ArchiveDone: %v, want a success", err)
	}
	then(t, out)
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

// TestDueSoonListsTheOpenTasksDueSoon is DueSoon's scenario "lists the open tasks due soon" (success).
func TestDueSoonListsTheOpenTasksDueSoon(t *testing.T) {
	uc, in, then := givenDueSoonListsTheOpenTasksDueSoon(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("DueSoon: %v, want a success", err)
	}
	then(t, out)
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

// TestMoveTaskRefusesToArchiveAnOpenTask is MoveTask's scenario "refuses to archive an open task" (refusal).
func TestMoveTaskRefusesToArchiveAnOpenTask(t *testing.T) {
	uc, in := givenMoveTaskRefusesToArchiveAnOpenTask(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, task.ErrNotAllowed) {
		t.Fatalf("MoveTask: %v, want task.ErrNotAllowed", err)
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
