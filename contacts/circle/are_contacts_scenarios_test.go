// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"testing"

	"github.com/kitsunium/todo/contacts/link"
)

// givenAreContactsSaysYesForAnAcceptedLinkOnly builds AreContacts for "says yes for an accepted link only": the use case, its input, and
// what its output must be.
func givenAreContactsSaysYesForAnAcceptedLinkOnly(t *testing.T) (AreContacts, AreContactsInput, func(*testing.T, AreContactsOutput)) {
	t.Helper()
	w := newWorld()
	w.linked("user_ada", "user_bob", link.RequestsAccepted)
	w.linked("user_ada", "user_cy", link.RequestsPending)
	uc := NewAreContacts(w)
	return uc, AreContactsInput{A: "user_bob", B: "user_ada"}, func(t *testing.T, out AreContactsOutput) {
		t.Helper()
		if !out.Yes {
			t.Error("Ada and Bob are contacts, whatever the order they are named in")
		}
		if pending, err := uc.Execute(t.Context(), AreContactsInput{A: "user_ada", B: "user_cy"}); err != nil || pending.Yes {
			t.Errorf("a pending request makes contacts: %v (%v)", pending.Yes, err)
		}
	}
}
