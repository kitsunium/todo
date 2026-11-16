// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:20b78385bcc831b04256ba24aa92ec0b4aec39dda784dd51beea005fcb21ccec

// Package linkstore — The links in the framework's store, their events on a
// topic, the directory asked of the identity, and the framework's clock.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package linkstore

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/contacts/circle"
	"github.com/kitsunium/todo/contacts/link"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 8, 3, 255), "NOT_IMPLEMENTED", "not implemented yet", "linkstore: a hand-written body is still a skeleton")

// Record — A link as the store keeps it, field for field.
type Record struct {
	ID          string       `json:"id" app:"public"`
	Pair        string       `json:"pair" app:"public"`
	RequesterID string       `json:"requesterId" app:"public"`
	AddresseeID string       `json:"addresseeId" app:"public"`
	State       WaitingState `json:"state" app:"public"`
	CreatedAt   time.Time    `json:"createdAt" app:"public"`
	UpdatedAt   time.Time    `json:"updatedAt" app:"public"`
}

// Event — What happened to a link, as the topic carries it to whoever
// listens.
type Event struct {
	// ID is the event's own: a listener that is delivered it twice keys what it
	// writes by it.
	ID string `json:"id" app:"public"`
	// Kind is contact-requested or contact-accepted.
	Kind   string `json:"kind" app:"public"`
	LinkID string `json:"linkId" app:"public"`
	// UserID is the member the event is told to.
	UserID    string    `json:"userId" app:"public"`
	ActorID   string    `json:"actorId" app:"public"`
	ActorName string    `json:"actorName" app:"personal"`
	At        time.Time `json:"at" app:"public"`
}

// Announcer is Notices@v1 on the topic "events".
type Announcer struct {
	announcer // its own state, hand-written
}

// NewAnnouncer builds Announcer: its body is the hand-written newAnnouncer.
func NewAnnouncer() *Announcer {
	return newAnnouncer()
}

// Tell is contacts/port/Notices@v1's.
func (a *Announcer) Tell(ctx context.Context, kind string, l link.Link, actor circle.Person) error {
	return a.tell(ctx, kind, l, actor)
}

var _ circle.NoticesV1 = (*Announcer)(nil)

// Directory is People@v1 on the identity's queries "member" and
// "member-by-email".
type Directory struct {
	directory // its own state, hand-written
}

// NewDirectory builds Directory: its body is the hand-written newDirectory.
func NewDirectory() *Directory {
	return newDirectory()
}

// ByEmail is contacts/port/People@v1's.
func (a *Directory) ByEmail(ctx context.Context, email string) (circle.Person, bool, error) {
	return a.byEmail(ctx, email)
}

// ByID is contacts/port/People@v1's.
func (a *Directory) ByID(ctx context.Context, id string) (circle.Person, bool, error) {
	return a.byID(ctx, id)
}

var _ circle.PeopleV1 = (*Directory)(nil)

// Links is Links@v1 on the store "links".
type Links struct {
	links // its own state, hand-written
}

// NewLinks builds Links: its body is the hand-written newLinks.
func NewLinks() *Links {
	return newLinks()
}

// ByPair is contacts/port/Links@v1's.
func (self *Links) ByPair(ctx context.Context, pair string) (link.Link, bool, error) {
	return self.byPair(ctx, pair)
}

// Delete is contacts/port/Links@v1's.
func (self *Links) Delete(ctx context.Context, id string) error {
	return self.deleteBody(ctx, id)
}

// Get is contacts/port/Links@v1's.
func (self *Links) Get(ctx context.Context, id string) (link.Link, bool, error) {
	return self.get(ctx, id)
}

// NewID is contacts/port/Links@v1's.
func (self *Links) NewID(ctx context.Context) (string, error) {
	return self.newID(ctx)
}

// Of is contacts/port/Links@v1's.
func (self *Links) Of(ctx context.Context, user string) ([]link.Link, error) {
	return self.of(ctx, user)
}

// Save is contacts/port/Links@v1's.
func (self *Links) Save(ctx context.Context, l link.Link) error {
	return self.save(ctx, l)
}

var _ circle.LinksV1 = (*Links)(nil)

// WallClock is Clock@v1 on the framework's clock, UTC, to the millisecond.
type WallClock struct {
	wallClock // its own state, hand-written
}

// NewWallClock builds WallClock: its body is the hand-written newWallClock.
func NewWallClock() *WallClock {
	return newWallClock()
}

// Now is contacts/port/Clock@v1's.
func (a *WallClock) Now(ctx context.Context) (time.Time, error) {
	return a.now(ctx)
}

var _ circle.ClockV1 = (*WallClock)(nil)

// WaitingState is a state of Waiting.
type WaitingState string

// The states of Waiting.
const (
	WaitingPending   WaitingState = "pending"
	WaitingAccepted  WaitingState = "accepted"
	WaitingDeclined  WaitingState = "declined"
	WaitingCancelled WaitingState = "cancelled"
	WaitingExpired   WaitingState = "expired"
)

// WaitingInitial is the state a Record starts in.
const WaitingInitial = WaitingPending

// The events of Waiting, as WaitingWorkflow.Fire takes them.
const (
	WaitingOnExpire = "expire"
)

// WaitingWorkflow — What time does to a link kept, a request nobody answered
// expires after fourteen days; its answers are the use cases', which write the
// state they decided.
var WaitingWorkflow = Service.Workflow("waiting", LinksStore, func(e *Record) *WaitingState { return &e.State }).
	Initial(WaitingPending).
	After(WaitingOnExpire, time.Duration(1209600000000000) /* 336h */, WaitingPending, WaitingExpired)

// Service is linkstore's framework service: what it declares, drawn and run.
var Service = app.NewService("linkstore", "The links in the framework's store, their events on a topic, the directory asked of the identity, and the framework's clock.")

// LinksStore — Every link, keyed by its id; one a pair at most.
var LinksStore = Service.Store("links", func(e Record) string { return e.ID },
	app.Unique[Record]("by-pair", func(e Record) string { return e.Pair }),
)

// EventsTopic — What happened to the links: a request received, a request
// accepted, delivered at least once.
var EventsTopic = Service.Topic[Event]("events")

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts linkstore calls it
// before its App runs.
func Wire() {

}
