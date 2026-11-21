// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"cmp"
	"context"
	"slices"

	"github.com/kitsunium/todo/identity/account"
)

// execute is ListSessions's body: the member's sessions that have not
// expired, the one last seen first.
func (h *ListSessionsHandler) execute(ctx context.Context, in *ListSessionsInput) (ListSessionsOutput, error) {
	all, err := h.sessions.Of(ctx, in.UserID)
	if err != nil {
		return ListSessionsOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return ListSessionsOutput{}, err
	}
	live := slices.DeleteFunc(slices.Clone(all), func(s account.Session) bool { return !now.Before(s.ExpiresAt) })
	slices.SortFunc(live, func(a, b account.Session) int {
		return cmp.Or(b.LastSeenAt.Compare(a.LastSeenAt), cmp.Compare(a.ID, b.ID))
	})
	return ListSessionsOutput{Sessions: live}, nil
}
