// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"testing"

	"github.com/kitsunium/todo/contacts/link"
)

// givenListContactsListsContactsAndRequestsApart builds ListContacts for "lists contacts and requests apart": the use case, its input, and
// what its output must be.
func givenListContactsListsContactsAndRequestsApart(t *testing.T) (ListContacts, ListContactsInput, func(*testing.T, ListContactsOutput)) {
	t.Helper()
	w := newWorld()
	w.people["user_dee"] = Person{ID: "user_dee", Email: "dee@example.com", Name: "Dee"}
	w.people["user_eve"] = Person{ID: "user_eve", Email: "eve@example.com", Name: "eve"}
	w.linked("user_ada", "user_bob", link.RequestsAccepted)
	w.linked("user_cy", "user_ada", link.RequestsPending)
	w.linked("user_ada", "user_dee", link.RequestsPending)
	w.linked("user_ada", "user_eve", link.RequestsDeclined)
	w.linked("user_bob", "user_cy", link.RequestsAccepted)
	older, newer := w.invited("user_ada", "old@example.com"), w.invited("user_ada", "new@example.com")
	w.invited("user_bob", "other@example.com")            // another member's
	joined := w.invited("user_ada", "joined@example.com") // already joined
	joined.Status = link.InviteJoined
	w.invites[joined.ID] = joined
	return NewListContacts(w, w, w.invitations()), ListContactsInput{User: "user_ada"}, func(t *testing.T, out ListContactsOutput) {
		t.Helper()
		one := func(what string, list []Contact, want string) {
			if len(list) != 1 || list[0].User.ID != want {
				t.Errorf("%s: %+v, want %s alone", what, list, want)
			}
		}
		one("the contacts", out.Contacts, "user_bob")
		one("the requests received", out.Incoming, "user_cy")
		one("the requests sent", out.Outgoing, "user_dee")
		if len(out.Invites) != 2 || out.Invites[0].ID != newer.ID || out.Invites[1].ID != older.ID || out.Invites[0].Email != "new@example.com" {
			t.Errorf("the invitations: %+v, want the two waiting ones of Ada, the newest first", out.Invites)
		}
	}
}
