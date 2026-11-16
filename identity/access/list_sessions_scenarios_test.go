// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"testing"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

// givenListSessionsListsTheLiveSessionsOfTheMember builds ListSessions for "lists the live sessions of the member": the use case, its input, and
// what its output must be.
func givenListSessionsListsTheLiveSessionsOfTheMember(t *testing.T) (ListSessions, ListSessionsInput, func(*testing.T, ListSessionsOutput)) {
	t.Helper()
	w := newWorld()
	ada, bob := w.member("ada@example.com"), w.member("bob@example.com")
	old := w.session(ada, "secret_old", epoch)
	recent := w.session(ada, "secret_recent", epoch.Add(time.Hour))
	w.session(ada, "secret_expired", epoch.Add(-account.SessionFor))
	w.session(bob, "secret_bob", epoch)
	w.now = epoch.Add(2 * time.Hour)
	return NewListSessions(sessionsPort{w}, clockPort{w}), ListSessionsInput{UserID: ada.ID}, func(t *testing.T, out ListSessionsOutput) {
		t.Helper()
		if len(out.Sessions) != 2 || out.Sessions[0].ID != recent.ID || out.Sessions[1].ID != old.ID {
			t.Errorf("the sessions: %+v, want the recent one then the old one", out.Sessions)
		}
	}
}
