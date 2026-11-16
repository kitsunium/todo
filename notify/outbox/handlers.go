// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package outbox

import (
	"cmp"
	"context"
	"slices"
)

// kept is GET /api/outbox's body: every mail sent, the newest first.
func kept(ctx context.Context, in NoInput) (out Mails, err error) {
	all, err := OutboxStore.List(ctx)
	if err != nil {
		return Mails{}, err
	}
	slices.SortFunc(all, func(a, b Sent) int { return cmp.Or(b.At.Compare(a.At), cmp.Compare(b.ID, a.ID)) })
	return Mails{Mails: all}, nil
}
