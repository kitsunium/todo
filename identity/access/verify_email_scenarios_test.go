// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"testing"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

// unverified is a world with one account still to verify, and its link.
func unverified(purpose string, expires time.Time) (*world, account.Account) {
	w := newWorld()
	a := w.member("ada@example.com")
	a.Verified = false
	w.accounts[a.ID] = a
	w.link(a, purpose, "secret_a", expires)
	return w, a
}

func (w *world) verify() VerifyEmail {
	return NewVerifyEmail(secretsPort{w}, linksPort{w}, clockPort{w}, accountsPort{w}, sessionsPort{w})
}

// givenVerifyEmailVerifiesTheAccountOfALink builds VerifyEmail for "verifies the account of a link": the use case, its input, and
// what its output must be.
func givenVerifyEmailVerifiesTheAccountOfALink(t *testing.T) (VerifyEmail, VerifyEmailInput, func(*testing.T, VerifyEmailOutput)) {
	t.Helper()
	w, a := unverified(account.PurposeVerify, epoch.Add(time.Hour))
	uc := w.verify()
	return uc, VerifyEmailInput{Secret: "secret_a"}, func(t *testing.T, out VerifyEmailOutput) {
		t.Helper()
		if out.Account.ID != a.ID || !w.accounts[a.ID].Verified || len(w.links) != 0 {
			t.Errorf("the account: %+v, the links left: %d", w.accounts[a.ID], len(w.links))
		}
		if kept := w.sessions[out.Session.ID]; out.SessionSecret == "" || kept.UserID != a.ID || kept.TokenHash != "digest:"+out.SessionSecret {
			t.Errorf("the session opened: %+v, secret %q", kept, out.SessionSecret)
		}
		if _, err := uc.Execute(t.Context(), VerifyEmailInput{Secret: "secret_a"}); err == nil {
			t.Error("the link works twice")
		}
	}
}

// givenVerifyEmailRefusesAnExpiredLink builds VerifyEmail for "refuses an expired link": the use case and the input it
// refuses with account.ErrBadLink.
func givenVerifyEmailRefusesAnExpiredLink(t *testing.T) (VerifyEmail, VerifyEmailInput) {
	t.Helper()
	w, _ := unverified(account.PurposeVerify, epoch)
	return w.verify(), VerifyEmailInput{Secret: "secret_a"}
}

// givenVerifyEmailRefusesALinkOfAnotherPurpose builds VerifyEmail for "refuses a link of another purpose": the use case and the input it
// refuses with account.ErrBadLink.
func givenVerifyEmailRefusesALinkOfAnotherPurpose(t *testing.T) (VerifyEmail, VerifyEmailInput) {
	t.Helper()
	w, _ := unverified(account.PurposeReset, epoch.Add(time.Hour))
	return w.verify(), VerifyEmailInput{Secret: "secret_a"}
}
