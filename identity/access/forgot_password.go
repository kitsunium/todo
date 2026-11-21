// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"

	"github.com/kitsunium/todo/identity/account"
)

// execute is ForgotPassword's body: the member of an address is mailed the
// link that chooses a new password, in place of the one they had. An
// address with no account is answered the same and gets nothing.
func (h *ForgotPasswordHandler) execute(ctx context.Context, in *ForgotPasswordInput) (ForgotPasswordOutput, error) {
	email, err := account.NormalizeEmail(in.Email)
	if err != nil {
		return ForgotPasswordOutput{}, nil
	}
	a, found, err := h.accounts.ByEmail(ctx, email)
	if err != nil || !found {
		return ForgotPasswordOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return ForgotPasswordOutput{}, err
	}
	secret, err := issue(ctx, h.links, h.secrets, a.ID, account.PurposeReset, now, account.ResetFor)
	if err != nil {
		return ForgotPasswordOutput{}, err
	}
	return ForgotPasswordOutput{}, h.courier.Reset(ctx, a, secret)
}
