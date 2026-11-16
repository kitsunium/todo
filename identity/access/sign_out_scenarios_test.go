// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"testing"
)

// givenSignOutEndsTheSession builds SignOut for "ends the session": the use case, its input, and
// what its output must be.
func givenSignOutEndsTheSession(t *testing.T) (SignOut, SignOutInput, func(*testing.T, SignOutOutput)) {
	t.Helper()
	w := newWorld()
	s := w.session(w.member("ada@example.com"), "secret_a", epoch)
	return NewSignOut(sessionsPort{w}), SignOutInput{SessionID: s.ID}, func(t *testing.T, _ SignOutOutput) {
		t.Helper()
		if len(w.sessions) != 0 {
			t.Errorf("the session is still kept: %+v", w.sessions)
		}
	}
}
