// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:dd6d9d76cd70cb3afe4d348b4ed0e2bdf2f8aa9f13db84b61b2cb1c34cdd43a0

// Package authapi — The HTTP API of the accounts, and the handler that says
// who calls a route.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package authapi

import (
	"context"

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

// Credentials — What a request proves its member with, the session cookie of
// a browser or the same secret as a bearer token.
type Credentials struct {
	Cookie        string `cookie:"todo_session"`
	Authorization string `header:"Authorization"`
}

// NoInput — A request, or an answer, that says nothing.
type NoInput struct {
}

// Principal — What the handler says of the caller.
type Principal struct {
	SessionID string `json:"sessionId" app:"public"`
	Email     string `json:"email" app:"personal"`
	Name      string `json:"name" app:"personal"`
}

// Profile — A member, as the API answers them.
type Profile struct {
	ID    string `json:"id" app:"public"`
	Email string `json:"email" app:"personal"`
	Name  string `json:"name" app:"personal"`
}

// SignInRequest — An address and its password.
type SignInRequest struct {
	Email    string `json:"email" app:"personal"`
	Password string `json:"password" app:"secret"`
}

// SignUpRequest — What a stranger sends to become a member.
type SignUpRequest struct {
	Email    string `json:"email" app:"personal"`
	Name     string `json:"name" app:"personal"`
	Password string `json:"password" app:"secret"`
}

// SignedIn — A session just opened, the member and the secret a client that
// keeps no cookie sends back as a bearer token.
type SignedIn struct {
	User  Profile `json:"user"`
	Token string  `json:"token" app:"secret"`
}

// Service is authapi's framework service: what it declares, drawn and run.
var Service = app.NewService("authapi", "The HTTP API of the accounts, and the handler that says who calls a route.")

var _ = Service.Endpoint("POST /api/auth/signup", SignUp)

// SignUp serves POST /api/auth/signup: its body is the hand-written signUp.
func SignUp(ctx context.Context, in SignUpRequest) (Profile, error) {
	return signUp(ctx, in)
}

var _ = Service.Endpoint("POST /api/auth/login", SignIn)

// SignIn serves POST /api/auth/login: its body is the hand-written signIn.
func SignIn(ctx context.Context, in SignInRequest) (SignedIn, error) {
	return signIn(ctx, in)
}

var _ = Service.Endpoint("POST /api/auth/logout", SignOut, app.Auth())

// SignOut serves POST /api/auth/logout: its body is the hand-written signOut.
func SignOut(ctx context.Context, in NoInput) (NoInput, error) {
	return signOut(ctx, in)
}

var _ = Service.Endpoint("GET /api/auth/me", Me, app.Auth())

// Me serves GET /api/auth/me: its body is the hand-written me.
func Me(ctx context.Context, in NoInput) (Profile, error) {
	return me(ctx, in)
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
	signUp       access.SignUp
	signIn       access.SignIn
	signOut      access.SignOut
	authenticate access.Authenticate
}

// Wire gives the handlers what they use: the role that mounts authapi calls it
// before its App runs.
func Wire(signUp access.SignUp, signIn access.SignIn, signOut access.SignOut, authenticate access.Authenticate) {
	wired.signUp = signUp
	wired.signIn = signIn
	wired.signOut = signOut
	wired.authenticate = authenticate
}
