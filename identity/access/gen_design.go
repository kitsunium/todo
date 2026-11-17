// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:ad3708a6c0120e4ecb328f05fe017711e138b2d9494fff3cf0c4cd1c6596edae

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
	// Change applies a change to the account of an id as it is kept at that
	// instant, alone among the writers of the store, and keeps the result; an
	// error of the change keeps nothing.
	Change(ctx context.Context, id string, change func(*account.Account) error) (account.Account, error)
	Get(ctx context.Context, id string) (account.Account, bool, error)
	NewID(ctx context.Context) (string, error)
	Save(ctx context.Context, a account.Account) error
}

// AnnouncementsV1 tells the rest of the product what happened to an account,
// for whoever listens.
type AnnouncementsV1 interface {
	// Verified says the address of an account is proven.
	Verified(ctx context.Context, acc account.Account) error
}

// ClockV1 tells the time.
type ClockV1 interface {
	Now(ctx context.Context) (time.Time, error)
}

// CourierV1 writes to a member the mails of their account, each with the
// link's secret only the mail carries.
type CourierV1 interface {
	AlreadyMember(ctx context.Context, to account.Account) error
	Reset(ctx context.Context, to account.Account, secret string) error
	Verification(ctx context.Context, to account.Account, secret string) error
}

// LinksV1 keeps the one-time links.
type LinksV1 interface {
	ByHash(ctx context.Context, hash string) (account.Link, bool, error)
	DeleteOf(ctx context.Context, userID string, purpose string) error
	NewID(ctx context.Context) (string, error)
	Save(ctx context.Context, l account.Link) error
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
	DeleteOf(ctx context.Context, userID string, keep string) error
	NewID(ctx context.Context) (string, error)
	Of(ctx context.Context, userID string) ([]account.Session, error)
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

// ChangePassword replaces a member's password for one who knows the current
// one, and ends their other sessions.
type ChangePassword interface {
	Execute(ctx context.Context, in ChangePasswordInput) (ChangePasswordOutput, error)
}

// ChangePasswordInput is ChangePassword's input: a field added breaks no caller.
type ChangePasswordInput struct {
	UserID    string
	SessionID string
	Current   string
	Password  string
}

// ChangePasswordOutput is ChangePassword's output.
type ChangePasswordOutput struct {
}

// ChangePasswordHandler runs ChangePassword: its body is the hand-written execute.
type ChangePasswordHandler struct {
	accounts  AccountsV1
	passwords PasswordsV1
	clock     ClockV1
	sessions  SessionsV1
}

// NewChangePassword builds ChangePassword on what it calls.
func NewChangePassword(accounts AccountsV1, passwords PasswordsV1, clock ClockV1, sessions SessionsV1) *ChangePasswordHandler {
	return &ChangePasswordHandler{accounts: accounts, passwords: passwords, clock: clock, sessions: sessions}
}

// Execute runs ChangePassword.
func (h *ChangePasswordHandler) Execute(ctx context.Context, in ChangePasswordInput) (ChangePasswordOutput, error) {
	return h.execute(ctx, &in)
}

var _ ChangePassword = (*ChangePasswordHandler)(nil)

// ForgotPassword mails an address the link that chooses a new password, when
// it has an account, and answers alike when it has none.
type ForgotPassword interface {
	Execute(ctx context.Context, in ForgotPasswordInput) (ForgotPasswordOutput, error)
}

// ForgotPasswordInput is ForgotPassword's input: a field added breaks no caller.
type ForgotPasswordInput struct {
	Email string
}

// ForgotPasswordOutput is ForgotPassword's output.
type ForgotPasswordOutput struct {
}

// ForgotPasswordHandler runs ForgotPassword: its body is the hand-written execute.
type ForgotPasswordHandler struct {
	accounts AccountsV1
	clock    ClockV1
	links    LinksV1
	secrets  SecretsV1
	courier  CourierV1
}

// NewForgotPassword builds ForgotPassword on what it calls.
func NewForgotPassword(accounts AccountsV1, clock ClockV1, links LinksV1, secrets SecretsV1, courier CourierV1) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{accounts: accounts, clock: clock, links: links, secrets: secrets, courier: courier}
}

// Execute runs ForgotPassword.
func (h *ForgotPasswordHandler) Execute(ctx context.Context, in ForgotPasswordInput) (ForgotPasswordOutput, error) {
	return h.execute(ctx, &in)
}

var _ ForgotPassword = (*ForgotPasswordHandler)(nil)

// ListSessions lists the sessions a member has open, the one last seen first.
type ListSessions interface {
	Execute(ctx context.Context, in ListSessionsInput) (ListSessionsOutput, error)
}

// ListSessionsInput is ListSessions's input: a field added breaks no caller.
type ListSessionsInput struct {
	UserID string
}

// ListSessionsOutput is ListSessions's output.
type ListSessionsOutput struct {
	Sessions []account.Session
}

// ListSessionsHandler runs ListSessions: its body is the hand-written execute.
type ListSessionsHandler struct {
	sessions SessionsV1
	clock    ClockV1
}

// NewListSessions builds ListSessions on what it calls.
func NewListSessions(sessions SessionsV1, clock ClockV1) *ListSessionsHandler {
	return &ListSessionsHandler{sessions: sessions, clock: clock}
}

// Execute runs ListSessions.
func (h *ListSessionsHandler) Execute(ctx context.Context, in ListSessionsInput) (ListSessionsOutput, error) {
	return h.execute(ctx, &in)
}

var _ ListSessions = (*ListSessionsHandler)(nil)

// ResendVerification mails a new link to an address whose account is not
// verified yet, and says nothing of the others.
type ResendVerification interface {
	Execute(ctx context.Context, in ResendVerificationInput) (ResendVerificationOutput, error)
}

// ResendVerificationInput is ResendVerification's input: a field added breaks no caller.
type ResendVerificationInput struct {
	Email string
}

// ResendVerificationOutput is ResendVerification's output.
type ResendVerificationOutput struct {
}

// ResendVerificationHandler runs ResendVerification: its body is the hand-written execute.
type ResendVerificationHandler struct {
	accounts AccountsV1
	clock    ClockV1
	links    LinksV1
	secrets  SecretsV1
	courier  CourierV1
}

// NewResendVerification builds ResendVerification on what it calls.
func NewResendVerification(accounts AccountsV1, clock ClockV1, links LinksV1, secrets SecretsV1, courier CourierV1) *ResendVerificationHandler {
	return &ResendVerificationHandler{accounts: accounts, clock: clock, links: links, secrets: secrets, courier: courier}
}

// Execute runs ResendVerification.
func (h *ResendVerificationHandler) Execute(ctx context.Context, in ResendVerificationInput) (ResendVerificationOutput, error) {
	return h.execute(ctx, &in)
}

var _ ResendVerification = (*ResendVerificationHandler)(nil)

// ResetPassword chooses a new password with the secret of a reset link, lifts
// the lock, proves the address and ends every session.
type ResetPassword interface {
	Execute(ctx context.Context, in ResetPasswordInput) (ResetPasswordOutput, error)
}

// ResetPasswordInput is ResetPassword's input: a field added breaks no caller.
type ResetPasswordInput struct {
	Secret   string
	Password string
}

// ResetPasswordOutput is ResetPassword's output.
type ResetPasswordOutput struct {
}

// ResetPasswordHandler runs ResetPassword: its body is the hand-written execute.
type ResetPasswordHandler struct {
	secrets       SecretsV1
	links         LinksV1
	clock         ClockV1
	accounts      AccountsV1
	passwords     PasswordsV1
	sessions      SessionsV1
	announcements AnnouncementsV1
}

// NewResetPassword builds ResetPassword on what it calls.
func NewResetPassword(secrets SecretsV1, links LinksV1, clock ClockV1, accounts AccountsV1, passwords PasswordsV1, sessions SessionsV1, announcements AnnouncementsV1) *ResetPasswordHandler {
	return &ResetPasswordHandler{secrets: secrets, links: links, clock: clock, accounts: accounts, passwords: passwords, sessions: sessions, announcements: announcements}
}

// Execute runs ResetPassword.
func (h *ResetPasswordHandler) Execute(ctx context.Context, in ResetPasswordInput) (ResetPasswordOutput, error) {
	return h.execute(ctx, &in)
}

var _ ResetPassword = (*ResetPasswordHandler)(nil)

// RevokeSession ends one of a member's sessions; another member's is not
// found.
type RevokeSession interface {
	Execute(ctx context.Context, in RevokeSessionInput) (RevokeSessionOutput, error)
}

// RevokeSessionInput is RevokeSession's input: a field added breaks no caller.
type RevokeSessionInput struct {
	UserID string
	ID     string
}

// RevokeSessionOutput is RevokeSession's output.
type RevokeSessionOutput struct {
}

// RevokeSessionHandler runs RevokeSession: its body is the hand-written execute.
type RevokeSessionHandler struct {
	sessions SessionsV1
}

// NewRevokeSession builds RevokeSession on what it calls.
func NewRevokeSession(sessions SessionsV1) *RevokeSessionHandler {
	return &RevokeSessionHandler{sessions: sessions}
}

// Execute runs RevokeSession.
func (h *RevokeSessionHandler) Execute(ctx context.Context, in RevokeSessionInput) (RevokeSessionOutput, error) {
	return h.execute(ctx, &in)
}

var _ RevokeSession = (*RevokeSessionHandler)(nil)

// SignIn opens a session for an address and its password; five wrong passwords
// in a row lock the account for fifteen minutes.
type SignIn interface {
	Execute(ctx context.Context, in SignInInput) (SignInOutput, error)
}

// SignInInput is SignIn's input: a field added breaks no caller.
type SignInInput struct {
	Email     string
	Password  string
	UserAgent string
	IP        string
	TimeZone  string
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

// SignUp creates an account and mails the link that proves its address; an
// address that already has an account is answered alike, and its member told.
type SignUp interface {
	Execute(ctx context.Context, in SignUpInput) (SignUpOutput, error)
}

// SignUpInput is SignUp's input: a field added breaks no caller.
type SignUpInput struct {
	Email    string
	Name     string
	Password string
	Trusted  bool
	Locale   string
	TimeZone string
}

// SignUpOutput is SignUp's output.
type SignUpOutput struct {
	Email    string
	Verified bool
}

// SignUpHandler runs SignUp: its body is the hand-written execute.
type SignUpHandler struct {
	accounts      AccountsV1
	passwords     PasswordsV1
	clock         ClockV1
	links         LinksV1
	secrets       SecretsV1
	courier       CourierV1
	announcements AnnouncementsV1
}

// NewSignUp builds SignUp on what it calls.
func NewSignUp(accounts AccountsV1, passwords PasswordsV1, clock ClockV1, links LinksV1, secrets SecretsV1, courier CourierV1, announcements AnnouncementsV1) *SignUpHandler {
	return &SignUpHandler{accounts: accounts, passwords: passwords, clock: clock, links: links, secrets: secrets, courier: courier, announcements: announcements}
}

// Execute runs SignUp.
func (h *SignUpHandler) Execute(ctx context.Context, in SignUpInput) (SignUpOutput, error) {
	return h.execute(ctx, &in)
}

var _ SignUp = (*SignUpHandler)(nil)

// UpdateProfile changes the name, the language and the time zone of a member,
// each only when it is given.
type UpdateProfile interface {
	Execute(ctx context.Context, in UpdateProfileInput) (UpdateProfileOutput, error)
}

// UpdateProfileInput is UpdateProfile's input: a field added breaks no caller.
type UpdateProfileInput struct {
	UserID   string
	Name     *string
	Locale   *string
	TimeZone *string
}

// UpdateProfileOutput is UpdateProfile's output.
type UpdateProfileOutput struct {
	Account account.Account
}

// UpdateProfileHandler runs UpdateProfile: its body is the hand-written execute.
type UpdateProfileHandler struct {
	clock    ClockV1
	accounts AccountsV1
}

// NewUpdateProfile builds UpdateProfile on what it calls.
func NewUpdateProfile(clock ClockV1, accounts AccountsV1) *UpdateProfileHandler {
	return &UpdateProfileHandler{clock: clock, accounts: accounts}
}

// Execute runs UpdateProfile.
func (h *UpdateProfileHandler) Execute(ctx context.Context, in UpdateProfileInput) (UpdateProfileOutput, error) {
	return h.execute(ctx, &in)
}

var _ UpdateProfile = (*UpdateProfileHandler)(nil)

// VerifyEmail proves an account's address with the secret of the link it was
// mailed, and signs its member in, since the link was in their inbox.
type VerifyEmail interface {
	Execute(ctx context.Context, in VerifyEmailInput) (VerifyEmailOutput, error)
}

// VerifyEmailInput is VerifyEmail's input: a field added breaks no caller.
type VerifyEmailInput struct {
	Secret    string
	UserAgent string
	IP        string
}

// VerifyEmailOutput is VerifyEmail's output.
type VerifyEmailOutput struct {
	Account       account.Account
	SessionSecret string
	Session       account.Session
}

// VerifyEmailHandler runs VerifyEmail: its body is the hand-written execute.
type VerifyEmailHandler struct {
	secrets       SecretsV1
	links         LinksV1
	clock         ClockV1
	accounts      AccountsV1
	sessions      SessionsV1
	announcements AnnouncementsV1
}

// NewVerifyEmail builds VerifyEmail on what it calls.
func NewVerifyEmail(secrets SecretsV1, links LinksV1, clock ClockV1, accounts AccountsV1, sessions SessionsV1, announcements AnnouncementsV1) *VerifyEmailHandler {
	return &VerifyEmailHandler{secrets: secrets, links: links, clock: clock, accounts: accounts, sessions: sessions, announcements: announcements}
}

// Execute runs VerifyEmail.
func (h *VerifyEmailHandler) Execute(ctx context.Context, in VerifyEmailInput) (VerifyEmailOutput, error) {
	return h.execute(ctx, &in)
}

var _ VerifyEmail = (*VerifyEmailHandler)(nil)
