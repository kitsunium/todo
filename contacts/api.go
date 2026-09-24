package contacts

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/identity"
	"github.com/kitsunium/todo/internal/wire"
)

// The contacts API. Every route acts for the signed-in user, on links they
// are part of: anyone else's link does not exist for them.
var (
	_ = Service.Endpoint("GET /api/contacts", List, kit.Auth())
	_ = Service.Endpoint("POST /api/contacts", Add, kit.Auth(), kit.RateLimitPerClient(2, 5))
	_ = Service.Endpoint("POST /api/contacts/{id}/accept", Accept, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("POST /api/contacts/{id}/decline", Decline, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("POST /api/contacts/{id}/cancel", Cancel, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("DELETE /api/contacts/{id}", Remove, kit.Auth(), kit.RateLimitPerClient(10, 20))

	// CheckAPI tells another service whether two users are contacts: tasks
	// shares only with contacts, groups invites only contacts.
	CheckAPI = Service.Endpoint("GET /internal/contacts/check", Check, kit.Private())
)

// Contact is someone the user works with.
type Contact struct {
	ID    string           `json:"id"`
	User  identity.UserRef `json:"user"`
	Since time.Time        `json:"since"`
}

// Request is a contact request, from the user or to them.
type Request struct {
	ID        string           `json:"id"`
	User      identity.UserRef `json:"user"`
	CreatedAt time.Time        `json:"createdAt"`
	Status    LinkStatus       `json:"status"`
}

// Invitation is an invitation by email the user sent.
type Invitation struct {
	ID        string       `json:"id"`
	Email     string       `json:"email"`
	CreatedAt time.Time    `json:"createdAt"`
	Status    InviteStatus `json:"status"`
}

// ListOutput is everyone the user works with or may soon.
type ListOutput struct {
	Contacts []Contact    `json:"contacts"`
	Incoming []Request    `json:"incoming"`
	Outgoing []Request    `json:"outgoing"`
	Invites  []Invitation `json:"invites"`
}

// me is the signed-in user.
func me(ctx context.Context) (string, error) {
	uid, ok := kit.UserID(ctx)
	if !ok {
		return "", kit.Unauthenticated("sign in first")
	}
	return string(uid), nil
}

// List returns the user's contacts, by name; the requests waiting for them
// and those they sent, newest first; and their invitations still waiting for
// an account.
func List(ctx context.Context, _ kit.Empty) (ListOutput, error) {
	uid, err := me(ctx)
	if err != nil {
		return ListOutput{}, err
	}
	links, err := Links.Find(ctx, "user", uid)
	if err != nil {
		return ListOutput{}, err
	}
	invites, err := Invites.Find(ctx, "inviter", uid)
	if err != nil {
		return ListOutput{}, err
	}
	var ids []string
	for _, l := range links {
		ids = append(ids, l.other(uid))
	}
	users, err := identity.Directory(ctx, ids...)
	if err != nil {
		return ListOutput{}, err
	}
	out := ListOutput{Contacts: []Contact{}, Incoming: []Request{}, Outgoing: []Request{}, Invites: []Invitation{}}
	for _, l := range links {
		user := users[l.other(uid)]
		switch {
		case l.Status == Accepted:
			out.Contacts = append(out.Contacts, contactOf(l, user))
		case l.Status == Pending && l.AddresseeID == uid:
			out.Incoming = append(out.Incoming, requestOf(l, user))
		case l.Status == Pending:
			out.Outgoing = append(out.Outgoing, requestOf(l, user))
		}
	}
	for _, i := range invites {
		if i.Status == Invited {
			out.Invites = append(out.Invites, invitationOf(i))
		}
	}
	slices.SortFunc(out.Contacts, func(a, b Contact) int {
		return cmp.Or(cmp.Compare(strings.ToLower(a.User.Name), strings.ToLower(b.User.Name)), cmp.Compare(a.ID, b.ID))
	})
	newest := func(a, b Request) int { return cmp.Or(b.CreatedAt.Compare(a.CreatedAt), cmp.Compare(a.ID, b.ID)) }
	slices.SortFunc(out.Incoming, newest)
	slices.SortFunc(out.Outgoing, newest)
	slices.SortFunc(out.Invites, func(a, b Invitation) int { return b.CreatedAt.Compare(a.CreatedAt) })
	return out, nil
}

func contactOf(l Link, user identity.UserRef) Contact {
	since := l.CreatedAt
	if l.AcceptedAt != nil {
		since = *l.AcceptedAt
	}
	return Contact{ID: l.ID, User: user, Since: since}
}

func requestOf(l Link, user identity.UserRef) Request {
	return Request{ID: l.ID, User: user, CreatedAt: l.CreatedAt, Status: l.Status}
}

func invitationOf(i Invite) Invitation {
	return Invitation{ID: i.ID, Email: i.Email, CreatedAt: i.CreatedAt, Status: i.Status}
}

// AddInput is the address of someone to work with.
type AddInput struct {
	Email string `json:"email" validate:"required,maxlen=254"`
}

// AddOutput is what adding someone did: a contact request when they have an
// account, an invitation by email otherwise.
type AddOutput struct {
	Kind    string      `json:"kind"`
	Request *Request    `json:"request,omitempty"`
	Invite  *Invitation `json:"invite,omitempty"`
}

// Add asks someone to become a contact. With an account, they get a request
// to accept; without one, an invitation by email to join, which becomes a
// request once they sign up.
func Add(ctx context.Context, in AddInput) (AddOutput, error) {
	uid, err := me(ctx)
	if err != nil {
		return AddOutput{}, err
	}
	email := identity.NormalizeEmail(in.Email)
	if !identity.ValidEmail(email) {
		return AddOutput{}, wire.Invalid("email", "email", "must be an email address")
	}
	found, err := identity.UserByEmailAPI.Call(ctx, identity.EmailQuery{Email: email})
	if err != nil {
		return AddOutput{}, err
	}
	if u := found.User; u != nil {
		if u.ID == uid {
			return AddOutput{}, wire.Invalid("email", "self", "is your own address")
		}
		l, err := request(ctx, uid, u.ID)
		if err != nil {
			return AddOutput{}, err
		}
		r := requestOf(l, *u)
		return AddOutput{Kind: "request", Request: &r}, nil
	}
	sent, err := Invites.Find(ctx, "email", email)
	if err != nil {
		return AddOutput{}, err
	}
	for _, i := range sent {
		if i.InviterID == uid && i.Status == Invited {
			return AddOutput{}, kit.Conflict("you already invited this address")
		}
	}
	inv := Invite{ID: kit.NewID("invite"), InviterID: uid, Email: email, Status: Invited, CreatedAt: wire.Now(ctx)}
	if err := Invites.Insert(ctx, inv); err != nil {
		return AddOutput{}, err
	}
	if err := Events.Publish(ctx, Event{
		ID: kit.NewID("event"), Kind: KindInvited, InviteID: inv.ID, ActorID: uid, Email: email, At: inv.CreatedAt,
	}); err != nil {
		return AddOutput{}, err
	}
	v := invitationOf(inv)
	return AddOutput{Kind: "invite", Invite: &v}, nil
}

// ByID addresses a contact, a request or an invitation.
type ByID struct {
	ID string `path:"id"`
}

// ContactOutput is a new contact: {contact}.
type ContactOutput struct {
	Contact Contact `json:"contact"`
}

// errNoLink answers for a link that does not exist, or not for the caller.
func errNoLink() error { return kit.NotFound("no such contact or request") }

// errAnswered answers for a request that is not waiting any more.
func errAnswered() error { return kit.Conflict("this request is not waiting for an answer any more") }

// linkOf returns the link with the given ID, if uid is part of it.
func linkOf(ctx context.Context, id, uid string) (Link, error) {
	l, err := Links.Lookup(ctx, "id", id)
	if wire.Is(err, kit.CodeNotFound) || (err == nil && l.RequesterID != uid && l.AddresseeID != uid) {
		return Link{}, errNoLink()
	}
	return l, err
}

// Accept makes the requester a contact. Only the addressee accepts.
func Accept(ctx context.Context, in ByID) (ContactOutput, error) {
	uid, err := me(ctx)
	if err != nil {
		return ContactOutput{}, err
	}
	l, err := linkOf(ctx, in.ID, uid)
	if err != nil {
		return ContactOutput{}, err
	}
	if l.AddresseeID != uid {
		return ContactOutput{}, kit.Forbidden("only the person asked can accept a request")
	}
	if l.Status != Pending {
		return ContactOutput{}, errAnswered()
	}
	l, err = Requests.Fire(ctx, l.Pair, "accept")
	if wire.Is(err, kit.CodeConflict) {
		return ContactOutput{}, errAnswered()
	}
	if err != nil {
		return ContactOutput{}, err
	}
	users, err := identity.Directory(ctx, l.RequesterID)
	if err != nil {
		return ContactOutput{}, err
	}
	return ContactOutput{Contact: contactOf(l, users[l.RequesterID])}, nil
}

// Decline turns a request down. Only the addressee declines; the requester
// is not told.
func Decline(ctx context.Context, in ByID) (kit.Empty, error) {
	uid, err := me(ctx)
	if err != nil {
		return kit.Empty{}, err
	}
	l, err := linkOf(ctx, in.ID, uid)
	if err != nil {
		return kit.Empty{}, err
	}
	if l.AddresseeID != uid {
		return kit.Empty{}, kit.Forbidden("only the person asked can decline a request")
	}
	if l.Status != Pending {
		return kit.Empty{}, errAnswered()
	}
	if _, err := Requests.Fire(ctx, l.Pair, "decline"); wire.Is(err, kit.CodeConflict) {
		return kit.Empty{}, errAnswered()
	} else if err != nil {
		return kit.Empty{}, err
	}
	return kit.Empty{}, nil
}

// Cancel withdraws a request, or an invitation by email, the user sent.
func Cancel(ctx context.Context, in ByID) (kit.Empty, error) {
	uid, err := me(ctx)
	if err != nil {
		return kit.Empty{}, err
	}
	if strings.HasPrefix(in.ID, "invite_") {
		return kit.Empty{}, withdrawInvite(ctx, in.ID, uid)
	}
	l, err := linkOf(ctx, in.ID, uid)
	if err != nil {
		return kit.Empty{}, err
	}
	if l.RequesterID != uid {
		return kit.Empty{}, kit.Forbidden("only the person who asked can cancel a request")
	}
	if l.Status != Pending {
		return kit.Empty{}, errAnswered()
	}
	if _, err := Requests.Fire(ctx, l.Pair, "cancel"); wire.Is(err, kit.CodeConflict) {
		return kit.Empty{}, errAnswered()
	} else if err != nil {
		return kit.Empty{}, err
	}
	return kit.Empty{}, nil
}

// Remove ends a contact — either side may — or deletes a request or an
// invitation the user is part of.
func Remove(ctx context.Context, in ByID) (kit.Empty, error) {
	uid, err := me(ctx)
	if err != nil {
		return kit.Empty{}, err
	}
	if strings.HasPrefix(in.ID, "invite_") {
		return kit.Empty{}, withdrawInvite(ctx, in.ID, uid)
	}
	l, err := linkOf(ctx, in.ID, uid)
	if err != nil {
		return kit.Empty{}, err
	}
	if err := Links.Delete(ctx, l.Pair); err != nil && !wire.Is(err, kit.CodeNotFound) {
		return kit.Empty{}, err
	}
	return kit.Empty{}, nil
}

// withdrawInvite deletes an invitation by email the user sent.
func withdrawInvite(ctx context.Context, id, uid string) error {
	i, err := Invites.Get(ctx, id)
	if wire.Is(err, kit.CodeNotFound) || (err == nil && i.InviterID != uid) {
		return errNoLink()
	}
	if err != nil {
		return err
	}
	if err := Invites.Delete(ctx, id); err != nil && !wire.Is(err, kit.CodeNotFound) {
		return err
	}
	return nil
}

// CheckInput names two users.
type CheckInput struct {
	A string `query:"a"`
	B string `query:"b"`
}

// CheckOutput says whether they are contacts.
type CheckOutput struct {
	Contacts bool `json:"contacts"`
}

// Check reports whether two users are contacts.
func Check(ctx context.Context, in CheckInput) (CheckOutput, error) {
	if in.A == "" || in.B == "" || in.A == in.B {
		return CheckOutput{}, nil
	}
	l, found, err := linkBetween(ctx, in.A, in.B)
	return CheckOutput{Contacts: found && l.Status == Accepted}, err
}
