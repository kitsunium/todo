// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"
	"errors"
	"strings"

	"github.com/kitsunium/todo/identity/account"
)

// execute is SignUp's body: the address, the name and the password are
// checked, then the account is kept with its password stretched, and the
// link that proves its address mailed. An address that already has an
// account is answered the same, and its member told someone tried: the
// answer never says whether an address has an account. Where the deployment
// trusts addresses, the account is verified at once, its address announced, and no link is mailed.
func (h *SignUpHandler) execute(ctx context.Context, in *SignUpInput) (SignUpOutput, error) {
	email, err := account.NormalizeEmail(in.Email)
	if err != nil {
		return SignUpOutput{}, err
	}
	name := strings.Join(strings.Fields(in.Name), " ")
	if name == "" {
		return SignUpOutput{}, account.ErrNameRequired
	}
	if !acceptable(in.Password) {
		return SignUpOutput{}, account.ErrWeakPassword
	}
	locale := in.Locale
	if locale == "" {
		locale = account.LocaleFR // nothing said, nothing negotiated
	}
	if locale, err = checkLocale(locale); err != nil {
		return SignUpOutput{}, err
	}
	zone, err := checkZone(in.TimeZone)
	if err != nil {
		return SignUpOutput{}, err
	}
	out := SignUpOutput{Email: email, Verified: in.Trusted}
	existing, taken, err := h.accounts.ByEmail(ctx, email)
	if err != nil {
		return SignUpOutput{}, err
	}
	hash, err := h.passwords.Hash(ctx, in.Password) // stretched either way: a taken address costs the same time
	if err != nil {
		return SignUpOutput{}, err
	}
	if taken {
		return out, h.courier.AlreadyMember(ctx, existing)
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return SignUpOutput{}, err
	}
	id, err := h.accounts.NewID(ctx)
	if err != nil {
		return SignUpOutput{}, err
	}
	a := account.Account{ID: id, Email: email, Name: name, Locale: locale, TimeZone: zone, PasswordHash: hash, Verified: in.Trusted, CreatedAt: now, UpdatedAt: now}
	if err := h.accounts.Save(ctx, a); errors.Is(err, account.ErrEmailTaken) {
		// Two sign-ups of one address at the same instant: the second is
		// answered like any taken address.
		winner, found, err := h.accounts.ByEmail(ctx, email)
		if err != nil || !found {
			return out, err
		}
		out.Verified = false
		return out, h.courier.AlreadyMember(ctx, winner)
	} else if err != nil {
		return SignUpOutput{}, err
	}
	if in.Trusted {
		return out, h.announcements.Verified(ctx, a)
	}
	secret, err := issue(ctx, h.links, h.secrets, a.ID, account.PurposeVerify, now, account.VerifyFor)
	if err != nil {
		return SignUpOutput{}, err
	}
	return out, h.courier.Verification(ctx, a, secret)
}
