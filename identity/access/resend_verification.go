// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"

	"github.com/kitsunium/todo/identity/account"
)

// execute is ResendVerification's body: an account not verified yet gets a
// new link, in place of the one it had. An address with no account, or one
// already verified, is answered the same and gets nothing.
func (h *ResendVerificationHandler) execute(ctx context.Context, in *ResendVerificationInput) (ResendVerificationOutput, error) {
	email, err := account.NormalizeEmail(in.Email)
	if err != nil {
		return ResendVerificationOutput{}, nil
	}
	a, found, err := h.accounts.ByEmail(ctx, email)
	if err != nil || !found || a.Verified {
		return ResendVerificationOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return ResendVerificationOutput{}, err
	}
	secret, err := issue(ctx, h.links, h.secrets, a.ID, account.PurposeVerify, now, account.VerifyFor)
	if err != nil {
		return ResendVerificationOutput{}, err
	}
	return ResendVerificationOutput{}, h.courier.Verification(ctx, a, secret)
}
