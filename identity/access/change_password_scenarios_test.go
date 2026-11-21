// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"testing"
)

func (w *world) change() ChangePassword {
	return NewChangePassword(accountsPort{w}, passwordsPort{w}, clockPort{w}, sessionsPort{w})
}

// givenChangePasswordReplacesThePassword builds ChangePassword for "replaces the password": the use case, its input, and
// what its output must be.
func givenChangePasswordReplacesThePassword(t *testing.T) (ChangePassword, ChangePasswordInput, func(*testing.T, ChangePasswordOutput)) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	here, there := w.session(a, "secret_here", epoch), w.session(a, "secret_there", epoch)
	in := ChangePasswordInput{UserID: a.ID, SessionID: here.ID, Current: passphrase, Password: "a brand new passphrase"}
	return w.change(), in, func(t *testing.T, _ ChangePasswordOutput) {
		t.Helper()
		if w.accounts[a.ID].PasswordHash != "hash:a brand new passphrase" {
			t.Errorf("the password kept: %q", w.accounts[a.ID].PasswordHash)
		}
		if _, kept := w.sessions[here.ID]; !kept || len(w.sessions) != 1 {
			t.Errorf("the sessions left: %+v, want the one of the request, not %s", w.sessions, there.ID)
		}
	}
}

// givenChangePasswordRefusesAWrongCurrentPassword builds ChangePassword for "refuses a wrong current password": the use case and the input it
// refuses with account.ErrInvalidCredentials.
func givenChangePasswordRefusesAWrongCurrentPassword(t *testing.T) (ChangePassword, ChangePasswordInput) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	return w.change(), ChangePasswordInput{UserID: a.ID, Current: "not the right one", Password: "a brand new passphrase"}
}
