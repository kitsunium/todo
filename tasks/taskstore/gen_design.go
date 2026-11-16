// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:35933b95eb37968572e6d46dc1379b8281e3503fec0e8451a525bac3fb04f4af

// Package taskstore — The tasks in the framework's store, their events on a
// topic, and the framework's clock.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package taskstore

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/tasks/planner"
	"github.com/kitsunium/todo/tasks/task"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 1, 3, 255), "NOT_IMPLEMENTED", "not implemented yet", "taskstore: a hand-written body is still a skeleton")

// Record — A task as the store keeps it, field for field.
type Record struct {
	ID        string    `json:"id" app:"public"`
	OwnerID   string    `json:"ownerId" app:"public"`
	Title     string    `json:"title" app:"public"`
	Notes     string    `json:"notes" app:"public"`
	Priority  string    `json:"priority" app:"public"`
	Assignee  string    `json:"assignee" app:"personal,subject"`
	Due       time.Time `json:"due" app:"public"`
	State     string    `json:"state" app:"public"`
	CreatedAt time.Time `json:"createdAt" app:"public"`
	UpdatedAt time.Time `json:"updatedAt" app:"public"`
	DoneAt    time.Time `json:"doneAt" app:"public"`
}

// Event — What happened to a task, as the topic carries it to whoever
// listens.
type Event struct {
	// ID is the event's own: a listener that is delivered it twice keys what it
	// writes by it.
	ID string `json:"id" app:"public"`
	// Kind is added, or the lifecycle event that moved the task.
	Kind     string    `json:"kind" app:"public"`
	TaskID   string    `json:"taskId" app:"public"`
	OwnerID  string    `json:"ownerId" app:"public"`
	Title    string    `json:"title" app:"public"`
	Assignee string    `json:"assignee" app:"personal"`
	At       time.Time `json:"at" app:"public"`
}

// Announcer is Journal@v1 on the topic "events".
type Announcer struct {
	announcer // its own state, hand-written
}

// NewAnnouncer builds Announcer: its body is the hand-written newAnnouncer.
func NewAnnouncer() *Announcer {
	return newAnnouncer()
}

// Record is tasks/port/Journal@v1's.
func (a *Announcer) Record(ctx context.Context, kind string, t task.Task) error {
	return a.record(ctx, kind, t)
}

var _ planner.JournalV1 = (*Announcer)(nil)

// AppClock is Clock@v1 on the framework's clock, UTC, to the millisecond.
type AppClock struct {
	appClock // its own state, hand-written
}

// NewAppClock builds AppClock: its body is the hand-written newAppClock.
func NewAppClock() *AppClock {
	return newAppClock()
}

// Now is tasks/port/Clock@v1's.
func (a *AppClock) Now(ctx context.Context) (time.Time, error) {
	return a.now(ctx)
}

var _ planner.ClockV1 = (*AppClock)(nil)

// Tasks is TaskStore@v1 on the store "tasks".
type Tasks struct {
	tasks // its own state, hand-written
}

// NewTasks builds Tasks: its body is the hand-written newTasks.
func NewTasks() *Tasks {
	return newTasks()
}

// Get is tasks/port/TaskStore@v1's.
func (self *Tasks) Get(ctx context.Context, id string) (task.Task, error) {
	return self.get(ctx, id)
}

// List is tasks/port/TaskStore@v1's.
func (self *Tasks) List(ctx context.Context) ([]task.Task, error) {
	return self.list(ctx)
}

// NewID is tasks/port/TaskStore@v1's.
func (self *Tasks) NewID(ctx context.Context) (string, error) {
	return self.newID(ctx)
}

// Save is tasks/port/TaskStore@v1's.
func (self *Tasks) Save(ctx context.Context, t task.Task) error {
	return self.save(ctx, t)
}

var _ planner.TaskStoreV1 = (*Tasks)(nil)

// Service is taskstore's framework service: what it declares, drawn and run.
var Service = app.NewService("taskstore", "The tasks in the framework's store, their events on a topic, and the framework's clock.")

// TasksStore — Every task, keyed by its id; who it is assigned to is erased
// a year after it was last written.
var TasksStore = Service.Store("tasks", func(e Record) string { return e.ID },
	app.EraseAfter[Record](time.Duration(31536000000000000) /* 8760h */, func(e Record) (time.Time, bool) { return e.UpdatedAt, !e.UpdatedAt.IsZero() }),
)

// EventsTopic — What happened to the tasks: one message a task added or
// moved, delivered at least once.
var EventsTopic = Service.Topic[Event]("events")

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts taskstore calls it
// before its App runs.
func Wire() {

}
