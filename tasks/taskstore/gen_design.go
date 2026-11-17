// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:c09190927b7a702a68883efedfd3b487bc8023ea307d3b6eeda776c8d6820e96

// Package taskstore — The tasks in the framework's store, the archiving of
// what was done long ago, their events on a topic, the directory asked of the
// identity, and the framework's clock.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package taskstore

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/groups/groupstore"
	"github.com/kitsunium/todo/tasks/planner"
	"github.com/kitsunium/todo/tasks/task"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 1, 3, 255), "NOT_IMPLEMENTED", "not implemented yet", "taskstore: a hand-written body is still a skeleton")

// Record — A task as the store keeps it, field for field.
type Record struct {
	ID         string     `json:"id" app:"public"`
	OwnerID    string     `json:"ownerId" app:"public"`
	SharedWith []string   `json:"sharedWith" app:"public"`
	Title      string     `json:"title" app:"public"`
	Notes      string     `json:"notes" app:"public"`
	Priority   string     `json:"priority" app:"public"`
	GroupID    string     `json:"groupId" app:"public"`
	AssigneeID string     `json:"assigneeId" app:"personal,subject"`
	Due        time.Time  `json:"due" app:"public"`
	State      ShelfState `json:"state" app:"public"`
	CreatedAt  time.Time  `json:"createdAt" app:"public"`
	UpdatedAt  time.Time  `json:"updatedAt" app:"public"`
	DoneAt     time.Time  `json:"doneAt" app:"public"`
}

// Event — What happened to a task, as the topic carries it to whoever
// listens.
type Event struct {
	// ID is the event's own: a listener that is delivered it twice keys what it
	// writes by it.
	ID string `json:"id" app:"public"`
	// Kind is added, or the lifecycle event that moved the task.
	Kind    string `json:"kind" app:"public"`
	TaskID  string `json:"taskId" app:"public"`
	OwnerID string `json:"ownerId" app:"public"`
	// Audience are the members the event is told to: the task's owner, those it is
	// shared with, its assignee and the members of its group.
	Audience []string  `json:"audience" app:"public"`
	Title    string    `json:"title" app:"public"`
	At       time.Time `json:"at" app:"public"`
	// ActorID is the member who did it, for an event that is about one member more
	// than the task: shared and assigned. Empty otherwise.
	ActorID string `json:"actorId,omitempty" app:"public"`
	// UserID is the member the task was shared with or assigned to. Empty
	// otherwise.
	UserID string `json:"userId,omitempty" app:"public"`
}

// Announcer is Journal@v1 on the topic "events".
type Announcer struct {
	announcer // its own state, hand-written
}

// NewAnnouncer builds Announcer: its body is the hand-written newAnnouncer.
func NewAnnouncer() *Announcer {
	return newAnnouncer()
}

// Involve is tasks/port/Journal@v1's.
func (a *Announcer) Involve(ctx context.Context, kind string, t task.Task, actor string, user string) error {
	return a.involve(ctx, kind, t, actor, user)
}

// Record is tasks/port/Journal@v1's.
func (a *Announcer) Record(ctx context.Context, kind string, t task.Task) error {
	return a.record(ctx, kind, t)
}

// Tell is tasks/port/Journal@v1's.
func (a *Announcer) Tell(ctx context.Context, kind string, t task.Task, actor string, to string) error {
	return a.tell(ctx, kind, t, actor, to)
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

// Directory is People@v1 on the identity's query "member".
type Directory struct {
	directory // its own state, hand-written
}

// NewDirectory builds Directory: its body is the hand-written newDirectory.
func NewDirectory() *Directory {
	return newDirectory()
}

// ByID is tasks/port/People@v1's.
func (a *Directory) ByID(ctx context.Context, id string) (planner.Person, bool, error) {
	return a.byID(ctx, id)
}

var _ planner.PeopleV1 = (*Directory)(nil)

// Friends is Circle@v1 on the contacts' query "are-contacts".
type Friends struct {
	friends // its own state, hand-written
}

// NewFriends builds Friends: its body is the hand-written newFriends.
func NewFriends() *Friends {
	return newFriends()
}

// AreContacts is tasks/port/Circle@v1's.
func (a *Friends) AreContacts(ctx context.Context, one string, other string) (bool, error) {
	return a.areContacts(ctx, one, other)
}

var _ planner.CircleV1 = (*Friends)(nil)

// Groups is Groups@v1 on the groups' queries "role" and "memberships".
type Groups struct {
	groups // its own state, hand-written
}

// NewGroups builds Groups: its body is the hand-written newGroups.
func NewGroups() *Groups {
	return newGroups()
}

// Memberships is tasks/port/Groups@v1's.
func (a *Groups) Memberships(ctx context.Context, user string) (map[string]string, error) {
	return a.memberships(ctx, user)
}

// RoleOf is tasks/port/Groups@v1's.
func (a *Groups) RoleOf(ctx context.Context, group string, user string) (string, error) {
	return a.roleOf(ctx, group, user)
}

var _ planner.GroupsV1 = (*Groups)(nil)

// Tasks is TaskStore@v1 on the store "tasks".
type Tasks struct {
	tasks // its own state, hand-written
}

// NewTasks builds Tasks: its body is the hand-written newTasks.
func NewTasks() *Tasks {
	return newTasks()
}

// Change is tasks/port/TaskStore@v1's.
func (self *Tasks) Change(ctx context.Context, id string, change func(*task.Task) error) (task.Task, error) {
	return self.change(ctx, id, change)
}

// Delete is tasks/port/TaskStore@v1's.
func (self *Tasks) Delete(ctx context.Context, id string) error {
	return self.deleteBody(ctx, id)
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

// ShelfState is a state of Shelf.
type ShelfState string

// The states of Shelf.
const (
	ShelfOpen     ShelfState = "open"
	ShelfDone     ShelfState = "done"
	ShelfArchived ShelfState = "archived"
)

// ShelfInitial is the state a Record starts in.
const ShelfInitial = ShelfOpen

// The events of Shelf, as ShelfWorkflow.Fire takes them.
const (
	ShelfOnAutoArchive = "autoArchive"
)

// ShelfWorkflow — What time does to a task kept, a done one is archived once
// archive-after has passed; its moves by a member are the use cases', which
// write the state they decided.
var ShelfWorkflow = Service.Workflow("shelf", TasksStore, func(e *Record) *ShelfState { return &e.State }).
	Initial(ShelfOpen).
	After(ShelfOnAutoArchive, ArchiveAfterSetting, ShelfDone, ShelfArchived).
	OnTransition(onShelfTransition)

// Service is taskstore's framework service: what it declares, drawn and run.
var Service = app.NewService("taskstore", "The tasks in the framework's store, the archiving of what was done long ago, their events on a topic, the directory asked of the identity, and the framework's clock.")

// ArchiveAfterSetting — How long a done task stays listed before it is
// archived.
var ArchiveAfterSetting = Service.Setting[time.Duration]("archive-after", 24*time.Hour)

// TasksStore — Every task, keyed by its id; who it is assigned to is erased
// a year after it was last written.
var TasksStore = Service.Store("tasks", func(e Record) string { return e.ID },
	app.EraseAfter[Record](time.Duration(31536000000000000) /* 8760h */, func(e Record) (time.Time, bool) { return e.UpdatedAt, !e.UpdatedAt.IsZero() }),
)

// EventsTopic — What happened to the tasks: one message a task added or
// moved, delivered at least once.
var EventsTopic = Service.Topic[Event]("events")

var _ = Service.Subscribe("groups", groupstore.EventsTopic, FollowGroups)

// FollowGroups — A group deleted, or a member who left it or was removed
// from it: the tasks follow.
func FollowGroups(ctx context.Context, msg groupstore.Event) error {
	return followGroups(ctx, msg)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
	followGroup planner.FollowGroup
}

// Wire gives the handlers what they use: the role that mounts taskstore calls it
// before its App runs.
func Wire(followGroup planner.FollowGroup) {
	wired.followGroup = followGroup
}
