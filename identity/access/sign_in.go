// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

// execute is SignIn's body. An unknown address and a wrong password are
// answered alike, and an unknown address still has a password checked, so
// neither the answer nor its time says whether an address has an account.
// The fifth wrong password in a row locks the account for LockFor; a right
// one clears the count.
func (h *SignInHandler) execute(ctx context.Context, in *SignInInput) (SignInOutput, error) {
	email, err := account.NormalizeEmail(in.Email)
	if err != nil {
		email = "" // no account has it: answered as an unknown address
	}
	a, found, err := h.accounts.ByEmail(ctx, email)
	if err != nil {
		return SignInOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return SignInOutput{}, err
	}
	if !found {
		_, _ = h.passwords.Verify(ctx, in.Password, "")
		return SignInOutput{}, account.ErrInvalidCredentials
	}
	if a.Locked(now) {
		return SignInOutput{}, account.ErrLocked
	}
	ok, err := h.passwords.Verify(ctx, in.Password, a.PasswordHash)
	if err != nil {
		return SignInOutput{}, err
	}
	if !ok {
		a.FailedLogins++
		if a.FailedLogins >= account.MaxFailedLogins {
			a.FailedLogins, a.LockedUntil = 0, now.Add(account.LockFor)
		}
		if err := h.accounts.Save(ctx, a); err != nil {
			return SignInOutput{}, err
		}
		return SignInOutput{}, account.ErrInvalidCredentials
	}
	if a.FailedLogins > 0 || !a.LockedUntil.IsZero() {
		a.FailedLogins, a.LockedUntil = 0, time.Time{}
		if err := h.accounts.Save(ctx, a); err != nil {
			return SignInOutput{}, err
		}
	}
	secret, digest, err := h.secrets.New(ctx)
	if err != nil {
		return SignInOutput{}, err
	}
	id, err := h.sessions.NewID(ctx)
	if err != nil {
		return SignInOutput{}, err
	}
	s := account.Session{ID: id, UserID: a.ID, TokenHash: digest, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(account.SessionFor)}
	if err := h.sessions.Save(ctx, s); err != nil {
		return SignInOutput{}, err
	}
	return SignInOutput{Secret: secret, Session: s, Account: a}, nil
}
