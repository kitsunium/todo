// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:620afbb2962f72a1a0c57e91267127ee1982802bbd70150e8fa1d8268bc9f808

// Package task — The task, its priority and its lifecycle.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package task

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/pkg/v1/errs"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 1, 1, 255), "NOT_IMPLEMENTED", "not implemented yet", "task: a hand-written body is still a skeleton")

var (
	// ErrEmptyTitle is "a task needs a title".
	ErrEmptyTitle = errs.New(errs.Pack(64, 1, 1, 1), "EMPTY_TITLE", "a task needs a title", "tasks/error/ErrEmptyTitle")
	// ErrNotAllowed is "the task cannot make this move from the state it is in".
	ErrNotAllowed = errs.New(errs.Pack(64, 1, 1, 5), "NOT_ALLOWED", "the task cannot make this move from the state it is in", "tasks/error/ErrNotAllowed")
	// ErrNotFound is "no such task".
	ErrNotFound = errs.New(errs.Pack(64, 1, 1, 3), "NOT_FOUND", "no such task", "tasks/error/ErrNotFound")
	// ErrTitleTooLong is "a title is 120 characters at most".
	ErrTitleTooLong = errs.New(errs.Pack(64, 1, 1, 2), "TITLE_TOO_LONG", "a title is 120 characters at most", "tasks/error/ErrTitleTooLong")
	// ErrUnknownPriority is "a priority is low, normal or high".
	ErrUnknownPriority = errs.New(errs.Pack(64, 1, 1, 4), "UNKNOWN_PRIORITY", "a priority is low, normal or high", "tasks/error/ErrUnknownPriority")
)

const (
	// MaxTitle is the longest title a task takes.
	MaxTitle int = 120
	// PriorityHigh is what comes first.
	PriorityHigh Priority = "high"
	// PriorityLow is what can wait.
	PriorityLow Priority = "low"
	// PriorityNormal is the priority of a task that names none.
	PriorityNormal Priority = "normal"
)

// Task — Something to do, for someone, by a date.
type Task struct {
	ID string `json:"id" app:"public"`
	// OwnerID is the member whose task it is: who added it, and the only one who
	// sees it.
	OwnerID  string   `json:"ownerId" app:"public"`
	Title    string   `json:"title" app:"public"`
	Notes    string   `json:"notes" app:"public"`
	Priority Priority `json:"priority" app:"public"`
	// Assignee is who does the task, as the team names them; an address gets the
	// reminders.
	Assignee string `json:"assignee" app:"personal"`
	// Due is when the task is due; the zero time when it has no date.
	Due       time.Time      `json:"due" app:"public"`
	State     LifecycleState `json:"state" app:"public"`
	CreatedAt time.Time      `json:"createdAt" app:"public"`
	UpdatedAt time.Time      `json:"updatedAt" app:"public"`
	// DoneAt is when the task was last completed; the zero time while it is open.
	DoneAt time.Time `json:"doneAt" app:"public"`
}

// Overdue reports whether the task is still open after its due date.
func (tx *Task) Overdue(now time.Time) bool {
	return tx.overdue(now)
}

// Priority — How urgent a task is.
type Priority string

// ParsePriority reads a priority as a member writes it; an empty one is
// normal.
func ParsePriority(s string) (Priority, error) {
	return parsePriority(s)
}

// LifecycleState is a state of Lifecycle.
type LifecycleState string

// The states of Lifecycle.
const (
	LifecycleOpen     LifecycleState = "open"
	LifecycleDone     LifecycleState = "done"
	LifecycleArchived LifecycleState = "archived"
)

// LifecycleInitial is the state a Task starts in.
const LifecycleInitial = LifecycleOpen

// LifecycleEvent is an event of Lifecycle.
type LifecycleEvent string

// The events of Lifecycle.
const (
	LifecycleOnComplete LifecycleEvent = "complete"
	LifecycleOnReopen   LifecycleEvent = "reopen"
	LifecycleOnArchive  LifecycleEvent = "archive"
	LifecycleOnRestore  LifecycleEvent = "restore"
)

var (
	// ErrLifecycleRefused is a transition its guard refused.
	ErrLifecycleRefused = errs.New(errs.Pack(64, 1, 1, 240), "LIFECYCLE_REFUSED", "the transition is refused", "tasks/workflow/Lifecycle")
	// ErrLifecycleNoTransition is an event its state does not answer.
	ErrLifecycleNoTransition = errs.New(errs.Pack(64, 1, 1, 241), "LIFECYCLE_NO_TRANSITION", "no transition for this event in this state", "tasks/workflow/Lifecycle")
)

// Lifecycle — A task is open, then done, then archived; a done one reopens,
// an archived one is restored.
type Lifecycle struct {
	lifecycle // its own state, hand-written
}

// NewLifecycle builds Lifecycle: its body is the hand-written newLifecycle.
func NewLifecycle() *Lifecycle {
	return newLifecycle()
}

// Fire moves e from state from by event: the state it enters, or the state
// it stays in and why.
func (w *Lifecycle) Fire(ctx context.Context, e *Task, from LifecycleState, event LifecycleEvent) (LifecycleState, error) {
	switch {
	case from == LifecycleOpen && event == LifecycleOnComplete:
		return LifecycleDone, nil
	case from == LifecycleDone && event == LifecycleOnReopen:
		return LifecycleOpen, nil
	case from == LifecycleDone && event == LifecycleOnArchive:
		return LifecycleArchived, nil
	case from == LifecycleArchived && event == LifecycleOnRestore:
		return LifecycleOpen, nil
	}
	return from, ErrLifecycleNoTransition
}
