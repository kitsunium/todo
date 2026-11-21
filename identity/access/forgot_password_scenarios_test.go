// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"strings"
	"testing"

	"github.com/kitsunium/todo/identity/account"
)

func (w *world) forgot() ForgotPassword {
	return NewForgotPassword(accountsPort{w}, clockPort{w}, linksPort{w}, secretsPort{w}, courierPort{w})
}

// givenForgotPasswordMailsTheLinkToAMember builds ForgotPassword for "mails the link to a member": the use case, its input, and
// what its output must be.
func givenForgotPasswordMailsTheLinkToAMember(t *testing.T) (ForgotPassword, ForgotPasswordInput, func(*testing.T, ForgotPasswordOutput)) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	return w.forgot(), ForgotPasswordInput{Email: "ada@example.com"}, func(t *testing.T, _ ForgotPasswordOutput) {
		t.Helper()
		if len(w.mails) != 1 || !strings.HasPrefix(w.mails[0], "reset ada@example.com ") || len(w.links) != 1 {
			t.Fatalf("the mails: %q, %d link(s)", w.mails, len(w.links))
		}
		for _, l := range w.links {
			if l.UserID != a.ID || l.Purpose != account.PurposeReset || !l.ExpiresAt.Equal(epoch.Add(account.ResetFor)) {
				t.Errorf("the link: %+v", l)
			}
		}
	}
}

// givenForgotPasswordSaysNothingOfAnUnknownAddress builds ForgotPassword for "says nothing of an unknown address": the use case, its input, and
// what its output must be.
func givenForgotPasswordSaysNothingOfAnUnknownAddress(t *testing.T) (ForgotPassword, ForgotPasswordInput, func(*testing.T, ForgotPasswordOutput)) {
	t.Helper()
	w := newWorld()
	return w.forgot(), ForgotPasswordInput{Email: "nobody@example.com"}, func(t *testing.T, _ ForgotPasswordOutput) {
		t.Helper()
		if len(w.mails) != 0 || len(w.links) != 0 {
			t.Errorf("%d mail(s), %d link(s): want nothing", len(w.mails), len(w.links))
		}
	}
}
