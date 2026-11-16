// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:620afbb2962f72a1a0c57e91267127ee1982802bbd70150e8fa1d8268bc9f808

// Package contactsapi — The HTTP API of a member's contacts.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package contactsapi

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/contacts/circle"
	"github.com/kitsunium/todo/contacts/link"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 8, 4, 255), "NOT_IMPLEMENTED", "not implemented yet", "contactsapi: a hand-written body is still a skeleton")

// AddRequest — The address of the member asked.
type AddRequest struct {
	Email string `json:"email" app:"personal"`
}

// ByID — The link a route names.
type ByID struct {
	ID string `path:"id"`
}

// Listing — A member's contacts, and the requests waiting.
type Listing struct {
	Contacts []circle.Contact `json:"contacts"`
	Incoming []circle.Contact `json:"incoming"`
	Outgoing []circle.Contact `json:"outgoing"`
}

// NoInput — A request, or an answer, that says nothing.
type NoInput struct {
}

// Service is contactsapi's framework service: what it declares, drawn and run.
var Service = app.NewService("contactsapi", "The HTTP API of a member's contacts.")

var _ = Service.Endpoint("GET /api/contacts", List, app.Auth())

// List serves GET /api/contacts: its body is the hand-written list.
func List(ctx context.Context, in NoInput) (Listing, error) {
	return list(ctx, in)
}

var _ = Service.Endpoint("POST /api/contacts", Add, app.Auth(), app.RateLimitPerClient(2, 5))

// Add serves POST /api/contacts: its body is the hand-written add.
func Add(ctx context.Context, in AddRequest) (link.Link, error) {
	return add(ctx, in)
}

var _ = Service.Endpoint("POST /api/contacts/{id}/accept", Accept, app.Auth(), app.RateLimitPerClient(10, 20))

// Accept serves POST /api/contacts/{id}/accept: its body is the hand-written accept.
func Accept(ctx context.Context, in ByID) (link.Link, error) {
	return accept(ctx, in)
}

var _ = Service.Endpoint("POST /api/contacts/{id}/decline", Decline, app.Auth(), app.RateLimitPerClient(10, 20))

// Decline serves POST /api/contacts/{id}/decline: its body is the hand-written decline.
func Decline(ctx context.Context, in ByID) (link.Link, error) {
	return decline(ctx, in)
}

var _ = Service.Endpoint("POST /api/contacts/{id}/cancel", Cancel, app.Auth(), app.RateLimitPerClient(10, 20))

// Cancel serves POST /api/contacts/{id}/cancel: its body is the hand-written cancel.
func Cancel(ctx context.Context, in ByID) (link.Link, error) {
	return cancel(ctx, in)
}

var _ = Service.Endpoint("DELETE /api/contacts/{id}", Remove, app.Auth(), app.RateLimitPerClient(10, 20))

// Remove serves DELETE /api/contacts/{id}: its body is the hand-written remove.
func Remove(ctx context.Context, in ByID) (NoInput, error) {
	return remove(ctx, in)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
	listContacts   circle.ListContacts
	requestContact circle.RequestContact
	answerRequest  circle.AnswerRequest
	removeContact  circle.RemoveContact
}

// Wire gives the handlers what they use: the role that mounts contactsapi calls it
// before its App runs.
func Wire(listContacts circle.ListContacts, requestContact circle.RequestContact, answerRequest circle.AnswerRequest, removeContact circle.RemoveContact) {
	wired.listContacts = listContacts
	wired.requestContact = requestContact
	wired.answerRequest = answerRequest
	wired.removeContact = removeContact
}
