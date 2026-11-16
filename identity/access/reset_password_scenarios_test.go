// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"testing"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

func (w *world) reset() ResetPassword {
	return NewResetPassword(secretsPort{w}, linksPort{w}, clockPort{w}, accountsPort{w}, passwordsPort{w}, sessionsPort{w}, announcementsPort{w})
}

// givenResetPasswordChoosesANewPassword builds ResetPassword for "chooses a new password": the use case, its input, and
// what its output must be.
func givenResetPasswordChoosesANewPassword(t *testing.T) (ResetPassword, ResetPasswordInput, func(*testing.T, ResetPasswordOutput)) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	a.Verified, a.LockedUntil = false, epoch.Add(time.Hour)
	w.accounts[a.ID] = a
	w.session(a, "secret_s", epoch)
	w.link(a, account.PurposeReset, "secret_a", epoch.Add(time.Hour))
	return w.reset(), ResetPasswordInput{Secret: "secret_a", Password: "a brand new passphrase"}, func(t *testing.T, _ ResetPasswordOutput) {
		t.Helper()
		got := w.accounts[a.ID]
		if got.PasswordHash != "hash:a brand new passphrase" || !got.Verified || got.Locked(w.now) {
			t.Errorf("the account: %+v, want the new password, verified, unlocked", got)
		}
		if len(w.links) != 0 || len(w.sessions) != 0 {
			t.Errorf("%d link(s) and %d session(s) left, want none", len(w.links), len(w.sessions))
		}
		if len(w.heard) != 1 || w.heard[0] != "verified ada@example.com" {
			t.Errorf("announced %q, want the address proven, once", w.heard)
		}
	}
}

// givenResetPasswordRefusesALinkUsedOnce builds ResetPassword for "refuses a link used once": the use case and the input it
// refuses with account.ErrBadLink.
func givenResetPasswordRefusesALinkUsedOnce(t *testing.T) (ResetPassword, ResetPasswordInput) {
	t.Helper()
	w := newWorld()
	w.link(w.member("ada@example.com"), account.PurposeReset, "secret_a", epoch.Add(time.Hour))
	uc, in := w.reset(), ResetPasswordInput{Secret: "secret_a", Password: "a brand new passphrase"}
	if _, err := uc.Execute(t.Context(), in); err != nil {
		t.Fatalf("the first use: %v", err)
	}
	return uc, in
}

// givenResetPasswordRefusesAShortPassword builds ResetPassword for "refuses a short password": the use case and the input it
// refuses with account.ErrWeakPassword.
func givenResetPasswordRefusesAShortPassword(t *testing.T) (ResetPassword, ResetPasswordInput) {
	t.Helper()
	w := newWorld()
	w.link(w.member("ada@example.com"), account.PurposeReset, "secret_a", epoch.Add(time.Hour))
	return w.reset(), ResetPasswordInput{Secret: "secret_a", Password: "short"}
}

// givenResetPasswordAnnouncesNothingForAnAddressAlreadyProven builds ResetPassword for "announces nothing for an address already proven": the use case, its input, and
// what its output must be.
func givenResetPasswordAnnouncesNothingForAnAddressAlreadyProven(t *testing.T) (ResetPassword, ResetPasswordInput, func(*testing.T, ResetPasswordOutput)) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com") // verified already
	w.link(a, account.PurposeReset, "secret_a", epoch.Add(time.Hour))
	return w.reset(), ResetPasswordInput{Secret: "secret_a", Password: "a brand new passphrase"}, func(t *testing.T, _ ResetPasswordOutput) {
		t.Helper()
		if len(w.heard) != 0 || w.accounts[a.ID].PasswordHash != "hash:a brand new passphrase" {
			t.Errorf("announced %q, account %+v: want nothing announced and the new password", w.heard, w.accounts[a.ID])
		}
	}
}
