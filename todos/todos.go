// Package todos is the heart of the product: the todo list, and the lifecycle
// every todo goes through.
package todos

import (
	"context"
	"time"

	"github.com/kitsunium/platform/kit"
)

// Service owns the todo list.
var Service = kit.NewService("todos", "The todo list, and the lifecycle every todo goes through.")

// Status is where a todo stands in its lifecycle.
type Status string

// The lifecycle states.
const (
	Open     Status = "open"
	Overdue  Status = "overdue"
	Done     Status = "done"
	Archived Status = "archived"
)

// Todo is one thing to do.
type Todo struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Status    Status     `json:"status"`
	Due       *time.Time `json:"due,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	DoneAt    *time.Time `json:"doneAt,omitempty"`
}

// Todos keeps every todo, keyed by ID.
var Todos = Service.Store("todos", func(t Todo) string { return t.ID })

// Change is published every time a todo changes state. Its ID identifies the
// change itself, so a subscriber can recognize a redelivery.
type Change struct {
	ID     string    `json:"id"`
	TodoID string    `json:"todoId"`
	Title  string    `json:"title"`
	Event  string    `json:"event"`
	From   Status    `json:"from,omitempty"`
	To     Status    `json:"to"`
	At     time.Time `json:"at"`
}

// Changes carries every lifecycle change to whoever listens.
var Changes = Service.Topic[Change]("changes")

// ArchiveAfter is how long a done todo stays on the list before it is archived.
var ArchiveAfter = kit.EnvDuration("TODO_ARCHIVE_AFTER", 2*time.Minute)

// Lifecycle is the state machine every todo lives in. Endpoints fire its
// events; its own loop fires the timed and guarded transitions.
var Lifecycle = Service.Workflow("lifecycle", Todos, func(t *Todo) *Status { return &t.Status }).
	Initial(Open).
	On("complete", Open, Done).
	On("complete", Overdue, Done).
	On("reopen", Done, Open).
	On("archive", Done, Archived).
	When("overdue", Open, Overdue, isOverdue).
	After("auto-archive", ArchiveAfter, Done, Archived).
	OnEnter(Done, stampDone).
	OnEnter(Open, clearDone).
	OnTransition(announce)

// isOverdue holds when a todo's due date has passed.
func isOverdue(t Todo, now time.Time) bool {
	return t.Due != nil && now.After(*t.Due)
}

// stampDone records when a todo was completed.
func stampDone(_ context.Context, t *Todo) error {
	now := time.Now().UTC()
	t.DoneAt = &now
	return nil
}

// clearDone forgets the completion of a reopened todo.
func clearDone(_ context.Context, t *Todo) error {
	t.DoneAt = nil
	return nil
}

// announce publishes every transition on Changes, whoever fired it: an
// endpoint, the auto-archive timer or the overdue guard.
func announce(ctx context.Context, c kit.Change[Todo, Status]) error {
	return Changes.Publish(ctx, Change{
		ID:     kit.NewID("change"),
		TodoID: c.Key,
		Title:  c.Entity.Title,
		Event:  c.Event,
		From:   c.From,
		To:     c.To,
		At:     c.At,
	})
}
