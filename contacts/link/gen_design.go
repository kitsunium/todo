// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:a531db8fdf9b03e4efb3cfcdc2c2b563c368ecbbcb719b2c42901921e493400b

// Package link — The link between two members, and the request that makes
// it.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package link

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/pkg/v1/errs"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 8, 1, 255), "NOT_IMPLEMENTED", "not implemented yet", "link: a hand-written body is still a skeleton")

var (
	// ErrAlready is "you are already contacts, or a request between you is waiting
	// for an answer".
	ErrAlready = errs.New(errs.Pack(64, 8, 1, 4), "ALREADY", "you are already contacts, or a request between you is waiting for an answer", "contacts/error/ErrAlready")
	// ErrNotAllowed is "this request is not yours to answer this way".
	ErrNotAllowed = errs.New(errs.Pack(64, 8, 1, 5), "NOT_ALLOWED", "this request is not yours to answer this way", "contacts/error/ErrNotAllowed")
	// ErrNotFound is "no such contact".
	ErrNotFound = errs.New(errs.Pack(64, 8, 1, 1), "NOT_FOUND", "no such contact", "contacts/error/ErrNotFound")
	// ErrSelf is "a member is not their own contact".
	ErrSelf = errs.New(errs.Pack(64, 8, 1, 3), "SELF", "a member is not their own contact", "contacts/error/ErrSelf")
	// ErrUnknownMember is "nobody has this address".
	ErrUnknownMember = errs.New(errs.Pack(64, 8, 1, 2), "UNKNOWN_MEMBER", "nobody has this address", "contacts/error/ErrUnknownMember")
)

// Link — A contact request between two members, and the contact it becomes
// once accepted.
type Link struct {
	ID string `json:"id" app:"public"`
	// Pair names the two members whatever their order: there is one link a pair at
	// most.
	Pair        string        `json:"pair" app:"public"`
	RequesterID string        `json:"requesterId" app:"public"`
	AddresseeID string        `json:"addresseeId" app:"public"`
	State       RequestsState `json:"state" app:"public"`
	CreatedAt   time.Time     `json:"createdAt" app:"public"`
	UpdatedAt   time.Time     `json:"updatedAt" app:"public"`
}

// Other is the member at the other end of the link from user, and whether user
// is at one end.
func (l *Link) Other(user string) (string, bool) {
	return l.other(user)
}

// PairOf names two members whatever their order.
func PairOf(a string, b string) string {
	return pairOf(a, b)
}

// RequestsState is a state of Requests.
type RequestsState string

// The states of Requests.
const (
	RequestsPending   RequestsState = "pending"
	RequestsAccepted  RequestsState = "accepted"
	RequestsDeclined  RequestsState = "declined"
	RequestsCancelled RequestsState = "cancelled"
	RequestsExpired   RequestsState = "expired"
)

// RequestsInitial is the state a Link starts in.
const RequestsInitial = RequestsPending

// RequestsEvent is an event of Requests.
type RequestsEvent string

// The events of Requests.
const (
	RequestsOnAccept  RequestsEvent = "accept"
	RequestsOnDecline RequestsEvent = "decline"
	RequestsOnCancel  RequestsEvent = "cancel"
)

var (
	// ErrRequestsRefused is a transition its guard refused.
	ErrRequestsRefused = errs.New(errs.Pack(64, 8, 1, 240), "REQUESTS_REFUSED", "the transition is refused", "contacts/workflow/Requests")
	// ErrRequestsNoTransition is an event its state does not answer.
	ErrRequestsNoTransition = errs.New(errs.Pack(64, 8, 1, 241), "REQUESTS_NO_TRANSITION", "no transition for this event in this state", "contacts/workflow/Requests")
)

// Requests — A request is pending until its addressee accepts or declines
// it, or its requester cancels it; one nobody answered for fourteen days has
// expired.
type Requests struct {
	requests // its own state, hand-written
}

// NewRequests builds Requests: its body is the hand-written newRequests.
func NewRequests() *Requests {
	return newRequests()
}

// Fire moves e from state from by event: the state it enters, or the state
// it stays in and why.
func (w *Requests) Fire(ctx context.Context, e *Link, from RequestsState, event RequestsEvent) (RequestsState, error) {
	switch {
	case from == RequestsPending && event == RequestsOnAccept:
		return RequestsAccepted, nil
	case from == RequestsPending && event == RequestsOnDecline:
		return RequestsDeclined, nil
	case from == RequestsPending && event == RequestsOnCancel:
		return RequestsCancelled, nil
	}
	return from, ErrRequestsNoTransition
}
