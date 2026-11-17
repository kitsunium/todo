// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"
	"strings"

	"github.com/kitsunium/todo/identity/account"
)

// execute is UpdateProfile's body: what is given is checked, then written to
// the account as it is at that instant; what is not given stays. A zone given
// empty is UTC.
func (h *UpdateProfileHandler) execute(ctx context.Context, in *UpdateProfileInput) (UpdateProfileOutput, error) {
	var name, locale, zone string
	var err error
	if in.Name != nil {
		if name = strings.Join(strings.Fields(*in.Name), " "); name == "" {
			return UpdateProfileOutput{}, account.ErrNameRequired
		}
	}
	if in.Locale != nil {
		if locale, err = checkLocale(*in.Locale); err != nil {
			return UpdateProfileOutput{}, err
		}
	}
	if in.TimeZone != nil {
		if zone, err = checkZone(*in.TimeZone); err != nil {
			return UpdateProfileOutput{}, err
		}
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return UpdateProfileOutput{}, err
	}
	a, err := h.accounts.Change(ctx, in.UserID, func(a *account.Account) error {
		if in.Name != nil {
			a.Name = name
		}
		if in.Locale != nil {
			a.Locale = locale
		}
		if in.TimeZone != nil {
			a.TimeZone = zone
		}
		a.UpdatedAt = now
		return nil
	})
	if err != nil {
		return UpdateProfileOutput{}, err
	}
	return UpdateProfileOutput{Account: a}, nil
}
