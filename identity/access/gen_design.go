// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:620afbb2962f72a1a0c57e91267127ee1982802bbd70150e8fa1d8268bc9f808

// Package access — Signing up, signing in and out, and naming the member a
// session belongs to.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package access

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/identity/account"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 6, 2, 255), "NOT_IMPLEMENTED", "not implemented yet", "access: a hand-written body is still a skeleton")

// AccountsV1 keeps the accounts.
type AccountsV1 interface {
	ByEmail(ctx context.Context, email string) (account.Account, bool, error)
	Get(ctx context.Context, id string) (account.Account, bool, error)
	NewID(ctx context.Context) (string, error)
	Save(ctx context.Context, a account.Account) error
}

// ClockV1 tells the time.
type ClockV1 interface {
	Now(ctx context.Context) (time.Time, error)
}

// PasswordsV1 stretches a password and checks one against what was kept.
type PasswordsV1 interface {
	Hash(ctx context.Context, password string) (string, error)
	Verify(ctx context.Context, password string, hash string) (bool, error)
}

// SecretsV1 makes a session's secret and digests one.
type SecretsV1 interface {
	Digest(ctx context.Context, secret string) (string, error)
	New(ctx context.Context) (string, string, error)
}

// SessionsV1 keeps the sessions.
type SessionsV1 interface {
	ByTokenHash(ctx context.Context, hash string) (account.Session, bool, error)
	Delete(ctx context.Context, id string) error
	NewID(ctx context.Context) (string, error)
	Save(ctx context.Context, s account.Session) error
}

// Authenticate names the member a session's secret belongs to, and keeps the
// session alive.
type Authenticate interface {
	Execute(ctx context.Context, in AuthenticateInput) (AuthenticateOutput, error)
}

// AuthenticateInput is Authenticate's input: a field added breaks no caller.
type AuthenticateInput struct {
	Secret string
}

// AuthenticateOutput is Authenticate's output.
type AuthenticateOutput struct {
	Session account.Session
	Account account.Account
	Renewed bool
}

// AuthenticateHandler runs Authenticate: its body is the hand-written execute.
type AuthenticateHandler struct {
	secrets  SecretsV1
	sessions SessionsV1
	clock    ClockV1
	accounts AccountsV1
}

// NewAuthenticate builds Authenticate on what it calls.
func NewAuthenticate(secrets SecretsV1, sessions SessionsV1, clock ClockV1, accounts AccountsV1) *AuthenticateHandler {
	return &AuthenticateHandler{secrets: secrets, sessions: sessions, clock: clock, accounts: accounts}
}

// Execute runs Authenticate.
func (h *AuthenticateHandler) Execute(ctx context.Context, in AuthenticateInput) (AuthenticateOutput, error) {
	return h.execute(ctx, &in)
}

var _ Authenticate = (*AuthenticateHandler)(nil)

// SignIn opens a session for an address and its password; five wrong passwords
// in a row lock the account for fifteen minutes.
type SignIn interface {
	Execute(ctx context.Context, in SignInInput) (SignInOutput, error)
}

// SignInInput is SignIn's input: a field added breaks no caller.
type SignInInput struct {
	Email    string
	Password string
}

// SignInOutput is SignIn's output.
type SignInOutput struct {
	Secret  string
	Session account.Session
	Account account.Account
}

// SignInHandler runs SignIn: its body is the hand-written execute.
type SignInHandler struct {
	accounts  AccountsV1
	clock     ClockV1
	passwords PasswordsV1
	secrets   SecretsV1
	sessions  SessionsV1
}

// NewSignIn builds SignIn on what it calls.
func NewSignIn(accounts AccountsV1, clock ClockV1, passwords PasswordsV1, secrets SecretsV1, sessions SessionsV1) *SignInHandler {
	return &SignInHandler{accounts: accounts, clock: clock, passwords: passwords, secrets: secrets, sessions: sessions}
}

// Execute runs SignIn.
func (h *SignInHandler) Execute(ctx context.Context, in SignInInput) (SignInOutput, error) {
	return h.execute(ctx, &in)
}

var _ SignIn = (*SignInHandler)(nil)

// SignOut ends a session.
type SignOut interface {
	Execute(ctx context.Context, in SignOutInput) (SignOutOutput, error)
}

// SignOutInput is SignOut's input: a field added breaks no caller.
type SignOutInput struct {
	SessionID string
}

// SignOutOutput is SignOut's output.
type SignOutOutput struct {
}

// SignOutHandler runs SignOut: its body is the hand-written execute.
type SignOutHandler struct {
	sessions SessionsV1
}

// NewSignOut builds SignOut on what it calls.
func NewSignOut(sessions SessionsV1) *SignOutHandler {
	return &SignOutHandler{sessions: sessions}
}

// Execute runs SignOut.
func (h *SignOutHandler) Execute(ctx context.Context, in SignOutInput) (SignOutOutput, error) {
	return h.execute(ctx, &in)
}

var _ SignOut = (*SignOutHandler)(nil)

// SignUp creates an account.
type SignUp interface {
	Execute(ctx context.Context, in SignUpInput) (SignUpOutput, error)
}

// SignUpInput is SignUp's input: a field added breaks no caller.
type SignUpInput struct {
	Email    string
	Name     string
	Password string
}

// SignUpOutput is SignUp's output.
type SignUpOutput struct {
	Account account.Account
}

// SignUpHandler runs SignUp: its body is the hand-written execute.
type SignUpHandler struct {
	accounts  AccountsV1
	passwords PasswordsV1
	clock     ClockV1
}

// NewSignUp builds SignUp on what it calls.
func NewSignUp(accounts AccountsV1, passwords PasswordsV1, clock ClockV1) *SignUpHandler {
	return &SignUpHandler{accounts: accounts, passwords: passwords, clock: clock}
}

// Execute runs SignUp.
func (h *SignUpHandler) Execute(ctx context.Context, in SignUpInput) (SignUpOutput, error) {
	return h.execute(ctx, &in)
}

var _ SignUp = (*SignUpHandler)(nil)
