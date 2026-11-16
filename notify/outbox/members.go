// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package outbox

import (
	"context"

	"github.com/kitsunium/todo/identity/vault"
)

// members is Members's own state: write it freely.
type members struct {
}

// newMembers is NewMembers's body.
func newMembers() *Members {
	return &Members{members{}}
}

// address is Members.Address's body.
func (a *Members) address(ctx context.Context, userID string) (r0 string, r1 bool, err error) {
	m, err := vault.MemberQuery.Ask(ctx, vault.ByID{ID: userID})
	if err != nil {
		return "", false, err
	}
	return m.Email, m.Found, nil
}
