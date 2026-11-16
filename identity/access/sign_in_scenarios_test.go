// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"errors"
	"testing"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

// givenSignInOpensASession builds SignIn for "opens a session": the use case, its input, and
// what its output must be.
func givenSignInOpensASession(t *testing.T) (SignIn, SignInInput, func(*testing.T, SignInOutput)) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	a.FailedLogins = 3
	w.accounts[a.ID] = a
	return w.signIn(), SignInInput{Email: "Ada@example.com", Password: passphrase}, func(t *testing.T, out SignInOutput) {
		t.Helper()
		s := out.Session
		if out.Secret == "" || s.UserID != a.ID || s.TokenHash != "digest:"+out.Secret || !s.ExpiresAt.Equal(epoch.Add(account.SessionFor)) {
			t.Errorf("the session opened: %+v, secret %q", s, out.Secret)
		}
		if kept, ok := w.sessions[s.ID]; !ok || kept != s {
			t.Errorf("the store keeps %+v, want %+v", kept, s)
		}
		if w.accounts[a.ID].FailedLogins != 0 {
			t.Errorf("a right password leaves %d wrong ones counted", w.accounts[a.ID].FailedLogins)
		}
	}
}

// givenSignInRefusesAWrongPassword builds SignIn for "refuses a wrong password": the use case and the input it
// refuses with account.ErrInvalidCredentials.
func givenSignInRefusesAWrongPassword(t *testing.T) (SignIn, SignInInput) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	t.Cleanup(func() {
		if got := w.accounts[a.ID].FailedLogins; got != 1 {
			t.Errorf("a wrong password counts %d, want 1", got)
		}
	})
	return w.signIn(), SignInInput{Email: "ada@example.com", Password: "not the right one"}
}

// givenSignInRefusesAnUnknownAddress builds SignIn for "refuses an unknown address": the use case and the input it
// refuses with account.ErrInvalidCredentials.
func givenSignInRefusesAnUnknownAddress(t *testing.T) (SignIn, SignInInput) {
	t.Helper()
	return newWorld().signIn(), SignInInput{Email: "nobody@example.com", Password: passphrase}
}

// givenSignInRefusesALockedAccount builds SignIn for "refuses a locked account": the use case and the input it
// refuses with account.ErrLocked. The account is locked by five wrong
// passwords, here, and its lock ends LockFor later.
func givenSignInRefusesALockedAccount(t *testing.T) (SignIn, SignInInput) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	uc := w.signIn()
	for range account.MaxFailedLogins {
		if _, err := uc.Execute(t.Context(), SignInInput{Email: a.Email, Password: "wrong"}); !errors.Is(err, account.ErrInvalidCredentials) {
			t.Fatalf("a wrong password: %v, want ErrInvalidCredentials", err)
		}
	}
	t.Cleanup(func() {
		w.now = epoch.Add(account.LockFor + time.Second)
		if _, err := uc.Execute(t.Context(), SignInInput{Email: a.Email, Password: passphrase}); err != nil {
			t.Errorf("after the lock: %v, want a session", err)
		}
	})
	return uc, SignInInput{Email: a.Email, Password: passphrase}
}
