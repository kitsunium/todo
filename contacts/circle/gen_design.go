// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:0c616a2ddc211ada253896609000dcaab6758455e2991084b86e140dd79690fe

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

// Contact — A link as one of its two members reads it.
type Contact struct {
	ID    string    `json:"id" app:"public"`
	User  Person    `json:"user"`
	Since time.Time `json:"since" app:"public"`
}

// Person — A member as a contact sees them.
type Person struct {
	ID    string `json:"id" app:"public"`
	Email string `json:"email" app:"personal"`
	Name  string `json:"name" app:"personal"`
}

// ClockV1 tells the time.
type ClockV1 interface {
	Now(ctx context.Context) (time.Time, error)
}

// LinksV1 keeps the links.
type LinksV1 interface {
	ByPair(ctx context.Context, pair string) (link.Link, bool, error)
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
	links   LinksV1
	clock   ClockV1
	people  PeopleV1
	notices NoticesV1
}

// NewAnswerRequest builds AnswerRequest on what it calls.
func NewAnswerRequest(links LinksV1, clock ClockV1, people PeopleV1, notices NoticesV1) *AnswerRequestHandler {
	return &AnswerRequestHandler{links: links, clock: clock, people: people, notices: notices}
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

// ListContacts lists a member's contacts by name, and the requests waiting,
// those they received and those they sent.
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
}

// ListContactsHandler runs ListContacts: its body is the hand-written execute.
type ListContactsHandler struct {
	links  LinksV1
	people PeopleV1
}

// NewListContacts builds ListContacts on what it calls.
func NewListContacts(links LinksV1, people PeopleV1) *ListContactsHandler {
	return &ListContactsHandler{links: links, people: people}
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

// RequestContact asks the member an address belongs to to be a contact.
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
	Link link.Link
}

// RequestContactHandler runs RequestContact: its body is the hand-written execute.
type RequestContactHandler struct {
	people  PeopleV1
	links   LinksV1
	clock   ClockV1
	notices NoticesV1
}

// NewRequestContact builds RequestContact on what it calls.
func NewRequestContact(people PeopleV1, links LinksV1, clock ClockV1, notices NoticesV1) *RequestContactHandler {
	return &RequestContactHandler{people: people, links: links, clock: clock, notices: notices}
}

// Execute runs RequestContact.
func (h *RequestContactHandler) Execute(ctx context.Context, in RequestContactInput) (RequestContactOutput, error) {
	return h.execute(ctx, &in)
}

var _ RequestContact = (*RequestContactHandler)(nil)
