// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"testing"

	"github.com/kitsunium/todo/contacts/link"
)

// givenRequestContactAsksAMember builds RequestContact for "asks a member": the use case, its input, and
// what its output must be.
func givenRequestContactAsksAMember(t *testing.T) (RequestContact, RequestContactInput, func(*testing.T, RequestContactOutput)) {
	t.Helper()
	w := newWorld()
	return w.request(), RequestContactInput{From: "user_ada", Email: "bob@example.com"}, func(t *testing.T, out RequestContactOutput) {
		t.Helper()
		l := out.Link
		if l.RequesterID != "user_ada" || l.AddresseeID != "user_bob" || l.State != link.RequestsPending || l.Pair != link.PairOf("user_bob", "user_ada") {
			t.Errorf("the request: %+v", l)
		}
		if kept := w.links[l.ID]; kept != l {
			t.Errorf("the store keeps %+v, want %+v", kept, l)
		}
		if len(w.notices) != 1 || w.notices[0] != "contact-requested by user_ada on "+l.ID {
			t.Errorf("the notices: %q", w.notices)
		}
	}
}

// givenRequestContactAsksAgainAfterARefusal builds RequestContact for "asks again after a refusal": the use case, its input, and
// what its output must be.
func givenRequestContactAsksAgainAfterARefusal(t *testing.T) (RequestContact, RequestContactInput, func(*testing.T, RequestContactOutput)) {
	t.Helper()
	w := newWorld()
	old := w.linked("user_bob", "user_ada", link.RequestsDeclined)
	return w.request(), RequestContactInput{From: "user_ada", Email: "bob@example.com"}, func(t *testing.T, out RequestContactOutput) {
		t.Helper()
		if _, still := w.links[old.ID]; still || len(w.links) != 1 || out.Link.State != link.RequestsPending || out.Link.RequesterID != "user_ada" {
			t.Errorf("the pair's links: %+v, want the new request alone", w.links)
		}
	}
}

// givenRequestContactRefusesAnUnknownAddress builds RequestContact for "refuses an unknown address": the use case and the input it
// refuses with link.ErrUnknownMember.
func givenRequestContactRefusesAnUnknownAddress(t *testing.T) (RequestContact, RequestContactInput) {
	t.Helper()
	return newWorld().request(), RequestContactInput{From: "user_ada", Email: "nobody@example.com"}
}

// givenRequestContactRefusesOneself builds RequestContact for "refuses oneself": the use case and the input it
// refuses with link.ErrSelf.
func givenRequestContactRefusesOneself(t *testing.T) (RequestContact, RequestContactInput) {
	t.Helper()
	return newWorld().request(), RequestContactInput{From: "user_ada", Email: "ada@example.com"}
}

// givenRequestContactRefusesAPairAlreadyLinked builds RequestContact for "refuses a pair already linked": the use case and the input it
// refuses with link.ErrAlready.
func givenRequestContactRefusesAPairAlreadyLinked(t *testing.T) (RequestContact, RequestContactInput) {
	t.Helper()
	w := newWorld()
	w.linked("user_bob", "user_ada", link.RequestsPending)
	return w.request(), RequestContactInput{From: "user_ada", Email: "bob@example.com"}
}
