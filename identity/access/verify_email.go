// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"

	"github.com/kitsunium/todo/identity/account"
)

// execute is VerifyEmail's body: the link's account is marked verified, and
// the link is spent.
func (h *VerifyEmailHandler) execute(ctx context.Context, in *VerifyEmailInput) (VerifyEmailOutput, error) {
	l, now, err := redeem(ctx, h.links, h.secrets, h.clock, in.Secret, account.PurposeVerify)
	if err != nil {
		return VerifyEmailOutput{}, err
	}
	a, found, err := h.accounts.Get(ctx, l.UserID)
	if err != nil {
		return VerifyEmailOutput{}, err
	}
	if !found {
		return VerifyEmailOutput{}, account.ErrBadLink
	}
	a.Verified, a.UpdatedAt = true, now
	if err := h.accounts.Save(ctx, a); err != nil {
		return VerifyEmailOutput{}, err
	}
	return VerifyEmailOutput{Account: a}, h.links.DeleteOf(ctx, a.ID, account.PurposeVerify)
}
