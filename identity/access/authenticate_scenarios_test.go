// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"testing"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

// givenAuthenticateNamesTheMemberOfALiveSession builds Authenticate for "names the member of a live session": the use case, its input, and
// what its output must be.
func givenAuthenticateNamesTheMemberOfALiveSession(t *testing.T) (Authenticate, AuthenticateInput, func(*testing.T, AuthenticateOutput)) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	s := w.session(a, "secret_a", epoch)
	w.now = epoch.Add(time.Hour)
	uc := w.authenticate()
	return uc, AuthenticateInput{Secret: "secret_a"}, func(t *testing.T, out AuthenticateOutput) {
		t.Helper()
		if out.Account.ID != a.ID || out.Session.ID != s.ID {
			t.Errorf("the member named: %+v in %+v", out.Account, out.Session)
		}
		if !out.Renewed || !w.sessions[s.ID].ExpiresAt.Equal(w.now.Add(account.SessionFor)) {
			t.Errorf("a session seen an hour later is not renewed: %+v", w.sessions[s.ID])
		}
		again, err := uc.Execute(t.Context(), AuthenticateInput{Secret: "secret_a"})
		if err != nil || again.Renewed {
			t.Errorf("seen again at once: renewed %v (%v), want no write", again.Renewed, err)
		}
	}
}

// givenAuthenticateRefusesAnExpiredSession builds Authenticate for "refuses an expired session": the use case and the input it
// refuses with account.ErrNoSession.
func givenAuthenticateRefusesAnExpiredSession(t *testing.T) (Authenticate, AuthenticateInput) {
	t.Helper()
	w := newWorld()
	w.session(w.member("ada@example.com"), "secret_a", epoch)
	w.now = epoch.Add(account.SessionFor)
	return w.authenticate(), AuthenticateInput{Secret: "secret_a"}
}

// givenAuthenticateRefusesAnUnknownSecret builds Authenticate for "refuses an unknown secret": the use case and the input it
// refuses with account.ErrNoSession.
func givenAuthenticateRefusesAnUnknownSecret(t *testing.T) (Authenticate, AuthenticateInput) {
	t.Helper()
	w := newWorld()
	w.session(w.member("ada@example.com"), "secret_a", epoch)
	return w.authenticate(), AuthenticateInput{Secret: "secret_b"}
}
