// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package outbox

import (
	"cmp"
	"context"
	"slices"

	"github.com/kitsunium/sdk/framework/app"
)

// kept is GET /api/outbox's body: every mail sent on the caller's behalf, the newest first.
func kept(ctx context.Context, in NoInput) (out Mails, err error) {
	uid, ok := app.UserID(ctx)
	if !ok {
		return Mails{}, app.Unauthenticated("sign in first")
	}
	all, err := OutboxStore.Filter(ctx, func(m Sent) bool { return m.UserID == string(uid) })
	if err != nil {
		return Mails{}, err
	}
	slices.SortFunc(all, func(a, b Sent) int { return cmp.Or(b.At.Compare(a.At), cmp.Compare(b.ID, a.ID)) })
	return Mails{Mails: all}, nil
}
