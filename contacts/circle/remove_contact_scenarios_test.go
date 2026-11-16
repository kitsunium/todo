// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"testing"

	"github.com/kitsunium/todo/contacts/link"
)

// givenRemoveContactEndsAContact builds RemoveContact for "ends a contact": the use case, its input, and
// what its output must be.
func givenRemoveContactEndsAContact(t *testing.T) (RemoveContact, RemoveContactInput, func(*testing.T, RemoveContactOutput)) {
	t.Helper()
	w := newWorld()
	l := w.linked("user_ada", "user_bob", link.RequestsAccepted)
	return NewRemoveContact(w), RemoveContactInput{User: "user_ada", ID: l.ID}, func(t *testing.T, _ RemoveContactOutput) {
		t.Helper()
		if len(w.links) != 0 {
			t.Errorf("the link is still kept: %+v", w.links)
		}
	}
}

// givenRemoveContactRefusesAContactOfOthers builds RemoveContact for "refuses a contact of others": the use case and the input it
// refuses with link.ErrNotFound.
func givenRemoveContactRefusesAContactOfOthers(t *testing.T) (RemoveContact, RemoveContactInput) {
	t.Helper()
	w := newWorld()
	l := w.linked("user_ada", "user_bob", link.RequestsAccepted)
	return NewRemoveContact(w), RemoveContactInput{User: "user_cy", ID: l.ID}
}
