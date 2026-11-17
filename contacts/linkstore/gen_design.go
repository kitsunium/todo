// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:32913ff1f1c2e4c8f3bc85573da565471d0c6a3bf2b70f14abb926ad227fee32

// Package linkstore — The links and the invitations in the framework's
// stores, their events on a topic, the directory asked of the identity, and
// the framework's clock.
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

// InviteRecord — An invitation as the store keeps it, field for field.
type InviteRecord struct {
	ID        string `json:"id" app:"public"`
	InviterID string `json:"inviterId" app:"public"`
	Email     string `json:"email" app:"personal,subject"`
	Status    string `json:"status" app:"public"`
	// Slot is the place a waiting invitation holds, its inviter and its address,
	// which no other invitation may hold; one that joined holds its own id.
	Slot      string    `json:"slot" app:"personal"`
	CreatedAt time.Time `json:"createdAt" app:"public"`
	UpdatedAt time.Time `json:"updatedAt" app:"public"`
}

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

// Courier is Inviter@v1; it writes to an address nobody has the invitation to
// sign up, in the language of the member who asks, through the product's
// mailer.
type Courier struct {
	courier // its own state, hand-written
}

// NewCourier builds Courier: its body is the hand-written newCourier.
func NewCourier() *Courier {
	return newCourier()
}

// Invite is contacts/port/Inviter@v1's.
func (a *Courier) Invite(ctx context.Context, to string, by circle.Person) error {
	return a.invite(ctx, to, by)
}

var _ circle.InviterV1 = (*Courier)(nil)

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

// Invitations is Invites@v1 on the store "invites".
type Invitations struct {
	invitations // its own state, hand-written
}

// NewInvitations builds Invitations: its body is the hand-written newInvitations.
func NewInvitations() *Invitations {
	return newInvitations()
}

// ByEmail is contacts/port/Invites@v1's.
func (self *Invitations) ByEmail(ctx context.Context, email string) ([]link.Invite, error) {
	return self.byEmail(ctx, email)
}

// Change is contacts/port/Invites@v1's.
func (self *Invitations) Change(ctx context.Context, id string, change func(*link.Invite) error) (link.Invite, error) {
	return self.change(ctx, id, change)
}

// Delete is contacts/port/Invites@v1's.
func (self *Invitations) Delete(ctx context.Context, id string) error {
	return self.deleteBody(ctx, id)
}

// Get is contacts/port/Invites@v1's.
func (self *Invitations) Get(ctx context.Context, id string) (link.Invite, bool, error) {
	return self.get(ctx, id)
}

// NewID is contacts/port/Invites@v1's.
func (self *Invitations) NewID(ctx context.Context) (string, error) {
	return self.newID(ctx)
}

// Of is contacts/port/Invites@v1's.
func (self *Invitations) Of(ctx context.Context, user string) ([]link.Invite, error) {
	return self.of(ctx, user)
}

// Save is contacts/port/Invites@v1's.
func (self *Invitations) Save(ctx context.Context, i link.Invite) error {
	return self.save(ctx, i)
}

var _ circle.InvitesV1 = (*Invitations)(nil)

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

// Change is contacts/port/Links@v1's.
func (self *Links) Change(ctx context.Context, id string, change func(*link.Link) error) (link.Link, error) {
	return self.change(ctx, id, change)
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
var Service = app.NewService("linkstore", "The links and the invitations in the framework's stores, their events on a topic, the directory asked of the identity, and the framework's clock.")

// LinksStore — Every link, keyed by its id; one a pair at most.
var LinksStore = Service.Store("links", func(e Record) string { return e.ID },
	app.Unique[Record]("by-pair", func(e Record) string { return e.Pair }),
)

// InvitesStore — Every invitation to an address with no account, keyed by
// its id; one is deleted ninety days after it was sent.
var InvitesStore = Service.Store("invites", func(e InviteRecord) string { return e.ID },
	app.DeleteAfter[InviteRecord](time.Duration(7776000000000000) /* 2160h */, func(e InviteRecord) (time.Time, bool) { return e.CreatedAt, !e.CreatedAt.IsZero() }),
	app.Index[InviteRecord]("by-email", func(e InviteRecord) []string { return []string{e.Email} }),
	app.Unique[InviteRecord]("by-slot", func(e InviteRecord) string { return e.Slot }),
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
