// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:390717aeedfce4c9881bfea93cf567e5dfb6e0f4b99777661b165bb23d396f17

// Package team — A group, its members and their roles, and the invitation
// that brings a member in.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package team

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/pkg/v1/errs"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 9, 1, 255), "NOT_IMPLEMENTED", "not implemented yet", "team: a hand-written body is still a skeleton")

var (
	// ErrAlreadyInvited is "this person is already invited".
	ErrAlreadyInvited = errs.New(errs.Pack(64, 9, 1, 10), "ALREADY_INVITED", "this person is already invited", "groups/error/ErrAlreadyInvited")
	// ErrAlreadyMember is "this person is already a member".
	ErrAlreadyMember = errs.New(errs.Pack(64, 9, 1, 9), "ALREADY_MEMBER", "this person is already a member", "groups/error/ErrAlreadyMember")
	// ErrBadColor is "the color is not one of the group colors".
	ErrBadColor = errs.New(errs.Pack(64, 9, 1, 14), "BAD_COLOR", "the color is not one of the group colors", "groups/error/ErrBadColor")
	// ErrBadName is "a name is one line of one to 60 characters".
	ErrBadName = errs.New(errs.Pack(64, 9, 1, 13), "BAD_NAME", "a name is one line of one to 60 characters", "groups/error/ErrBadName")
	// ErrBadRole is "a role is admin or member".
	ErrBadRole = errs.New(errs.Pack(64, 9, 1, 15), "BAD_ROLE", "a role is admin or member", "groups/error/ErrBadRole")
	// ErrCannotRemove is "you cannot remove this member".
	ErrCannotRemove = errs.New(errs.Pack(64, 9, 1, 7), "CANNOT_REMOVE", "you cannot remove this member", "groups/error/ErrCannotRemove")
	// ErrNoInvitation is "no such invitation".
	ErrNoInvitation = errs.New(errs.Pack(64, 9, 1, 2), "NO_INVITATION", "no such invitation", "groups/error/ErrNoInvitation")
	// ErrNoMember is "no such member".
	ErrNoMember = errs.New(errs.Pack(64, 9, 1, 3), "NO_MEMBER", "no such member", "groups/error/ErrNoMember")
	// ErrNotContact is "you can only invite your contacts".
	ErrNotContact = errs.New(errs.Pack(64, 9, 1, 6), "NOT_CONTACT", "you can only invite your contacts", "groups/error/ErrNotContact")
	// ErrNotFound is "no such group".
	ErrNotFound = errs.New(errs.Pack(64, 9, 1, 1), "NOT_FOUND", "no such group", "groups/error/ErrNotFound")
	// ErrNotManager is "only the group's owner and admins can do this".
	ErrNotManager = errs.New(errs.Pack(64, 9, 1, 4), "NOT_MANAGER", "only the group's owner and admins can do this", "groups/error/ErrNotManager")
	// ErrNotOwner is "only the group's owner can do this".
	ErrNotOwner = errs.New(errs.Pack(64, 9, 1, 5), "NOT_OWNER", "only the group's owner can do this", "groups/error/ErrNotOwner")
	// ErrNotPending is "this invitation is not waiting for an answer any more".
	ErrNotPending = errs.New(errs.Pack(64, 9, 1, 12), "NOT_PENDING", "this invitation is not waiting for an answer any more", "groups/error/ErrNotPending")
	// ErrOwnerRole is "the owner's role cannot change".
	ErrOwnerRole = errs.New(errs.Pack(64, 9, 1, 8), "OWNER_ROLE", "the owner's role cannot change", "groups/error/ErrOwnerRole")
	// ErrOwnerStays is "the owner cannot leave the group: delete it instead".
	ErrOwnerStays = errs.New(errs.Pack(64, 9, 1, 11), "OWNER_STAYS", "the owner cannot leave the group: delete it instead", "groups/error/ErrOwnerStays")
)

const (
	// MaxName is the longest name a group takes.
	MaxName int = 60
	// RoleAdmin invites and removes members and edits the group.
	RoleAdmin Role = "admin"
	// RoleMember is a member who works on the group's tasks.
	RoleMember Role = "member"
	// RoleOwner created the group and alone deletes it or changes a role.
	RoleOwner Role = "owner"
)

// Group — A team, its members and the role each one has in it.
type Group struct {
	ID        string       `json:"id" app:"public"`
	Name      string       `json:"name" app:"public"`
	Color     string       `json:"color" app:"public"`
	Members   []Membership `json:"members"`
	CreatedAt time.Time    `json:"createdAt" app:"public"`
	UpdatedAt time.Time    `json:"updatedAt" app:"public"`
}

// MemberIDs are the ids of the members, in the order the group keeps them.
func (g *Group) MemberIDs() []string {
	return g.memberIDs()
}

// RoleOf is the role a member has in the group, empty for a stranger.
func (g *Group) RoleOf(user string) Role {
	return g.roleOf(user)
}

// Invitation — An invitation of a contact into a group, and what became of
// it.
type Invitation struct {
	ID        string        `json:"id" app:"public"`
	GroupID   string        `json:"groupId" app:"public"`
	InviterID string        `json:"inviterId" app:"public"`
	InviteeID string        `json:"inviteeId" app:"public"`
	State     InvitingState `json:"state" app:"public"`
	CreatedAt time.Time     `json:"createdAt" app:"public"`
	// RespondedAt is when the invitee answered; the zero time while nobody has.
	RespondedAt time.Time `json:"respondedAt" app:"public"`
}

// Membership — One member of a group.
type Membership struct {
	UserID   string    `json:"userId" app:"public"`
	Role     Role      `json:"role" app:"public"`
	JoinedAt time.Time `json:"joinedAt" app:"public"`
}

// Role — What a member may do in a group.
type Role string

// CleanName reads a group's name as a member writes it, trimmed, one line of
// one to MaxName characters.
func CleanName(name string) (string, error) {
	return cleanName(name)
}

// ColorFor checks a group's color, or picks one from its name when none is
// sent.
func ColorFor(color string, name string) (string, error) {
	return colorFor(color, name)
}

// Manages says whether a role runs the group: invites, removes, edits.
func Manages(role Role) bool {
	return manages(role)
}

// ParseRole reads the role a member is given, admin or member; the owner is
// never given.
func ParseRole(s string) (Role, error) {
	return parseRole(s)
}

// Ranked is the members, the owner first, then the admins, then the members,
// each rank by when they joined.
func Ranked(members []Membership) []Membership {
	return ranked(members)
}

// InvitingState is a state of Inviting.
type InvitingState string

// The states of Inviting.
const (
	InvitingPending   InvitingState = "pending"
	InvitingAccepted  InvitingState = "accepted"
	InvitingDeclined  InvitingState = "declined"
	InvitingWithdrawn InvitingState = "withdrawn"
	InvitingExpired   InvitingState = "expired"
)

// InvitingInitial is the state a Invitation starts in.
const InvitingInitial = InvitingPending

// InvitingEvent is an event of Inviting.
type InvitingEvent string

// The events of Inviting.
const (
	InvitingOnAccept   InvitingEvent = "accept"
	InvitingOnDecline  InvitingEvent = "decline"
	InvitingOnWithdraw InvitingEvent = "withdraw"
)

var (
	// ErrInvitingRefused is a transition its guard refused.
	ErrInvitingRefused = errs.New(errs.Pack(64, 9, 1, 240), "INVITING_REFUSED", "the transition is refused", "groups/workflow/Inviting")
	// ErrInvitingNoTransition is an event its state does not answer.
	ErrInvitingNoTransition = errs.New(errs.Pack(64, 9, 1, 241), "INVITING_NO_TRANSITION", "no transition for this event in this state", "groups/workflow/Inviting")
)

// Inviting — An invitation is pending until its invitee accepts or declines
// it, or a manager of the group withdraws it; one nobody answered for a week
// has expired.
type Inviting struct {
	inviting // its own state, hand-written
}

// NewInviting builds Inviting: its body is the hand-written newInviting.
func NewInviting() *Inviting {
	return newInviting()
}

// Fire moves e from state from by event: the state it enters, or the state
// it stays in and why.
func (w *Inviting) Fire(ctx context.Context, e *Invitation, from InvitingState, event InvitingEvent) (InvitingState, error) {
	switch {
	case from == InvitingPending && event == InvitingOnAccept:
		return InvitingAccepted, nil
	case from == InvitingPending && event == InvitingOnDecline:
		return InvitingDeclined, nil
	case from == InvitingPending && event == InvitingOnWithdraw:
		return InvitingWithdrawn, nil
	}
	return from, ErrInvitingNoTransition
}
