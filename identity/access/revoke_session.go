// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"

	"github.com/kitsunium/todo/identity/account"
)

// execute is RevokeSession's body: a session is ended by the member it
// belongs to; anyone else's is not found.
func (h *RevokeSessionHandler) execute(ctx context.Context, in *RevokeSessionInput) (RevokeSessionOutput, error) {
	mine, err := h.sessions.Of(ctx, in.UserID)
	if err != nil {
		return RevokeSessionOutput{}, err
	}
	for _, s := range mine {
		if s.ID == in.ID {
			return RevokeSessionOutput{}, h.sessions.Delete(ctx, s.ID)
		}
	}
	return RevokeSessionOutput{}, account.ErrNoSession
}
