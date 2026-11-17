// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:fff3fa1ce1d5d0ffb043be2f4bd6a1b7fa06f149490a7503c5b7744197cb6684

// Package authapi — The HTTP API of the accounts, and the handler that says
// who calls a route.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package authapi

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/identity/access"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 6, 4, 255), "NOT_IMPLEMENTED", "not implemented yet", "authapi: a hand-written body is still a skeleton")

const (
	// CookieName is the cookie a browser keeps its session's secret in.
	CookieName string = "todo_session"
)

// ChangeRequest — The current password and the new one.
type ChangeRequest struct {
	Current  string `json:"current" app:"secret"`
	Password string `json:"password" app:"secret"`
}

// Credentials — What a request proves its member with, the session cookie of
// a browser or the same secret as a bearer token.
type Credentials struct {
	Cookie        string `cookie:"todo_session"`
	Authorization string `header:"Authorization"`
}

// EmailRequest — An address.
type EmailRequest struct {
	Email string `json:"email" app:"personal"`
}

// LinkRequest — The secret of a one-time link.
type LinkRequest struct {
	Token string `json:"token" app:"secret"`
}

// NoInput — A request, or an answer, that says nothing.
type NoInput struct {
}

// Principal — What the handler says of the caller.
type Principal struct {
	SessionID string `json:"sessionId" app:"public"`
	Email     string `json:"email" app:"personal"`
	Name      string `json:"name" app:"personal"`
	Locale    string `json:"locale" app:"public"`
	TimeZone  string `json:"timeZone" app:"personal"`
}

// Profile — A member, as the API answers them.
type Profile struct {
	ID       string `json:"id" app:"public"`
	Email    string `json:"email" app:"personal"`
	Name     string `json:"name" app:"personal"`
	Locale   string `json:"locale" app:"public"`
	TimeZone string `json:"timeZone" app:"personal"`
}

// ProfileRequest — What a member changes of their account; a member left out
// stays as it is.
type ProfileRequest struct {
	Name *string `json:"name,omitempty" app:"personal"`
	// Locale is the new language, fr or en: the next mails are in it.
	Locale *string `json:"locale,omitempty" app:"public"`
	// TimeZone is the new IANA zone; an empty name is UTC.
	TimeZone *string `json:"timeZone,omitempty" app:"personal"`
}

// ResetRequest — The secret of a reset link and the new password.
type ResetRequest struct {
	Token    string `json:"token" app:"secret"`
	Password string `json:"password" app:"secret"`
}

// SessionID — The session a route names.
type SessionID struct {
	ID string `path:"id"`
}

// SessionList — The caller's sessions, the one last seen first.
type SessionList struct {
	Sessions []SessionView `json:"sessions"`
}

// SessionView — A session of the caller, as the API answers it.
type SessionView struct {
	ID         string    `json:"id" app:"public"`
	UserAgent  string    `json:"userAgent" app:"personal"`
	IP         string    `json:"ip" app:"personal"`
	CreatedAt  time.Time `json:"createdAt" app:"public"`
	LastSeenAt time.Time `json:"lastSeenAt" app:"public"`
	// Current says the request runs in this session.
	Current bool `json:"current" app:"public"`
}

// SignInRequest — An address and its password, and the time zone of the
// browser it comes from.
type SignInRequest struct {
	Email    string `json:"email" app:"personal"`
	Password string `json:"password" app:"secret"`
	// TimeZone is the IANA zone of the browser; when it is sent, it becomes the
	// account's.
	TimeZone string `json:"timeZone"`
}

// SignUpRequest — What a stranger sends to become a member.
type SignUpRequest struct {
	Email    string `json:"email" app:"personal"`
	Name     string `json:"name" app:"personal"`
	Password string `json:"password" app:"secret"`
	// Locale is fr or en; when it is not sent, the language the Accept-Language
	// header prefers, French by default.
	Locale string `json:"locale"`
	// TimeZone is the IANA zone of the member's browser, like Europe/Paris; empty
	// reads UTC.
	TimeZone       string `json:"timeZone"`
	AcceptLanguage string `header:"Accept-Language"`
}

// SignedIn — A session just opened, the member and the secret a client that
// keeps no cookie sends back as a bearer token.
type SignedIn struct {
	User  Profile `json:"user"`
	Token string  `json:"token" app:"secret"`
}

// SignedUp — What a sign-up answers, the same whether the address had an
// account or not.
type SignedUp struct {
	Email string `json:"email" app:"personal"`
	// Status is verification_sent, or verified where the deployment trusts
	// addresses.
	Status string `json:"status" app:"public"`
}

// Service is authapi's framework service: what it declares, drawn and run.
var Service = app.NewService("authapi", "The HTTP API of the accounts, and the handler that says who calls a route.")

// VerifyAddressesSetting — Whether a new account proves its address by a
// mailed link before it signs in; off where the deployment has no mail relay.
var VerifyAddressesSetting = Service.Setting[bool]("verify-addresses", true)

var _ = Service.Endpoint("POST /api/auth/signup", SignUp, app.RateLimitPerClient(2, 5))

// SignUp serves POST /api/auth/signup: its body is the hand-written signUp.
func SignUp(ctx context.Context, in SignUpRequest) (SignedUp, error) {
	return signUp(ctx, in)
}

var _ = Service.Endpoint("POST /api/auth/verify", Verify, app.RateLimitPerClient(2, 5))

// Verify serves POST /api/auth/verify: its body is the hand-written verify.
func Verify(ctx context.Context, in LinkRequest) (SignedIn, error) {
	return verify(ctx, in)
}

var _ = Service.Endpoint("POST /api/auth/verify/resend", Resend, app.RateLimitPerClient(2, 5))

// Resend serves POST /api/auth/verify/resend: its body is the hand-written resend.
func Resend(ctx context.Context, in EmailRequest) (NoInput, error) {
	return resend(ctx, in)
}

var _ = Service.Endpoint("POST /api/auth/password/forgot", Forgot, app.RateLimitPerClient(2, 5))

// Forgot serves POST /api/auth/password/forgot: its body is the hand-written forgot.
func Forgot(ctx context.Context, in EmailRequest) (NoInput, error) {
	return forgot(ctx, in)
}

var _ = Service.Endpoint("POST /api/auth/password/reset", Reset, app.RateLimitPerClient(2, 5))

// Reset serves POST /api/auth/password/reset: its body is the hand-written reset.
func Reset(ctx context.Context, in ResetRequest) (NoInput, error) {
	return reset(ctx, in)
}

var _ = Service.Endpoint("POST /api/auth/password", ChangePassword, app.Auth(), app.RateLimitPerClient(2, 5))

// ChangePassword serves POST /api/auth/password: its body is the hand-written changePassword.
func ChangePassword(ctx context.Context, in ChangeRequest) (NoInput, error) {
	return changePassword(ctx, in)
}

var _ = Service.Endpoint("POST /api/auth/login", SignIn, app.RateLimitPerClient(2, 5))

// SignIn serves POST /api/auth/login: its body is the hand-written signIn.
func SignIn(ctx context.Context, in SignInRequest) (SignedIn, error) {
	return signIn(ctx, in)
}

var _ = Service.Endpoint("POST /api/auth/logout", SignOut, app.Auth(), app.RateLimitPerClient(10, 20))

// SignOut serves POST /api/auth/logout: its body is the hand-written signOut.
func SignOut(ctx context.Context, in NoInput) (NoInput, error) {
	return signOut(ctx, in)
}

var _ = Service.Endpoint("GET /api/auth/sessions", Sessions, app.Auth())

// Sessions serves GET /api/auth/sessions: its body is the hand-written sessions.
func Sessions(ctx context.Context, in NoInput) (SessionList, error) {
	return sessions(ctx, in)
}

var _ = Service.Endpoint("DELETE /api/auth/sessions/{id}", Revoke, app.Auth(), app.RateLimitPerClient(10, 20))

// Revoke serves DELETE /api/auth/sessions/{id}: its body is the hand-written revoke.
func Revoke(ctx context.Context, in SessionID) (NoInput, error) {
	return revoke(ctx, in)
}

var _ = Service.Endpoint("GET /api/auth/me", Me, app.Auth())

// Me serves GET /api/auth/me: its body is the hand-written me.
func Me(ctx context.Context, in NoInput) (Profile, error) {
	return me(ctx, in)
}

var _ = Service.Endpoint("PATCH /api/auth/me", UpdateMe, app.Auth(), app.RateLimitPerClient(10, 20))

// UpdateMe serves PATCH /api/auth/me: its body is the hand-written updateMe.
func UpdateMe(ctx context.Context, in ProfileRequest) (Profile, error) {
	return updateMe(ctx, in)
}

var _ = Service.AuthHandler("session", Authenticate)

// Authenticate is the authentication "session": it reads a request's credentials and answers
// the user's id and data, an empty id for no credentials. Its body is the
// hand-written authenticate.
func Authenticate(ctx context.Context, in Credentials) (app.UID, Principal, error) {
	return authenticate(ctx, in)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
	signUp                 access.SignUp
	verifyAddressesSetting *app.Setting[bool]
	verifyEmail            access.VerifyEmail
	resendVerification     access.ResendVerification
	forgotPassword         access.ForgotPassword
	resetPassword          access.ResetPassword
	changePassword         access.ChangePassword
	signIn                 access.SignIn
	signOut                access.SignOut
	listSessions           access.ListSessions
	revokeSession          access.RevokeSession
	updateProfile          access.UpdateProfile
	authenticate           access.Authenticate
}

// Wire gives the handlers what they use: the role that mounts authapi calls it
// before its App runs.
func Wire(signUp access.SignUp, verifyAddressesSetting *app.Setting[bool], verifyEmail access.VerifyEmail, resendVerification access.ResendVerification, forgotPassword access.ForgotPassword, resetPassword access.ResetPassword, changePassword access.ChangePassword, signIn access.SignIn, signOut access.SignOut, listSessions access.ListSessions, revokeSession access.RevokeSession, updateProfile access.UpdateProfile, authenticate access.Authenticate) {
	wired.signUp = signUp
	wired.verifyAddressesSetting = verifyAddressesSetting
	wired.verifyEmail = verifyEmail
	wired.resendVerification = resendVerification
	wired.forgotPassword = forgotPassword
	wired.resetPassword = resetPassword
	wired.changePassword = changePassword
	wired.signIn = signIn
	wired.signOut = signOut
	wired.listSessions = listSessions
	wired.revokeSession = revokeSession
	wired.updateProfile = updateProfile
	wired.authenticate = authenticate
}
