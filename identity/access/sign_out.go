// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"
)

// execute is SignOut's body: the session is deleted, and its secret opens
// nothing any more.
func (h *SignOutHandler) execute(ctx context.Context, in *SignOutInput) (SignOutOutput, error) {
	return SignOutOutput{}, h.sessions.Delete(ctx, in.SessionID)
}
