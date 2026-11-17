// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package outbox

import (
	"context"

	"github.com/kitsunium/todo/identity/vault"
	"github.com/kitsunium/todo/notify/reminders"
)

// members is Members's own state: write it freely.
type members struct {
}

// newMembers is NewMembers's body.
func newMembers() *Members {
	return &Members{members{}}
}

// recipient is Members.Recipient's body: who a member is written to, from the
// identity's query "member".
func (a *Members) recipient(ctx context.Context, userID string) (r0 reminders.Recipient, r1 bool, err error) {
	m, err := vault.MemberQuery.Ask(ctx, vault.ByID{ID: userID})
	if err != nil || !m.Found {
		return reminders.Recipient{}, false, err
	}
	return reminders.Recipient{Address: m.Email, Name: m.Name, Locale: m.Locale, TimeZone: m.TimeZone}, true, nil
}
