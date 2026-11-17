// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:ad3708a6c0120e4ecb328f05fe017711e138b2d9494fff3cf0c4cd1c6596edae

// Package groupstore — The groups and their invitations in the framework's
// store, what happens in a group on a topic, the directory asked of the
// identity, the contacts asked of theirs, and the framework's clock.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package groupstore

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/groups/roster"
	"github.com/kitsunium/todo/groups/team"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 9, 3, 255), "NOT_IMPLEMENTED", "not implemented yet", "groupstore: a hand-written body is still a skeleton")

// GroupRecord — A group as the store keeps it, field for field.
type GroupRecord struct {
	ID        string            `json:"id" app:"public"`
	Name      string            `json:"name" app:"public"`
	Color     string            `json:"color" app:"public"`
	Members   []team.Membership `json:"members"`
	CreatedAt time.Time         `json:"createdAt" app:"public"`
	UpdatedAt time.Time         `json:"updatedAt" app:"public"`
}

// InvitationRecord — An invitation as the store keeps it, field for field.
type InvitationRecord struct {
	ID          string       `json:"id" app:"public"`
	GroupID     string       `json:"groupId" app:"public"`
	InviterID   string       `json:"inviterId" app:"public"`
	InviteeID   string       `json:"inviteeId" app:"public"`
	State       LapsingState `json:"state" app:"public"`
	CreatedAt   time.Time    `json:"createdAt" app:"public"`
	RespondedAt time.Time    `json:"respondedAt" app:"public"`
}

// Event — What happened in a group, as the topic carries it to whoever
// listens, one a member it is told to.
type Event struct {
	// ID is the event's own: a listener that is delivered it twice keys what it
	// writes by it.
	ID string `json:"id" app:"public"`
	// Kind is group-invited, group-joined, group-left, group-removed, group-role
	// or group-deleted.
	Kind      string `json:"kind" app:"public"`
	GroupID   string `json:"groupId" app:"public"`
	GroupName string `json:"groupName" app:"public"`
	// UserID is the member the event is told to.
	UserID    string    `json:"userId" app:"public"`
	ActorID   string    `json:"actorId" app:"public"`
	ActorName string    `json:"actorName" app:"personal"`
	At        time.Time `json:"at" app:"public"`
}

// Announcer is Notices@v1 on the topic events, one event for each member told,
// the actor never.
type Announcer struct {
	announcer // its own state, hand-written
}

// NewAnnouncer builds Announcer: its body is the hand-written newAnnouncer.
func NewAnnouncer() *Announcer {
	return newAnnouncer()
}

// Tell is groups/port/Notices@v1's.
func (a *Announcer) Tell(ctx context.Context, kind string, g team.Group, actor roster.Person, to []string) error {
	return a.tell(ctx, kind, g, actor, to)
}

var _ roster.NoticesV1 = (*Announcer)(nil)

// Directory is People@v1 on the identity's query "member".
type Directory struct {
	directory // its own state, hand-written
}

// NewDirectory builds Directory: its body is the hand-written newDirectory.
func NewDirectory() *Directory {
	return newDirectory()
}

// ByID is groups/port/People@v1's.
func (a *Directory) ByID(ctx context.Context, id string) (roster.Person, bool, error) {
	return a.byID(ctx, id)
}

var _ roster.PeopleV1 = (*Directory)(nil)

// Friends is Circle@v1 on the contacts' query "are-contacts".
type Friends struct {
	friends // its own state, hand-written
}

// NewFriends builds Friends: its body is the hand-written newFriends.
func NewFriends() *Friends {
	return newFriends()
}

// AreContacts is groups/port/Circle@v1's.
func (a *Friends) AreContacts(ctx context.Context, one string, other string) (bool, error) {
	return a.areContacts(ctx, one, other)
}

var _ roster.CircleV1 = (*Friends)(nil)

// Groups is Groups@v1 on the store "groups".
type Groups struct {
	groups // its own state, hand-written
}

// NewGroups builds Groups: its body is the hand-written newGroups.
func NewGroups() *Groups {
	return newGroups()
}

// Change is groups/port/Groups@v1's.
func (self *Groups) Change(ctx context.Context, id string, change func(*team.Group) error) (team.Group, error) {
	return self.change(ctx, id, change)
}

// Delete is groups/port/Groups@v1's.
func (self *Groups) Delete(ctx context.Context, id string) error {
	return self.deleteBody(ctx, id)
}

// Get is groups/port/Groups@v1's.
func (self *Groups) Get(ctx context.Context, id string) (team.Group, bool, error) {
	return self.get(ctx, id)
}

// NewID is groups/port/Groups@v1's.
func (self *Groups) NewID(ctx context.Context) (string, error) {
	return self.newID(ctx)
}

// Of is groups/port/Groups@v1's.
func (self *Groups) Of(ctx context.Context, user string) ([]team.Group, error) {
	return self.of(ctx, user)
}

// Save is groups/port/Groups@v1's.
func (self *Groups) Save(ctx context.Context, g team.Group) error {
	return self.save(ctx, g)
}

var _ roster.GroupsV1 = (*Groups)(nil)

// Invitations is Invitations@v1 on the store "invitations".
type Invitations struct {
	invitations // its own state, hand-written
}

// NewInvitations builds Invitations: its body is the hand-written newInvitations.
func NewInvitations() *Invitations {
	return newInvitations()
}

// Change is groups/port/Invitations@v1's.
func (a *Invitations) Change(ctx context.Context, id string, change func(*team.Invitation) error) (team.Invitation, error) {
	return a.change(ctx, id, change)
}

// Delete is groups/port/Invitations@v1's.
func (a *Invitations) Delete(ctx context.Context, id string) error {
	return a.deleteBody(ctx, id)
}

// Get is groups/port/Invitations@v1's.
func (a *Invitations) Get(ctx context.Context, id string) (team.Invitation, bool, error) {
	return a.get(ctx, id)
}

// NewID is groups/port/Invitations@v1's.
func (a *Invitations) NewID(ctx context.Context) (string, error) {
	return a.newID(ctx)
}

// Of is groups/port/Invitations@v1's.
func (a *Invitations) Of(ctx context.Context, user string) ([]team.Invitation, error) {
	return a.of(ctx, user)
}

// OfGroup is groups/port/Invitations@v1's.
func (a *Invitations) OfGroup(ctx context.Context, group string) ([]team.Invitation, error) {
	return a.ofGroup(ctx, group)
}

// Save is groups/port/Invitations@v1's.
func (a *Invitations) Save(ctx context.Context, inv team.Invitation) error {
	return a.save(ctx, inv)
}

var _ roster.InvitationsV1 = (*Invitations)(nil)

// WallClock is Clock@v1 on the framework's clock, UTC, to the millisecond.
type WallClock struct {
	wallClock // its own state, hand-written
}

// NewWallClock builds WallClock: its body is the hand-written newWallClock.
func NewWallClock() *WallClock {
	return newWallClock()
}

// Now is groups/port/Clock@v1's.
func (a *WallClock) Now(ctx context.Context) (time.Time, error) {
	return a.now(ctx)
}

var _ roster.ClockV1 = (*WallClock)(nil)

// LapsingState is a state of Lapsing.
type LapsingState string

// The states of Lapsing.
const (
	LapsingPending   LapsingState = "pending"
	LapsingAccepted  LapsingState = "accepted"
	LapsingDeclined  LapsingState = "declined"
	LapsingWithdrawn LapsingState = "withdrawn"
	LapsingExpired   LapsingState = "expired"
)

// LapsingInitial is the state a InvitationRecord starts in.
const LapsingInitial = LapsingPending

// The events of Lapsing, as LapsingWorkflow.Fire takes them.
const (
	LapsingOnExpire = "expire"
)

// LapsingWorkflow — What time does to an invitation kept, one nobody
// answered expires after a week; its answers are the use cases', which write
// the state they decided.
var LapsingWorkflow = Service.Workflow("lapsing", InvitationsStore, func(e *InvitationRecord) *LapsingState { return &e.State }).
	Initial(LapsingPending).
	After(LapsingOnExpire, time.Duration(604800000000000) /* 168h */, LapsingPending, LapsingExpired)

// Service is groupstore's framework service: what it declares, drawn and run.
var Service = app.NewService("groupstore", "The groups and their invitations in the framework's store, what happens in a group on a topic, the directory asked of the identity, the contacts asked of theirs, and the framework's clock.")

// GroupsStore — Every group, keyed by its id.
var GroupsStore = Service.Store("groups", func(e GroupRecord) string { return e.ID })

// InvitationsStore — Every invitation, keyed by its id, found by whom it
// invites and by the group it is into.
var InvitationsStore = Service.Store("invitations", func(e InvitationRecord) string { return e.ID },
	app.Index[InvitationRecord]("by-invitee", func(e InvitationRecord) []string { return []string{e.InviteeID} }),
	app.Index[InvitationRecord]("by-group", func(e InvitationRecord) []string { return []string{e.GroupID} }),
)

// EventsTopic — What happened in the groups: an invitation, a member who
// joined, left, was removed or was given a role, a group deleted, delivered at
// least once.
var EventsTopic = Service.Topic[Event]("events")

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts groupstore calls it
// before its App runs.
func Wire() {

}
