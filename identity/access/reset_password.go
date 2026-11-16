// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

// execute is ResetPassword's body: the link's account takes the new
// password. Its lock is lifted, its address is proven (the link was in its
// inbox), the link is spent and every session of the account ends: whoever
// held one signs in again.
func (h *ResetPasswordHandler) execute(ctx context.Context, in *ResetPasswordInput) (ResetPasswordOutput, error) {
	if !acceptable(in.Password) {
		return ResetPasswordOutput{}, account.ErrWeakPassword
	}
	l, now, err := redeem(ctx, h.links, h.secrets, h.clock, in.Secret, account.PurposeReset)
	if err != nil {
		return ResetPasswordOutput{}, err
	}
	a, found, err := h.accounts.Get(ctx, l.UserID)
	if err != nil {
		return ResetPasswordOutput{}, err
	}
	if !found {
		return ResetPasswordOutput{}, account.ErrBadLink
	}
	hash, err := h.passwords.Hash(ctx, in.Password)
	if err != nil {
		return ResetPasswordOutput{}, err
	}
	a.PasswordHash, a.Verified, a.FailedLogins, a.LockedUntil, a.UpdatedAt = hash, true, 0, time.Time{}, now
	if err := h.accounts.Save(ctx, a); err != nil {
		return ResetPasswordOutput{}, err
	}
	if err := h.links.DeleteOf(ctx, a.ID, account.PurposeReset); err != nil {
		return ResetPasswordOutput{}, err
	}
	return ResetPasswordOutput{}, h.sessions.DeleteOf(ctx, a.ID, "")
}
