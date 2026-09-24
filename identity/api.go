package identity

import (
	"cmp"
	"context"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/sdk/pkg/v1/logger"
	"github.com/kitsunium/sdk/pkg/v1/password"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/notify"
)

// The account API. What a stranger may call — sign up, verify, sign in,
// recover a password — answers the same whether an address has an account
// or not, and is rate limited hard: every one of them mails someone or
// checks a password.
var (
	_ = Service.Endpoint("POST /api/auth/signup", Signup, kit.RateLimitPerClient(2, 5))
	_ = Service.Endpoint("POST /api/auth/verify", Verify, kit.RateLimitPerClient(2, 5))
	_ = Service.Endpoint("POST /api/auth/verify/resend", ResendVerification, kit.RateLimitPerClient(2, 5))
	_ = Service.Endpoint("POST /api/auth/login", Login, kit.RateLimitPerClient(2, 5))
	_ = Service.Endpoint("POST /api/auth/logout", Logout, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("GET /api/auth/me", Me, kit.Auth())
	_ = Service.Endpoint("PATCH /api/auth/me", UpdateProfile, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("POST /api/auth/password", ChangePassword, kit.Auth(), kit.RateLimitPerClient(2, 5))
	_ = Service.Endpoint("POST /api/auth/password/forgot", ForgotPassword, kit.RateLimitPerClient(2, 5))
	_ = Service.Endpoint("POST /api/auth/password/reset", ResetPassword, kit.RateLimitPerClient(2, 5))
	_ = Service.Endpoint("GET /api/auth/sessions", ListSessions, kit.Auth())
	_ = Service.Endpoint("DELETE /api/auth/sessions/{id}", RevokeSession, kit.Auth(), kit.RateLimitPerClient(10, 20))
)

// The errors of sign-in. Their codes are the contract the web app branches
// on.
func errInvalidCredentials() error {
	return &kit.Error{Status: http.StatusUnauthorized, Code: "invalid_credentials", Message: "the email address or the password is not right"}
}

func errUnverified() error {
	return &kit.Error{Status: http.StatusForbidden, Code: "email_unverified", Message: "confirm your email address first: the link is in your inbox"}
}

func errLocked() error {
	return &kit.Error{Status: http.StatusForbidden, Code: "account_locked", Message: "too many wrong passwords: the account is locked for 15 minutes, or until you reset the password"}
}

// hashPassword stretches a password with the SDK's PBKDF2-SHA256.
func hashPassword(pw string) (string, error) {
	return password.Hash(password.PBKDF2SHA256, []byte(pw))
}

// decoy is a hash no password matches, checked when an address has no
// account so that a wrong address takes as long as a wrong password.
var decoy = sync.OnceValue(func() string {
	secret, _ := newSecret()
	h, _ := hashPassword(secret)
	return h
})

// SignupInput is a new account.
type SignupInput struct {
	Email    string `json:"email" validate:"required,maxlen=254"`
	Name     string `json:"name" validate:"required,maxlen=80"`
	Password string `json:"password" validate:"required,minlen=10,maxlen=128"`
	// Locale is the language of the account, "fr" or "en" — the one the
	// sign-up form was shown in. Without it, AcceptLanguage decides.
	Locale string `json:"locale,omitempty"`
	// AcceptLanguage is the browser's preference, negotiated over the
	// product's languages: French when it names neither.
	AcceptLanguage string `header:"Accept-Language"`
}

// SignupOutput says a verification mail is on its way — whether or not the
// address already had an account.
type SignupOutput struct {
	Status string `json:"status"`
	Email  string `json:"email"`
}

// Signup opens an unverified account and mails the link that verifies it,
// in the account's language: the one the form asked for, or the browser's.
// An address that already has an account gets an "account exists" mail
// instead — in its owner's language — and the caller the very same answer:
// sign-up never tells a stranger who uses the product.
func Signup(ctx context.Context, in SignupInput) (SignupOutput, error) {
	email := NormalizeEmail(in.Email)
	if !ValidEmail(email) {
		return SignupOutput{}, wire.Invalid("email", "email", "must be an email address")
	}
	name, err := wire.Line("name", in.Name, 80)
	if err != nil {
		return SignupOutput{}, err
	}
	locale := wire.NegotiateLocale(in.AcceptLanguage)
	if in.Locale != "" {
		if locale, err = wire.CheckLocale("locale", in.Locale); err != nil {
			return SignupOutput{}, err
		}
	}
	out := SignupOutput{Status: "verification_sent", Email: email}
	// Hash before looking the address up, so both answers take as long.
	hash, err := hashPassword(in.Password)
	if err != nil {
		return SignupOutput{}, err
	}
	existing, found, err := accountByEmail(ctx, email)
	if err != nil {
		return SignupOutput{}, err
	}
	if found {
		return out, sendAccountExists(ctx, existing)
	}
	now := wire.Now(ctx)
	acct, err := AccountLifecycle.Start(ctx, Account{
		ID: kit.NewID("user"), Email: email, Name: name, Locale: locale, PasswordHash: hash, CreatedAt: now, UpdatedAt: now,
	})
	if wire.Is(err, kit.CodeConflict) {
		// A concurrent sign-up took the address first: it is now an
		// address with an account.
		winner, found, err := accountByEmail(ctx, email)
		if err != nil || !found {
			return out, err
		}
		return out, sendAccountExists(ctx, winner)
	}
	if err != nil {
		return SignupOutput{}, err
	}
	logger.Info(ctx, kit.Log(ctx), "account created", logger.String("user", acct.ID), logger.String("locale", string(acct.Locale)))
	return out, sendVerification(ctx, acct)
}

// sendVerification mails a fresh verification link to the account.
func sendVerification(ctx context.Context, a Account) error {
	secret, err := issue(ctx, a.ID, PurposeVerify, VerifyFor)
	if err != nil {
		return err
	}
	_, err = notify.SendAPI.Call(ctx, notify.SendInput{
		Template: notify.TemplateVerifyEmail, To: a.Email, Name: a.Name, Locale: a.locale(), Data: map[string]string{"token": secret},
	})
	return err
}

// sendAccountExists tells the owner of an address that someone tried to
// sign up with it.
func sendAccountExists(ctx context.Context, a Account) error {
	_, err := notify.SendAPI.Call(ctx, notify.SendInput{Template: notify.TemplateAccountExists, To: a.Email, Name: a.Name, Locale: a.locale()})
	return err
}

// TokenInput is the secret of a one-time link.
type TokenInput struct {
	Token string `json:"token" validate:"required,maxlen=128"`
}

// Verify proves an account's address with the link of its verification
// mail, and signs it in. A link is good once, for an account still waiting
// for it.
func Verify(ctx context.Context, in TokenInput) (UserOutput, error) {
	t, err := redeem(ctx, in.Token, PurposeVerify)
	if err != nil {
		return UserOutput{}, err
	}
	a, err := AccountLifecycle.Fire(ctx, t.UserID, "verify")
	if wire.Is(err, kit.CodeConflict) {
		// Verified already, by another link of the same account: this one
		// signs no one in.
		return UserOutput{}, errBadLink()
	}
	if err != nil {
		return UserOutput{}, err
	}
	if err := openSession(ctx, a.ID); err != nil {
		return UserOutput{}, err
	}
	return UserOutput{User: a.user()}, nil
}

// EmailInput is an address.
type EmailInput struct {
	Email string `json:"email" validate:"required,maxlen=254"`
}

// SentOutput says a mail went out — or would have: the answer is the same
// for an address without an account.
type SentOutput struct {
	Status string `json:"status"`
}

// ResendVerification mails a new verification link to an account that has
// not verified its address yet.
func ResendVerification(ctx context.Context, in EmailInput) (SentOutput, error) {
	email := NormalizeEmail(in.Email)
	if !ValidEmail(email) {
		return SentOutput{}, wire.Invalid("email", "email", "must be an email address")
	}
	a, found, err := accountByEmail(ctx, email)
	if err != nil || !found || a.Status != Unverified {
		return SentOutput{Status: "sent"}, err
	}
	return SentOutput{Status: "sent"}, sendVerification(ctx, a)
}

// LoginInput is an address and a password.
type LoginInput struct {
	Email    string `json:"email" validate:"required,maxlen=254"`
	Password string `json:"password" validate:"required,maxlen=128"`
}

// Login signs a user in with their password. Five wrong passwords in a row
// lock the account for fifteen minutes; an unverified account is told to
// verify only once its password proved right, so the answer never tells a
// stranger whether an address has an account.
func Login(ctx context.Context, in LoginInput) (UserOutput, error) {
	a, found, err := accountByEmail(ctx, NormalizeEmail(in.Email))
	if err != nil {
		return UserOutput{}, err
	}
	if !found {
		_, _ = password.Verify([]byte(in.Password), decoy())
		return UserOutput{}, errInvalidCredentials()
	}
	if a.Status == Locked || (a.Status == Active && a.FailedLogins >= MaxFailedLogins) {
		return UserOutput{}, errLocked()
	}
	ok, err := password.Verify([]byte(in.Password), a.PasswordHash)
	if err != nil {
		return UserOutput{}, err
	}
	if !ok {
		if a.Status == Active {
			// The account's lock guard reads the count on its next sweep.
			if _, err := Accounts.Update(ctx, a.ID, func(a *Account) error { a.FailedLogins++; return nil }); err != nil {
				return UserOutput{}, err
			}
			logger.Warn(ctx, kit.Log(ctx), "wrong password", logger.String("user", a.ID), logger.Int("failures", a.FailedLogins+1))
		}
		return UserOutput{}, errInvalidCredentials()
	}
	if a.Status == Unverified {
		return UserOutput{}, errUnverified()
	}
	rehash := ""
	if password.NeedsRehash(a.PasswordHash) {
		// The policy grew since this hash was made: store a stronger one
		// now that the password is at hand.
		if rehash, err = hashPassword(in.Password); err != nil {
			return UserOutput{}, err
		}
	}
	if a.FailedLogins > 0 || rehash != "" {
		if a, err = Accounts.Update(ctx, a.ID, func(a *Account) error {
			a.FailedLogins = 0
			if rehash != "" {
				a.PasswordHash = rehash
			}
			return nil
		}); err != nil {
			return UserOutput{}, err
		}
	}
	if err := openSession(ctx, a.ID); err != nil {
		return UserOutput{}, err
	}
	return UserOutput{User: a.user()}, nil
}

// Logout ends the caller's session and forgets its cookie.
func Logout(ctx context.Context, _ kit.Empty) (kit.Empty, error) {
	_, p, err := caller(ctx)
	if err != nil {
		return kit.Empty{}, err
	}
	if err := Sessions.Delete(ctx, p.SessionID); err != nil && !wire.Is(err, kit.CodeNotFound) {
		return kit.Empty{}, err
	}
	kit.ClearCookie(ctx, CookieName)
	return kit.Empty{}, nil
}

// Me returns the caller's account.
func Me(ctx context.Context, _ kit.Empty) (UserOutput, error) {
	uid, _, err := caller(ctx)
	if err != nil {
		return UserOutput{}, err
	}
	a, err := Accounts.Get(ctx, uid)
	if err != nil {
		return UserOutput{}, err
	}
	return UserOutput{User: a.user()}, nil
}

// ProfileInput is what the caller changes of their account: the display
// name, the language, or both. A member left out stays as it is.
type ProfileInput struct {
	// Name is the new display name: one line, at most 80 characters.
	Name *string `json:"name,omitempty"`
	// Locale is the new language: "fr" or "en". The next mails are in it.
	Locale *string `json:"locale,omitempty"`
}

// UpdateProfile changes the caller's display name and language.
func UpdateProfile(ctx context.Context, in ProfileInput) (UserOutput, error) {
	uid, _, err := caller(ctx)
	if err != nil {
		return UserOutput{}, err
	}
	var name string
	if in.Name != nil {
		if name, err = wire.Line("name", *in.Name, 80); err != nil {
			return UserOutput{}, err
		}
	}
	var locale wire.Locale
	if in.Locale != nil {
		if locale, err = wire.CheckLocale("locale", *in.Locale); err != nil {
			return UserOutput{}, err
		}
	}
	if in.Name == nil && in.Locale == nil {
		a, err := Accounts.Get(ctx, uid)
		if err != nil {
			return UserOutput{}, err
		}
		return UserOutput{User: a.user()}, nil
	}
	now := wire.Now(ctx)
	a, err := Accounts.Update(ctx, uid, func(a *Account) error {
		if in.Name != nil {
			a.Name = name
		}
		if in.Locale != nil {
			a.Locale = locale
		}
		a.UpdatedAt = now
		return nil
	})
	if err != nil {
		return UserOutput{}, err
	}
	return UserOutput{User: a.user()}, nil
}

// ChangePasswordInput is the current password and the new one.
type ChangePasswordInput struct {
	Current  string `json:"current" validate:"required,maxlen=128"`
	Password string `json:"password" validate:"required,minlen=10,maxlen=128"`
}

// ChangePassword replaces the caller's password, once the current one is
// proved, and signs out every other session.
func ChangePassword(ctx context.Context, in ChangePasswordInput) (kit.Empty, error) {
	uid, p, err := caller(ctx)
	if err != nil {
		return kit.Empty{}, err
	}
	a, err := Accounts.Get(ctx, uid)
	if err != nil {
		return kit.Empty{}, err
	}
	ok, err := password.Verify([]byte(in.Current), a.PasswordHash)
	if err != nil {
		return kit.Empty{}, err
	}
	if !ok {
		return kit.Empty{}, wire.Invalid("current", "match", "is not your current password")
	}
	hash, err := hashPassword(in.Password)
	if err != nil {
		return kit.Empty{}, err
	}
	now := wire.Now(ctx)
	if a, err = Accounts.Update(ctx, uid, func(a *Account) error {
		a.PasswordHash, a.UpdatedAt = hash, now
		return nil
	}); err != nil {
		return kit.Empty{}, err
	}
	if err := revokeSessions(ctx, uid, p.SessionID); err != nil {
		return kit.Empty{}, err
	}
	return kit.Empty{}, publishAccount(ctx, a, PasswordChanged)
}

// ForgotPassword mails a password reset link to the account of an address,
// if it has one. The answer is the same when it has none.
func ForgotPassword(ctx context.Context, in EmailInput) (SentOutput, error) {
	email := NormalizeEmail(in.Email)
	if !ValidEmail(email) {
		return SentOutput{}, wire.Invalid("email", "email", "must be an email address")
	}
	out := SentOutput{Status: "sent"}
	a, found, err := accountByEmail(ctx, email)
	if err != nil || !found {
		return out, err
	}
	secret, err := issue(ctx, a.ID, PurposeReset, ResetFor)
	if err != nil {
		return SentOutput{}, err
	}
	_, err = notify.SendAPI.Call(ctx, notify.SendInput{
		Template: notify.TemplateResetPassword, To: a.Email, Name: a.Name, Locale: a.locale(), Data: map[string]string{"token": secret},
	})
	return out, err
}

// ResetInput is the secret of a reset link and the new password.
type ResetInput struct {
	Token    string `json:"token" validate:"required,maxlen=128"`
	Password string `json:"password" validate:"required,minlen=10,maxlen=128"`
}

// ResetPassword chooses a new password with the link of a reset mail. It
// unlocks a locked account and verifies an unverified one — the link proved
// the address — signs out every session, and signs this one in.
func ResetPassword(ctx context.Context, in ResetInput) (UserOutput, error) {
	hash, err := hashPassword(in.Password) // before the link is used up
	if err != nil {
		return UserOutput{}, err
	}
	t, err := redeem(ctx, in.Token, PurposeReset)
	if err != nil {
		return UserOutput{}, err
	}
	now := wire.Now(ctx)
	a, err := Accounts.Update(ctx, t.UserID, func(a *Account) error {
		a.PasswordHash, a.FailedLogins, a.UpdatedAt = hash, 0, now
		return nil
	})
	if err != nil {
		return UserOutput{}, err
	}
	if a.Status != Active {
		if a, err = AccountLifecycle.Fire(ctx, a.ID, "reset"); err != nil {
			return UserOutput{}, err
		}
	}
	if err := revokeSessions(ctx, a.ID, ""); err != nil {
		return UserOutput{}, err
	}
	if err := openSession(ctx, a.ID); err != nil {
		return UserOutput{}, err
	}
	if err := publishAccount(ctx, a, PasswordChanged); err != nil {
		return UserOutput{}, err
	}
	return UserOutput{User: a.user()}, nil
}

// SessionView is one of the caller's sessions.
type SessionView struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	UserAgent  string    `json:"userAgent"`
	IP         string    `json:"ip"`
	Current    bool      `json:"current"`
}

// SessionsOutput lists the caller's sessions, most recently used first.
type SessionsOutput struct {
	Sessions []SessionView `json:"sessions"`
}

// ListSessions lists the caller's live sessions: where they are signed in.
func ListSessions(ctx context.Context, _ kit.Empty) (SessionsOutput, error) {
	uid, p, err := caller(ctx)
	if err != nil {
		return SessionsOutput{}, err
	}
	sessions, err := Sessions.Find(ctx, "user", uid)
	if err != nil {
		return SessionsOutput{}, err
	}
	now := wire.Now(ctx)
	out := SessionsOutput{Sessions: []SessionView{}}
	for _, s := range sessions {
		if !now.Before(s.ExpiresAt) {
			continue
		}
		out.Sessions = append(out.Sessions, SessionView{
			ID: s.ID, CreatedAt: s.CreatedAt, LastSeenAt: s.LastSeenAt, UserAgent: s.UserAgent, IP: s.IP, Current: s.ID == p.SessionID,
		})
	}
	slices.SortFunc(out.Sessions, func(a, b SessionView) int {
		return cmp.Or(b.LastSeenAt.Compare(a.LastSeenAt), cmp.Compare(a.ID, b.ID))
	})
	return out, nil
}

// SessionID addresses one session.
type SessionID struct {
	ID string `path:"id"`
}

// RevokeSession signs one of the caller's sessions out. Revoking the current
// one also forgets its cookie.
func RevokeSession(ctx context.Context, in SessionID) (kit.Empty, error) {
	uid, p, err := caller(ctx)
	if err != nil {
		return kit.Empty{}, err
	}
	s, err := Sessions.Get(ctx, in.ID)
	if wire.Is(err, kit.CodeNotFound) || (err == nil && s.UserID != uid) {
		return kit.Empty{}, kit.NotFound("no such session")
	}
	if err != nil {
		return kit.Empty{}, err
	}
	if err := Sessions.Delete(ctx, s.ID); err != nil && !wire.Is(err, kit.CodeNotFound) {
		return kit.Empty{}, err
	}
	if s.ID == p.SessionID {
		kit.ClearCookie(ctx, CookieName)
	}
	return kit.Empty{}, nil
}
