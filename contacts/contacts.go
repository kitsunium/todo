// Package contacts is who a user works with. Two users become contacts when
// one asks and the other accepts; someone without an account is invited by
// email, and the invitation becomes a request once they sign up. Sharing a
// task and inviting someone into a group both need a contact.
package contacts

import (
	"context"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/internal/wire"
)

// Service owns contacts, contact requests and invitations.
var Service = kit.NewService("contacts", "Who you work with: contact requests between users, and invitations for people who have no account yet.")

// LinkStatus is where a contact request stands.
type LinkStatus string

// The states of a contact request.
const (
	Pending   LinkStatus = "pending"
	Accepted  LinkStatus = "accepted"
	Declined  LinkStatus = "declined"
	Cancelled LinkStatus = "cancelled"
	Expired   LinkStatus = "expired"
)

// Link is what binds two users: a request, and once accepted, a contact.
// There is one per pair of users, whoever asked.
type Link struct {
	ID string `json:"id"`
	// Pair is the two user IDs, sorted: the store key, which makes a second
	// link between the same two users impossible.
	Pair        string     `json:"pair"`
	RequesterID string     `json:"requesterId"`
	AddresseeID string     `json:"addresseeId"`
	Status      LinkStatus `json:"status"`
	CreatedAt   time.Time  `json:"createdAt"`
	AcceptedAt  *time.Time `json:"acceptedAt,omitempty"`
}

// other returns the user of the link who is not uid.
func (l Link) other(uid string) string {
	if l.RequesterID == uid {
		return l.AddresseeID
	}
	return l.RequesterID
}

// pairOf is the store key of the link between two users.
func pairOf(a, b string) string {
	if a > b {
		a, b = b, a
	}
	return a + ":" + b
}

// Links keeps one link per pair of users, found by ID and listed per user.
var Links = Service.Store("links", func(l Link) string { return l.Pair },
	kit.Unique("id", func(l Link) string { return l.ID }),
	kit.Index("user", func(l Link) []string { return []string{l.RequesterID, l.AddresseeID} }))

// RequestsExpireAfter is how long a request waits for an answer.
const RequestsExpireAfter = 14 * 24 * time.Hour

// Requests is the life of a contact request: the addressee accepts or
// declines it, the requester cancels it, or it expires unanswered.
var Requests = Service.Workflow("requests", Links, func(l *Link) *LinkStatus { return &l.Status }).
	Initial(Pending).
	On("accept", Pending, Accepted).
	On("decline", Pending, Declined).
	On("cancel", Pending, Cancelled).
	After("expire", RequestsExpireAfter, Pending, Expired).
	OnEnter(Accepted, stampAccepted).
	OnTransition(announceLink)

// stampAccepted records since when two users are contacts.
func stampAccepted(ctx context.Context, l *Link) error {
	now := wire.Now(ctx)
	l.AcceptedAt = &now
	return nil
}

// InviteStatus is where an invitation by email stands.
type InviteStatus string

// The states of an invitation by email.
const (
	// Invited: the person has no account yet.
	Invited InviteStatus = "pending"
	// Joined: they signed up; the invitation became a contact request.
	Joined InviteStatus = "joined"
)

// Invite is an invitation to someone who has no account yet.
type Invite struct {
	ID        string       `json:"id"`
	InviterID string       `json:"inviterId"`
	Email     string       `json:"email"`
	Status    InviteStatus `json:"status"`
	CreatedAt time.Time    `json:"createdAt"`
	JoinedAt  *time.Time   `json:"joinedAt,omitempty"`
}

// Invites keeps the invitations by email, found by the address they wait for
// and by who sent them.
var Invites = Service.Store("invites", func(i Invite) string { return i.ID },
	kit.Index("email", func(i Invite) []string { return []string{i.Email} }),
	kit.Index("inviter", func(i Invite) []string { return []string{i.InviterID} }))

// The kinds of Event.
const (
	KindRequested = "requested"
	KindAccepted  = "accepted"
	KindInvited   = "invited"
)

// Event is what happened between two people.
type Event struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	LinkID   string `json:"linkId,omitempty"`
	InviteID string `json:"inviteId,omitempty"`
	// ActorID asked, accepted or invited.
	ActorID string `json:"actorId"`
	// UserID is the other user: the addressee of a request, the requester
	// of an accepted one.
	UserID string `json:"userId,omitempty"`
	// Email is the invited address, for an invitation.
	Email string    `json:"email,omitempty"`
	At    time.Time `json:"at"`
}

// Events announces requests, acceptances and invitations: notify mails them,
// activity writes them in the feeds.
var Events = Service.Topic[Event]("events")

// announceLink publishes a new request and an accepted one, whoever made
// them: a user, or an invitation claimed by a new account.
func announceLink(ctx context.Context, c kit.Change[Link, LinkStatus]) error {
	e := Event{ID: kit.NewID("event"), LinkID: c.Entity.ID, At: wire.Now(ctx)}
	switch c.To {
	case Pending:
		e.Kind, e.ActorID, e.UserID = KindRequested, c.Entity.RequesterID, c.Entity.AddresseeID
	case Accepted:
		e.Kind, e.ActorID, e.UserID = KindAccepted, c.Entity.AddresseeID, c.Entity.RequesterID
	default:
		return nil
	}
	return Events.Publish(ctx, e)
}

// linkBetween returns the link between two users, if there is one. It reads
// one user's links through the index, so a pair never linked is an answer,
// not an error.
func linkBetween(ctx context.Context, a, b string) (Link, bool, error) {
	links, err := Links.Find(ctx, "user", a)
	if err != nil {
		return Link{}, false, err
	}
	for _, l := range links {
		if l.other(a) == b {
			return l, true, nil
		}
	}
	return Link{}, false, nil
}

// request asks, on behalf of from, to become a contact of to. A pair already
// linked answers a conflict; a request declined, cancelled or expired is
// replaced by the new one.
func request(ctx context.Context, from, to string) (Link, error) {
	old, found, err := linkBetween(ctx, from, to)
	if err != nil {
		return Link{}, err
	}
	switch {
	case found && old.Status == Accepted:
		return Link{}, kit.Conflict("you are already contacts")
	case found && old.Status == Pending:
		return Link{}, kit.Conflict("a contact request between you is already waiting for an answer")
	case found:
		if err := Links.Delete(ctx, old.Pair); err != nil && !wire.Is(err, kit.CodeNotFound) {
			return Link{}, err
		}
	}
	l, err := Requests.Start(ctx, Link{
		ID: kit.NewID("link"), Pair: pairOf(from, to), RequesterID: from, AddresseeID: to, CreatedAt: wire.Now(ctx),
	})
	if wire.Is(err, kit.CodeConflict) {
		return Link{}, kit.Conflict("a contact request between you is already waiting for an answer")
	}
	return l, err
}
