// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"testing"
)

// givenRevokeSessionEndsASessionOfTheMember builds RevokeSession for "ends a session of the member": the use case, its input, and
// what its output must be.
func givenRevokeSessionEndsASessionOfTheMember(t *testing.T) (RevokeSession, RevokeSessionInput, func(*testing.T, RevokeSessionOutput)) {
	t.Helper()
	w := newWorld()
	ada := w.member("ada@example.com")
	gone, kept := w.session(ada, "secret_a", epoch), w.session(ada, "secret_b", epoch)
	return NewRevokeSession(sessionsPort{w}), RevokeSessionInput{UserID: ada.ID, ID: gone.ID}, func(t *testing.T, _ RevokeSessionOutput) {
		t.Helper()
		if _, still := w.sessions[gone.ID]; still || len(w.sessions) != 1 || w.sessions[kept.ID].ID != kept.ID {
			t.Errorf("the sessions left: %+v, want the other one alone", w.sessions)
		}
	}
}

// givenRevokeSessionRefusesASessionOfAnotherMember builds RevokeSession for "refuses a session of another member": the use case and the input it
// refuses with account.ErrNoSession.
func givenRevokeSessionRefusesASessionOfAnotherMember(t *testing.T) (RevokeSession, RevokeSessionInput) {
	t.Helper()
	w := newWorld()
	theirs := w.session(w.member("bob@example.com"), "secret_b", epoch)
	return NewRevokeSession(sessionsPort{w}), RevokeSessionInput{UserID: w.member("ada@example.com").ID, ID: theirs.ID}
}
