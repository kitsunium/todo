// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"testing"
)

// givenSignUpCreatesAnAccount builds SignUp for "creates an account": the use case, its input, and
// what its output must be.
func givenSignUpCreatesAnAccount(t *testing.T) (SignUp, SignUpInput, func(*testing.T, SignUpOutput)) {
	t.Helper()
	w := newWorld()
	in := SignUpInput{Email: "  Ada@Example.com ", Name: " Ada   Lovelace ", Password: passphrase}
	return w.signUp(), in, func(t *testing.T, out SignUpOutput) {
		t.Helper()
		a := out.Account
		if a.ID == "" || a.Email != "ada@example.com" || a.Name != "Ada Lovelace" || !a.CreatedAt.Equal(epoch) {
			t.Errorf("the account created: %+v", a)
		}
		if a.PasswordHash == "" || a.PasswordHash == passphrase {
			t.Errorf("the password is kept as %q: it must be stretched", a.PasswordHash)
		}
		if kept, ok := w.accounts[a.ID]; !ok || kept != a {
			t.Errorf("the store keeps %+v, want %+v", kept, a)
		}
	}
}

// givenSignUpRefusesAnInvalidAddress builds SignUp for "refuses an invalid address": the use case and the input it
// refuses with account.ErrInvalidEmail.
func givenSignUpRefusesAnInvalidAddress(t *testing.T) (SignUp, SignUpInput) {
	t.Helper()
	return newWorld().signUp(), SignUpInput{Email: "ada at example.com", Name: "Ada", Password: passphrase}
}

// givenSignUpRefusesAShortPassword builds SignUp for "refuses a short password": the use case and the input it
// refuses with account.ErrWeakPassword.
func givenSignUpRefusesAShortPassword(t *testing.T) (SignUp, SignUpInput) {
	t.Helper()
	return newWorld().signUp(), SignUpInput{Email: "ada@example.com", Name: "Ada", Password: "too short"}
}

// givenSignUpRefusesATakenAddress builds SignUp for "refuses a taken address": the use case and the input it
// refuses with account.ErrEmailTaken.
func givenSignUpRefusesATakenAddress(t *testing.T) (SignUp, SignUpInput) {
	t.Helper()
	w := newWorld()
	w.member("ada@example.com")
	return w.signUp(), SignUpInput{Email: "ADA@example.com", Name: "Another Ada", Password: passphrase}
}
