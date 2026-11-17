// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"

	"github.com/kitsunium/todo/identity/account"
)

// execute is ChangePassword's body: a member who gives their current
// password takes a new one, and their other sessions end; the one the
// request runs in stays.
func (h *ChangePasswordHandler) execute(ctx context.Context, in *ChangePasswordInput) (ChangePasswordOutput, error) {
	if !acceptable(in.Password) {
		return ChangePasswordOutput{}, account.ErrWeakPassword
	}
	a, found, err := h.accounts.Get(ctx, in.UserID)
	if err != nil {
		return ChangePasswordOutput{}, err
	}
	if !found {
		return ChangePasswordOutput{}, account.ErrInvalidCredentials
	}
	ok, err := h.passwords.Verify(ctx, in.Current, a.PasswordHash)
	if err != nil {
		return ChangePasswordOutput{}, err
	}
	if !ok {
		return ChangePasswordOutput{}, account.ErrInvalidCredentials
	}
	hash, err := h.passwords.Hash(ctx, in.Password)
	if err != nil {
		return ChangePasswordOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return ChangePasswordOutput{}, err
	}
	// Written to the account as it is at that instant: what another request
	// changed meanwhile stays, and a password that changed meanwhile is no
	// longer the one that was proven.
	proven := a.PasswordHash
	if _, err := h.accounts.Change(ctx, a.ID, func(a *account.Account) error {
		if a.PasswordHash != proven {
			return account.ErrInvalidCredentials
		}
		a.PasswordHash, a.UpdatedAt = hash, now
		return nil
	}); err != nil {
		return ChangePasswordOutput{}, err
	}
	return ChangePasswordOutput{}, h.sessions.DeleteOf(ctx, a.ID, in.SessionID)
}
