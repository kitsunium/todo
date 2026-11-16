// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package authapi

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/observe/logger"
	"github.com/kitsunium/todo/identity/access"
	"github.com/kitsunium/todo/identity/account"
)

// signUp is POST /api/auth/signup's body; what it uses is in wired.
func signUp(ctx context.Context, in SignUpRequest) (out Profile, err error) {
	res, err := wired.signUp.Execute(ctx, access.SignUpInput{Email: in.Email, Name: in.Name, Password: in.Password})
	if err != nil {
		return Profile{}, refusal(err)
	}
	logger.Info(ctx, app.Log(ctx), "account created", logger.String("user", res.Account.ID))
	return profileOf(res.Account), nil
}

// signIn is POST /api/auth/login's body; what it uses is in wired. The
// session's secret goes to a browser as its cookie, and in the answer for a
// client that keeps none.
func signIn(ctx context.Context, in SignInRequest) (out SignedIn, err error) {
	res, err := wired.signIn.Execute(ctx, access.SignInInput{Email: in.Email, Password: in.Password})
	if err != nil {
		return SignedIn{}, refusal(err)
	}
	setCookie(ctx, res.Secret)
	return SignedIn{User: profileOf(res.Account), Token: res.Secret}, nil
}

// signOut is POST /api/auth/logout's body; what it uses is in wired.
func signOut(ctx context.Context, in NoInput) (out NoInput, err error) {
	p, _ := app.AuthData[Principal](ctx)
	if _, err := wired.signOut.Execute(ctx, access.SignOutInput{SessionID: p.SessionID}); err != nil {
		return NoInput{}, err
	}
	app.ClearCookie(ctx, CookieName)
	return NoInput{}, nil
}

// me is GET /api/auth/me's body: the caller, as the session handler named
// them.
func me(ctx context.Context, in NoInput) (out Profile, err error) {
	uid, ok := app.UserID(ctx)
	if !ok {
		return Profile{}, app.Unauthenticated("sign in first")
	}
	p, _ := app.AuthData[Principal](ctx)
	return Profile{ID: string(uid), Email: p.Email, Name: p.Name}, nil
}

// authenticate is the session handler's body: it says who calls. The secret
// is the session cookie's, else a bearer token's; a request with neither
// has no user, which is no error. A secret that opens no session is refused,
// and its cookie cleared; a session renewed has its cookie renewed.
func authenticate(ctx context.Context, in Credentials) (uid app.UID, data Principal, err error) {
	secret, fromCookie := in.Cookie, in.Cookie != ""
	if !fromCookie {
		scheme, token, ok := strings.Cut(in.Authorization, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			return "", Principal{}, nil
		}
		secret = token
	}
	res, err := wired.authenticate.Execute(ctx, access.AuthenticateInput{Secret: secret})
	switch {
	case errors.Is(err, account.ErrNoSession):
		if fromCookie {
			app.ClearCookie(ctx, CookieName)
		}
		return "", Principal{}, app.Unauthenticated("your session has expired: sign in again")
	case err != nil:
		return "", Principal{}, err
	}
	if fromCookie && res.Renewed {
		setCookie(ctx, secret)
	}
	return app.UID(res.Account.ID), Principal{SessionID: res.Session.ID, Email: res.Account.Email, Name: res.Account.Name}, nil
}

// setCookie gives the browser the session's secret. The framework adds
// Secure outside dev; HttpOnly and SameSite=Lax keep it from scripts and
// from cross-site requests.
func setCookie(ctx context.Context, secret string) {
	app.SetCookie(ctx, &http.Cookie{
		Name: CookieName, Value: secret, Path: "/", MaxAge: int(account.SessionFor / time.Second),
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}

// profileOf is an account as the API answers it.
func profileOf(a account.Account) Profile {
	return Profile{ID: a.ID, Email: a.Email, Name: a.Name}
}

// refusal is a use case's error as the API answers it, with the codes a
// client branches on: 400 with the field at fault, 409 for an address
// already taken, 401 invalid_credentials, 403 account_locked.
func refusal(err error) error {
	invalid := func(path, rule string, cause error) error {
		e := app.Invalid("the request is invalid")
		e.Violations = []app.Violation{{Path: path, Rule: rule, Message: cause.Error()}}
		return e
	}
	switch {
	case errors.Is(err, account.ErrInvalidEmail):
		return invalid("email", "email", err)
	case errors.Is(err, account.ErrNameRequired):
		return invalid("name", "required", err)
	case errors.Is(err, account.ErrWeakPassword):
		return invalid("password", "length", err)
	case errors.Is(err, account.ErrEmailTaken):
		return app.Conflict("this address already has an account")
	case errors.Is(err, account.ErrInvalidCredentials):
		return &app.Error{Status: http.StatusUnauthorized, Code: "invalid_credentials", Message: "the email address or the password is not right"}
	case errors.Is(err, account.ErrLocked):
		return &app.Error{Status: http.StatusForbidden, Code: "account_locked", Message: "too many wrong passwords: the account is locked for 15 minutes"}
	}
	return err
}
