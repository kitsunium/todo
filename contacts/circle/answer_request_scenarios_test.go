// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"testing"

	"github.com/kitsunium/todo/contacts/link"
)

// givenAnswerRequestAcceptsARequest builds AnswerRequest for "accepts a request": the use case, its input, and
// what its output must be.
func givenAnswerRequestAcceptsARequest(t *testing.T) (AnswerRequest, AnswerRequestInput, func(*testing.T, AnswerRequestOutput)) {
	t.Helper()
	w := newWorld()
	l := w.linked("user_ada", "user_bob", link.RequestsPending)
	return w.answer(), AnswerRequestInput{User: "user_bob", ID: l.ID, Event: "accept"}, func(t *testing.T, out AnswerRequestOutput) {
		t.Helper()
		if out.Link.State != link.RequestsAccepted || !out.Link.UpdatedAt.After(epoch) || w.links[l.ID] != out.Link {
			t.Errorf("the link accepted: %+v, kept %+v", out.Link, w.links[l.ID])
		}
		if len(w.notices) != 1 || w.notices[0] != "contact-accepted by user_bob on "+l.ID {
			t.Errorf("the notices: %q", w.notices)
		}
	}
}

// givenAnswerRequestRefusesTheRequestOfOthers builds AnswerRequest for "refuses the request of others": the use case and the input it
// refuses with link.ErrNotFound.
func givenAnswerRequestRefusesTheRequestOfOthers(t *testing.T) (AnswerRequest, AnswerRequestInput) {
	t.Helper()
	w := newWorld()
	l := w.linked("user_ada", "user_bob", link.RequestsPending)
	return w.answer(), AnswerRequestInput{User: "user_cy", ID: l.ID, Event: "accept"}
}

// givenAnswerRequestRefusesTheRequesterWhoAccepts builds AnswerRequest for "refuses the requester who accepts": the use case and the input it
// refuses with link.ErrNotAllowed.
func givenAnswerRequestRefusesTheRequesterWhoAccepts(t *testing.T) (AnswerRequest, AnswerRequestInput) {
	t.Helper()
	w := newWorld()
	l := w.linked("user_ada", "user_bob", link.RequestsPending)
	return w.answer(), AnswerRequestInput{User: "user_ada", ID: l.ID, Event: "accept"}
}

// givenAnswerRequestRefusesARequestAlreadyAnswered builds AnswerRequest for "refuses a request already answered": the use case and the input it
// refuses with link.ErrNotAllowed.
func givenAnswerRequestRefusesARequestAlreadyAnswered(t *testing.T) (AnswerRequest, AnswerRequestInput) {
	t.Helper()
	w := newWorld()
	l := w.linked("user_ada", "user_bob", link.RequestsAccepted)
	return w.answer(), AnswerRequestInput{User: "user_bob", ID: l.ID, Event: "decline"}
}
