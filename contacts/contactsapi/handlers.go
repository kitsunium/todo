// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package contactsapi

import (
	"context"
	"errors"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/contacts/circle"
	"github.com/kitsunium/todo/contacts/link"
)

// list is GET /api/contacts's body; what it uses is in wired.
func list(ctx context.Context, in NoInput) (out Listing, err error) {
	me, err := caller(ctx)
	if err != nil {
		return Listing{}, err
	}
	res, err := wired.listContacts.Execute(ctx, circle.ListContactsInput{User: me})
	if err != nil {
		return Listing{}, err
	}
	return Listing{Contacts: res.Contacts, Incoming: res.Incoming, Outgoing: res.Outgoing}, nil
}

// add is POST /api/contacts's body; what it uses is in wired.
func add(ctx context.Context, in AddRequest) (out link.Link, err error) {
	me, err := caller(ctx)
	if err != nil {
		return link.Link{}, err
	}
	res, err := wired.requestContact.Execute(ctx, circle.RequestContactInput{From: me, Email: in.Email})
	if err != nil {
		return link.Link{}, refusal(err)
	}
	return res.Link, nil
}

// accept is POST /api/contacts/{id}/accept's body; what it uses is in wired.
func accept(ctx context.Context, in ByID) (out link.Link, err error) {
	return answer(ctx, in.ID, link.RequestsOnAccept)
}

// decline is POST /api/contacts/{id}/decline's body; what it uses is in wired.
func decline(ctx context.Context, in ByID) (out link.Link, err error) {
	return answer(ctx, in.ID, link.RequestsOnDecline)
}

// cancel is POST /api/contacts/{id}/cancel's body; what it uses is in wired.
func cancel(ctx context.Context, in ByID) (out link.Link, err error) {
	return answer(ctx, in.ID, link.RequestsOnCancel)
}

// remove is DELETE /api/contacts/{id}'s body; what it uses is in wired.
func remove(ctx context.Context, in ByID) (out NoInput, err error) {
	me, err := caller(ctx)
	if err != nil {
		return NoInput{}, err
	}
	if _, err := wired.removeContact.Execute(ctx, circle.RemoveContactInput{User: me, ID: in.ID}); err != nil {
		return NoInput{}, refusal(err)
	}
	return NoInput{}, nil
}

// answer runs AnswerRequest for one of the three routes that move a request.
func answer(ctx context.Context, id string, event link.RequestsEvent) (link.Link, error) {
	me, err := caller(ctx)
	if err != nil {
		return link.Link{}, err
	}
	res, err := wired.answerRequest.Execute(ctx, circle.AnswerRequestInput{User: me, ID: id, Event: string(event)})
	if err != nil {
		return link.Link{}, refusal(err)
	}
	return res.Link, nil
}

// check is the query "are-contacts"'s body.
func check(ctx context.Context, in Pair) (out Verdict, err error) {
	res, err := wired.areContacts.Execute(ctx, circle.AreContactsInput{A: in.One, B: in.Other})
	return Verdict{Yes: res.Yes}, err
}

// caller is the member the request comes from; every route asks for one.
func caller(ctx context.Context) (string, error) {
	uid, ok := app.UserID(ctx)
	if !ok {
		return "", app.Unauthenticated("sign in first")
	}
	return string(uid), nil
}

// refusal is a use case's error as the API answers it: 404 for a link that
// is not the caller's or an address nobody has, 409 for a pair already
// linked or a request that cannot be answered this way, 400 for oneself.
func refusal(err error) error {
	switch {
	case errors.Is(err, link.ErrNotFound):
		return app.NotFound("no such contact")
	case errors.Is(err, link.ErrUnknownMember):
		return app.NotFound("nobody has this address")
	case errors.Is(err, link.ErrSelf):
		e := app.Invalid("the request is invalid")
		e.Violations = []app.Violation{{Path: "email", Rule: "other", Message: err.Error()}}
		return e
	case errors.Is(err, link.ErrAlready):
		return app.Conflict("you are already contacts, or a request between you is waiting for an answer")
	case errors.Is(err, link.ErrNotAllowed):
		return app.Conflict("this request is not yours to answer this way")
	}
	return err
}
