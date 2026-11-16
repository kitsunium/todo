// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

// slideEvery is how often a session seen again has its end pushed back: a
// member who keeps coming stays signed in, without a write per request.
const slideEvery = 5 * time.Minute

// execute is Authenticate's body: the session whose secret digests to what
// is kept, if it has not expired, and its member. A session not seen for
// slideEvery is renewed for SessionFor, and the answer says so.
func (h *AuthenticateHandler) execute(ctx context.Context, in *AuthenticateInput) (AuthenticateOutput, error) {
	digest, err := h.secrets.Digest(ctx, in.Secret)
	if err != nil {
		return AuthenticateOutput{}, err
	}
	s, found, err := h.sessions.ByTokenHash(ctx, digest)
	if err != nil {
		return AuthenticateOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return AuthenticateOutput{}, err
	}
	if !found || !now.Before(s.ExpiresAt) {
		return AuthenticateOutput{}, account.ErrNoSession
	}
	renewed := now.Sub(s.LastSeenAt) >= slideEvery
	if renewed {
		s.LastSeenAt, s.ExpiresAt = now, now.Add(account.SessionFor)
		if err := h.sessions.Save(ctx, s); err != nil {
			return AuthenticateOutput{}, err
		}
	}
	a, found, err := h.accounts.Get(ctx, s.UserID)
	if err != nil {
		return AuthenticateOutput{}, err
	}
	if !found {
		return AuthenticateOutput{}, account.ErrNoSession
	}
	return AuthenticateOutput{Session: s, Account: a, Renewed: renewed}, nil
}
