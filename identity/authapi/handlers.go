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

// signUp is POST /api/auth/signup's body; what it uses is in wired. Its
// answer is the same whether the address had an account or not.
func signUp(ctx context.Context, in SignUpRequest) (out SignedUp, err error) {
	res, err := wired.signUp.Execute(ctx, access.SignUpInput{
		Email: in.Email, Name: in.Name, Password: in.Password, Trusted: !wired.verifyAddressesSetting.Get(),
	})
	if err != nil {
		return SignedUp{}, refusal(err)
	}
	out = SignedUp{Email: res.Email, Status: "verification_sent"}
	if res.Verified {
		out.Status = "verified"
	}
	return out, nil
}

// verify is POST /api/auth/verify's body; what it uses is in wired.
func verify(ctx context.Context, in LinkRequest) (out Profile, err error) {
	res, err := wired.verifyEmail.Execute(ctx, access.VerifyEmailInput{Secret: in.Token})
	if err != nil {
		return Profile{}, refusal(err)
	}
	logger.Info(ctx, app.Log(ctx), "address verified", logger.String("user", res.Account.ID))
	return profileOf(res.Account), nil
}

// resend is POST /api/auth/verify/resend's body; what it uses is in wired.
func resend(ctx context.Context, in EmailRequest) (out NoInput, err error) {
	_, err = wired.resendVerification.Execute(ctx, access.ResendVerificationInput{Email: in.Email})
	return NoInput{}, err
}

// forgot is POST /api/auth/password/forgot's body; what it uses is in wired.
func forgot(ctx context.Context, in EmailRequest) (out NoInput, err error) {
	_, err = wired.forgotPassword.Execute(ctx, access.ForgotPasswordInput{Email: in.Email})
	return NoInput{}, err
}

// reset is POST /api/auth/password/reset's body; what it uses is in wired.
func reset(ctx context.Context, in ResetRequest) (out NoInput, err error) {
	if _, err := wired.resetPassword.Execute(ctx, access.ResetPasswordInput{Secret: in.Token, Password: in.Password}); err != nil {
		return NoInput{}, refusal(err)
	}
	return NoInput{}, nil
}

// changePassword is POST /api/auth/password's body; what it uses is in wired.
func changePassword(ctx context.Context, in ChangeRequest) (out NoInput, err error) {
	uid, ok := app.UserID(ctx)
	if !ok {
		return NoInput{}, app.Unauthenticated("sign in first")
	}
	p, _ := app.AuthData[Principal](ctx)
	_, err = wired.changePassword.Execute(ctx, access.ChangePasswordInput{
		UserID: string(uid), SessionID: p.SessionID, Current: in.Current, Password: in.Password,
	})
	if err != nil {
		return NoInput{}, refusal(err)
	}
	return NoInput{}, nil
}

// signIn is POST /api/auth/login's body; what it uses is in wired. The
// session's secret goes to a browser as its cookie, and in the answer for a
// client that keeps none.
func signIn(ctx context.Context, in SignInRequest) (out SignedIn, err error) {
	res, err := wired.signIn.Execute(ctx, access.SignInInput{
		Email: in.Email, Password: in.Password, UserAgent: app.UserAgent(ctx), IP: app.ClientIP(ctx),
	})
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

// sessions is GET /api/auth/sessions's body; what it uses is in wired.
func sessions(ctx context.Context, in NoInput) (out SessionList, err error) {
	uid, ok := app.UserID(ctx)
	if !ok {
		return SessionList{}, app.Unauthenticated("sign in first")
	}
	p, _ := app.AuthData[Principal](ctx)
	res, err := wired.listSessions.Execute(ctx, access.ListSessionsInput{UserID: string(uid)})
	if err != nil {
		return SessionList{}, err
	}
	out.Sessions = make([]SessionView, len(res.Sessions))
	for i, s := range res.Sessions {
		out.Sessions[i] = SessionView{
			ID: s.ID, UserAgent: s.UserAgent, IP: s.IP, CreatedAt: s.CreatedAt, LastSeenAt: s.LastSeenAt, Current: s.ID == p.SessionID,
		}
	}
	return out, nil
}

// revoke is DELETE /api/auth/sessions/{id}'s body; what it uses is in wired.
// Ending the session the request runs in also forgets its cookie.
func revoke(ctx context.Context, in SessionID) (out NoInput, err error) {
	uid, ok := app.UserID(ctx)
	if !ok {
		return NoInput{}, app.Unauthenticated("sign in first")
	}
	if _, err := wired.revokeSession.Execute(ctx, access.RevokeSessionInput{UserID: string(uid), ID: in.ID}); err != nil {
		if errors.Is(err, account.ErrNoSession) {
			return NoInput{}, app.NotFound("no such session")
		}
		return NoInput{}, err
	}
	if p, _ := app.AuthData[Principal](ctx); p.SessionID == in.ID {
		app.ClearCookie(ctx, CookieName)
	}
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
// client branches on: 400 with the field at fault or invalid_token for a link,
// 401 invalid_credentials, 403 account_locked or email_unverified.
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
	case errors.Is(err, account.ErrUnverified):
		return &app.Error{Status: http.StatusForbidden, Code: "email_unverified", Message: "confirm your email address first: the link is in your inbox"}
	case errors.Is(err, account.ErrBadLink):
		return &app.Error{Status: http.StatusBadRequest, Code: "invalid_token", Message: "this link is invalid or has expired: ask for a new one"}
	case errors.Is(err, account.ErrInvalidCredentials):
		return &app.Error{Status: http.StatusUnauthorized, Code: "invalid_credentials", Message: "the email address or the password is not right"}
	case errors.Is(err, account.ErrLocked):
		return &app.Error{Status: http.StatusForbidden, Code: "account_locked", Message: "too many wrong passwords: the account is locked for 15 minutes"}
	}
	return err
}
