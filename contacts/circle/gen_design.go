// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:4a7a0b8d96d46792f2d63823e44fbbe64f6e4094e6b613890cc9ef5bff2ece56

// Package circle — Asking a member to be a contact, answering, and listing
// one's contacts.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package circle

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/contacts/link"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 8, 2, 255), "NOT_IMPLEMENTED", "not implemented yet", "circle: a hand-written body is still a skeleton")

const (
	// KindInvite is the answer to an address nobody has, which was invited to sign
	// up.
	KindInvite string = "invite"
	// KindRequest is the answer to a member who was asked to be a contact.
	KindRequest string = "request"
)

// Contact — A link as one of its two members reads it.
type Contact struct {
	ID    string    `json:"id" app:"public"`
	User  Person    `json:"user"`
	Since time.Time `json:"since" app:"public"`
}

// Invitation — An invitation a member sent, as they read it.
type Invitation struct {
	ID        string    `json:"id" app:"public"`
	Email     string    `json:"email" app:"personal"`
	CreatedAt time.Time `json:"createdAt" app:"public"`
}

// Person — A member as a contact sees them.
type Person struct {
	ID    string `json:"id" app:"public"`
	Email string `json:"email" app:"personal"`
	Name  string `json:"name" app:"personal"`
	// Locale is the language the member reads, fr or en: it is never answered, it
	// decides the language of a mail written for them.
	Locale string `json:"-" app:"public"`
}

// ClockV1 tells the time.
type ClockV1 interface {
	Now(ctx context.Context) (time.Time, error)
}

// InviterV1 mails an address nobody has the invitation to sign up.
type InviterV1 interface {
	Invite(ctx context.Context, to string, by Person) error
}

// InvitesV1 keeps the invitations to those who have no account yet.
type InvitesV1 interface {
	// ByEmail is every invitation waiting for an address, whoever sent it.
	ByEmail(ctx context.Context, email string) ([]link.Invite, error)
	// Change applies a change to the invitation of an id as it is kept at that
	// instant, alone among the writers of the store, and keeps the result; an
	// error of the change keeps nothing, and an invitation that is not there is
	// ErrNotFound.
	Change(ctx context.Context, id string, change func(*link.Invite) error) (link.Invite, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (link.Invite, bool, error)
	NewID(ctx context.Context) (string, error)
	// Of is the invitations a member sent.
	Of(ctx context.Context, user string) ([]link.Invite, error)
	Save(ctx context.Context, i link.Invite) error
}

// LinksV1 keeps the links.
type LinksV1 interface {
	ByPair(ctx context.Context, pair string) (link.Link, bool, error)
	// Change applies a change to the link of an id as it is kept at that instant,
	// alone among the writers of the store, and keeps the result; an error of the
	// change keeps nothing, and a link that is not there is ErrNotFound.
	Change(ctx context.Context, id string, change func(*link.Link) error) (link.Link, error)
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (link.Link, bool, error)
	NewID(ctx context.Context) (string, error)
	Of(ctx context.Context, user string) ([]link.Link, error)
	Save(ctx context.Context, l link.Link) error
}

// NoticesV1 is told what happened to a link, for whoever tells the members.
type NoticesV1 interface {
	Tell(ctx context.Context, kind string, l link.Link, actor Person) error
}

// PeopleV1 is the directory of the members.
type PeopleV1 interface {
	ByEmail(ctx context.Context, email string) (Person, bool, error)
	ByID(ctx context.Context, id string) (Person, bool, error)
}

// AnswerRequest moves a pending request by one event, accept or decline for
// its addressee, cancel for its requester.
type AnswerRequest interface {
	Execute(ctx context.Context, in AnswerRequestInput) (AnswerRequestOutput, error)
}

// AnswerRequestInput is AnswerRequest's input: a field added breaks no caller.
type AnswerRequestInput struct {
	User  string
	ID    string
	Event string
}

// AnswerRequestOutput is AnswerRequest's output.
type AnswerRequestOutput struct {
	Link link.Link
}

// AnswerRequestHandler runs AnswerRequest: its body is the hand-written execute.
type AnswerRequestHandler struct {
	clock   ClockV1
	links   LinksV1
	people  PeopleV1
	notices NoticesV1
}

// NewAnswerRequest builds AnswerRequest on what it calls.
func NewAnswerRequest(clock ClockV1, links LinksV1, people PeopleV1, notices NoticesV1) *AnswerRequestHandler {
	return &AnswerRequestHandler{clock: clock, links: links, people: people, notices: notices}
}

// Execute runs AnswerRequest.
func (h *AnswerRequestHandler) Execute(ctx context.Context, in AnswerRequestInput) (AnswerRequestOutput, error) {
	return h.execute(ctx, &in)
}

var _ AnswerRequest = (*AnswerRequestHandler)(nil)

// AreContacts says whether two members are contacts.
type AreContacts interface {
	Execute(ctx context.Context, in AreContactsInput) (AreContactsOutput, error)
}

// AreContactsInput is AreContacts's input: a field added breaks no caller.
type AreContactsInput struct {
	A string
	B string
}

// AreContactsOutput is AreContacts's output.
type AreContactsOutput struct {
	Yes bool
}

// AreContactsHandler runs AreContacts: its body is the hand-written execute.
type AreContactsHandler struct {
	links LinksV1
}

// NewAreContacts builds AreContacts on what it calls.
func NewAreContacts(links LinksV1) *AreContactsHandler {
	return &AreContactsHandler{links: links}
}

// Execute runs AreContacts.
func (h *AreContactsHandler) Execute(ctx context.Context, in AreContactsInput) (AreContactsOutput, error) {
	return h.execute(ctx, &in)
}

var _ AreContacts = (*AreContactsHandler)(nil)

// CancelInvite withdraws an invitation of the member's own, which nobody else
// may.
type CancelInvite interface {
	Execute(ctx context.Context, in CancelInviteInput) (CancelInviteOutput, error)
}

// CancelInviteInput is CancelInvite's input: a field added breaks no caller.
type CancelInviteInput struct {
	User string
	ID   string
}

// CancelInviteOutput is CancelInvite's output.
type CancelInviteOutput struct {
	Invitation Invitation
}

// CancelInviteHandler runs CancelInvite: its body is the hand-written execute.
type CancelInviteHandler struct {
	invites InvitesV1
}

// NewCancelInvite builds CancelInvite on what it calls.
func NewCancelInvite(invites InvitesV1) *CancelInviteHandler {
	return &CancelInviteHandler{invites: invites}
}

// Execute runs CancelInvite.
func (h *CancelInviteHandler) Execute(ctx context.Context, in CancelInviteInput) (CancelInviteOutput, error) {
	return h.execute(ctx, &in)
}

var _ CancelInvite = (*CancelInviteHandler)(nil)

// ClaimInvites turns every invitation waiting for an address that was just
// proven into a request from its inviter to the new member, and marks it
// joined; a pair already linked and an inviter who is the member are left
// alone, and claiming twice changes nothing.
type ClaimInvites interface {
	Execute(ctx context.Context, in ClaimInvitesInput) (ClaimInvitesOutput, error)
}

// ClaimInvitesInput is ClaimInvites's input: a field added breaks no caller.
type ClaimInvitesInput struct {
	UserID string
}

// ClaimInvitesOutput is ClaimInvites's output.
type ClaimInvitesOutput struct {
}

// ClaimInvitesHandler runs ClaimInvites: its body is the hand-written execute.
type ClaimInvitesHandler struct {
	people  PeopleV1
	invites InvitesV1
	links   LinksV1
	clock   ClockV1
	notices NoticesV1
}

// NewClaimInvites builds ClaimInvites on what it calls.
func NewClaimInvites(people PeopleV1, invites InvitesV1, links LinksV1, clock ClockV1, notices NoticesV1) *ClaimInvitesHandler {
	return &ClaimInvitesHandler{people: people, invites: invites, links: links, clock: clock, notices: notices}
}

// Execute runs ClaimInvites.
func (h *ClaimInvitesHandler) Execute(ctx context.Context, in ClaimInvitesInput) (ClaimInvitesOutput, error) {
	return h.execute(ctx, &in)
}

var _ ClaimInvites = (*ClaimInvitesHandler)(nil)

// ListContacts lists a member's contacts by name, the requests waiting, those
// they received and those they sent, and the invitations still waiting for an
// account, the newest first.
type ListContacts interface {
	Execute(ctx context.Context, in ListContactsInput) (ListContactsOutput, error)
}

// ListContactsInput is ListContacts's input: a field added breaks no caller.
type ListContactsInput struct {
	User string
}

// ListContactsOutput is ListContacts's output.
type ListContactsOutput struct {
	Contacts []Contact
	Incoming []Contact
	Outgoing []Contact
	Invites  []Invitation
}

// ListContactsHandler runs ListContacts: its body is the hand-written execute.
type ListContactsHandler struct {
	links   LinksV1
	people  PeopleV1
	invites InvitesV1
}

// NewListContacts builds ListContacts on what it calls.
func NewListContacts(links LinksV1, people PeopleV1, invites InvitesV1) *ListContactsHandler {
	return &ListContactsHandler{links: links, people: people, invites: invites}
}

// Execute runs ListContacts.
func (h *ListContactsHandler) Execute(ctx context.Context, in ListContactsInput) (ListContactsOutput, error) {
	return h.execute(ctx, &in)
}

var _ ListContacts = (*ListContactsHandler)(nil)

// RemoveContact ends a contact, for either of its two members.
type RemoveContact interface {
	Execute(ctx context.Context, in RemoveContactInput) (RemoveContactOutput, error)
}

// RemoveContactInput is RemoveContact's input: a field added breaks no caller.
type RemoveContactInput struct {
	User string
	ID   string
}

// RemoveContactOutput is RemoveContact's output.
type RemoveContactOutput struct {
}

// RemoveContactHandler runs RemoveContact: its body is the hand-written execute.
type RemoveContactHandler struct {
	links LinksV1
}

// NewRemoveContact builds RemoveContact on what it calls.
func NewRemoveContact(links LinksV1) *RemoveContactHandler {
	return &RemoveContactHandler{links: links}
}

// Execute runs RemoveContact.
func (h *RemoveContactHandler) Execute(ctx context.Context, in RemoveContactInput) (RemoveContactOutput, error) {
	return h.execute(ctx, &in)
}

var _ RemoveContact = (*RemoveContactHandler)(nil)

// RequestContact asks the member an address belongs to to be a contact; an
// address nobody has is invited to sign up instead, and the invitation waits
// for it.
type RequestContact interface {
	Execute(ctx context.Context, in RequestContactInput) (RequestContactOutput, error)
}

// RequestContactInput is RequestContact's input: a field added breaks no caller.
type RequestContactInput struct {
	From  string
	Email string
}

// RequestContactOutput is RequestContact's output.
type RequestContactOutput struct {
	// Kind is request when a member was asked, invite when an address nobody has
	// was invited.
	Kind       string
	Link       link.Link
	Invitation Invitation
}

// RequestContactHandler runs RequestContact: its body is the hand-written execute.
type RequestContactHandler struct {
	people  PeopleV1
	links   LinksV1
	clock   ClockV1
	notices NoticesV1
	invites InvitesV1
	inviter InviterV1
}

// NewRequestContact builds RequestContact on what it calls.
func NewRequestContact(people PeopleV1, links LinksV1, clock ClockV1, notices NoticesV1, invites InvitesV1, inviter InviterV1) *RequestContactHandler {
	return &RequestContactHandler{people: people, links: links, clock: clock, notices: notices, invites: invites, inviter: inviter}
}

// Execute runs RequestContact.
func (h *RequestContactHandler) Execute(ctx context.Context, in RequestContactInput) (RequestContactOutput, error) {
	return h.execute(ctx, &in)
}

var _ RequestContact = (*RequestContactHandler)(nil)
