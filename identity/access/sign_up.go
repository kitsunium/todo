// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/kitsunium/todo/identity/account"
)

// maxPassword is the longest password an account takes: stretching one
// costs time in proportion to nothing a caller should choose.
const maxPassword = 128

// execute is SignUp's body: the address, the name and the password are
// checked, then the account is kept with its password stretched.
func (h *SignUpHandler) execute(ctx context.Context, in *SignUpInput) (SignUpOutput, error) {
	email, err := account.NormalizeEmail(in.Email)
	if err != nil {
		return SignUpOutput{}, err
	}
	name := strings.Join(strings.Fields(in.Name), " ")
	if name == "" {
		return SignUpOutput{}, account.ErrNameRequired
	}
	if n := utf8.RuneCountInString(in.Password); n < account.MinPassword || n > maxPassword {
		return SignUpOutput{}, account.ErrWeakPassword
	}
	if _, taken, err := h.accounts.ByEmail(ctx, email); err != nil {
		return SignUpOutput{}, err
	} else if taken {
		return SignUpOutput{}, account.ErrEmailTaken
	}
	hash, err := h.passwords.Hash(ctx, in.Password)
	if err != nil {
		return SignUpOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return SignUpOutput{}, err
	}
	id, err := h.accounts.NewID(ctx)
	if err != nil {
		return SignUpOutput{}, err
	}
	a := account.Account{ID: id, Email: email, Name: name, PasswordHash: hash, CreatedAt: now, UpdatedAt: now}
	if err := h.accounts.Save(ctx, a); err != nil {
		return SignUpOutput{}, err
	}
	return SignUpOutput{Account: a}, nil
}
