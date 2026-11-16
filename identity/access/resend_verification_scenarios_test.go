// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"testing"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

func (w *world) resend() ResendVerification {
	return NewResendVerification(accountsPort{w}, clockPort{w}, linksPort{w}, secretsPort{w}, courierPort{w})
}

// givenResendVerificationMailsANewLinkToAnUnverifiedAccount builds ResendVerification for "mails a new link to an unverified account": the use case, its input, and
// what its output must be.
func givenResendVerificationMailsANewLinkToAnUnverifiedAccount(t *testing.T) (ResendVerification, ResendVerificationInput, func(*testing.T, ResendVerificationOutput)) {
	t.Helper()
	w, a := unverified(account.PurposeVerify, epoch.Add(time.Hour))
	return w.resend(), ResendVerificationInput{Email: " ADA@example.com"}, func(t *testing.T, _ ResendVerificationOutput) {
		t.Helper()
		if len(w.links) != 1 || len(w.mails) != 1 {
			t.Fatalf("%d link(s), %d mail(s): want the new link alone, mailed", len(w.links), len(w.mails))
		}
		for _, l := range w.links {
			if l.Hash == "digest:secret_a" || l.UserID != a.ID {
				t.Errorf("the link kept: %+v, want a new one", l)
			}
		}
	}
}

// givenResendVerificationSaysNothingOfAnUnknownAddress builds ResendVerification for "says nothing of an unknown address": the use case, its input, and
// what its output must be.
func givenResendVerificationSaysNothingOfAnUnknownAddress(t *testing.T) (ResendVerification, ResendVerificationInput, func(*testing.T, ResendVerificationOutput)) {
	t.Helper()
	w := newWorld()
	w.member("verified@example.com")
	uc := w.resend()
	return uc, ResendVerificationInput{Email: "nobody@example.com"}, func(t *testing.T, _ ResendVerificationOutput) {
		t.Helper()
		if _, err := uc.Execute(t.Context(), ResendVerificationInput{Email: "verified@example.com"}); err != nil {
			t.Errorf("a verified account: %v, want the same silence", err)
		}
		if len(w.mails) != 0 || len(w.links) != 0 {
			t.Errorf("%d mail(s), %d link(s): want nothing", len(w.mails), len(w.links))
		}
	}
}
