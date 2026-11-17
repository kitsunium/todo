// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:fff3fa1ce1d5d0ffb043be2f4bd6a1b7fa06f149490a7503c5b7744197cb6684

// Package account — The account, its lock after wrong passwords, and the
// session.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package account

import (
	"time"

	"github.com/kitsunium/sdk/pkg/v1/errs"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 6, 1, 255), "NOT_IMPLEMENTED", "not implemented yet", "account: a hand-written body is still a skeleton")

var (
	// ErrBadLink is "this link is invalid or has expired: ask for a new one".
	ErrBadLink = errs.New(errs.Pack(64, 6, 1, 9), "BAD_LINK", "this link is invalid or has expired: ask for a new one", "identity/error/ErrBadLink")
	// ErrEmailTaken is what the store answers when two sign-ups of one address
	// race: the use case answers the second like any taken address, and never says
	// it.
	ErrEmailTaken = errs.New(errs.Pack(64, 6, 1, 4), "EMAIL_TAKEN", "this address already has an account", "identity/error/ErrEmailTaken")
	// ErrInvalidCredentials is "the email address or the password is not right".
	ErrInvalidCredentials = errs.New(errs.Pack(64, 6, 1, 5), "INVALID_CREDENTIALS", "the email address or the password is not right", "identity/error/ErrInvalidCredentials")
	// ErrInvalidEmail is "this is not an email address".
	ErrInvalidEmail = errs.New(errs.Pack(64, 6, 1, 1), "INVALID_EMAIL", "this is not an email address", "identity/error/ErrInvalidEmail")
	// ErrInvalidLocale is "the language is fr or en".
	ErrInvalidLocale = errs.New(errs.Pack(64, 6, 1, 10), "INVALID_LOCALE", "the language is fr or en", "identity/error/ErrInvalidLocale")
	// ErrInvalidTimeZone is "the time zone is a name of the IANA database, like
	// Europe/Paris".
	ErrInvalidTimeZone = errs.New(errs.Pack(64, 6, 1, 11), "INVALID_TIME_ZONE", "the time zone is a name of the IANA database, like Europe/Paris", "identity/error/ErrInvalidTimeZone")
	// ErrLocked is "too many wrong passwords: the account is locked for 15
	// minutes".
	ErrLocked = errs.New(errs.Pack(64, 6, 1, 6), "LOCKED", "too many wrong passwords: the account is locked for 15 minutes", "identity/error/ErrLocked")
	// ErrNameRequired is "an account needs a name".
	ErrNameRequired = errs.New(errs.Pack(64, 6, 1, 3), "NAME_REQUIRED", "an account needs a name", "identity/error/ErrNameRequired")
	// ErrNoSession is "the session has expired: sign in again".
	ErrNoSession = errs.New(errs.Pack(64, 6, 1, 7), "NO_SESSION", "the session has expired: sign in again", "identity/error/ErrNoSession")
	// ErrUnverified is "confirm your email address first: the link is in your
	// inbox".
	ErrUnverified = errs.New(errs.Pack(64, 6, 1, 8), "UNVERIFIED", "confirm your email address first: the link is in your inbox", "identity/error/ErrUnverified")
	// ErrWeakPassword is "a password is 10 to 128 characters long".
	ErrWeakPassword = errs.New(errs.Pack(64, 6, 1, 2), "WEAK_PASSWORD", "a password is 10 to 128 characters long", "identity/error/ErrWeakPassword")
)

const (
	// LocaleEN is English.
	LocaleEN string = "en"
	// LocaleFR is French, the language of an account that chose none.
	LocaleFR string = "fr"
	// LockFor is how long a locked account refuses every password.
	LockFor time.Duration = 15 * time.Minute
	// MaxFailedLogins is how many wrong passwords in a row lock an account.
	MaxFailedLogins int = 5
	// MinPassword is the shortest password an account takes.
	MinPassword int = 10
	// PurposeReset is the purpose of a link that chooses a new password.
	PurposeReset string = "reset"
	// PurposeVerify is the purpose of a link that proves an address.
	PurposeVerify string = "verify"
	// ResetFor is how long a link that chooses a new password is good for.
	ResetFor time.Duration = time.Hour
	// SessionFor is how long a session lives after it was last seen.
	SessionFor time.Duration = 30 * 24 * time.Hour
	// VerifyFor is how long a link that proves an address is good for.
	VerifyFor time.Duration = 24 * time.Hour
)

// Account — A member of the team, who signs in with an address and a
// password.
type Account struct {
	ID    string `json:"id" app:"public"`
	Email string `json:"email" app:"personal"`
	Name  string `json:"name" app:"personal"`
	// Locale is the language the member reads their mails in, fr or en; an account
	// that has none reads French.
	Locale string `json:"locale" app:"public"`
	// TimeZone is the IANA zone the member's mails write times in; empty reads
	// UTC.
	TimeZone string `json:"timeZone" app:"personal"`
	// PasswordHash is the password, stretched: never the password, never answered.
	PasswordHash string `json:"-" app:"secret"`
	// Verified says the member proved the address is theirs; an account that is
	// not signs nobody in.
	Verified bool `json:"verified" app:"public"`
	// FailedLogins counts the wrong passwords in a row.
	FailedLogins int `json:"-" app:"public"`
	// LockedUntil is when the lock of too many wrong passwords ends; the zero time
	// when there is none.
	LockedUntil time.Time `json:"-" app:"public"`
	CreatedAt   time.Time `json:"createdAt" app:"public"`
	UpdatedAt   time.Time `json:"updatedAt" app:"public"`
}

// Locked reports whether the account refuses every password at now.
func (a *Account) Locked(now time.Time) bool {
	return a.locked(now)
}

// Link — A one-time link mailed to a member, to prove an address or to
// choose a new password.
type Link struct {
	ID string `json:"id" app:"public"`
	// Hash is the SHA-256 of the link's secret, which only the mail carries.
	Hash string `json:"-" app:"secret"`
	// Purpose is verify or reset: a link does one thing.
	Purpose   string    `json:"purpose" app:"public"`
	UserID    string    `json:"userId" app:"public"`
	CreatedAt time.Time `json:"createdAt" app:"public"`
	ExpiresAt time.Time `json:"expiresAt" app:"public"`
}

// Session — A member signed in, from one browser or one client.
type Session struct {
	ID     string `json:"id" app:"public"`
	UserID string `json:"userId" app:"public"`
	// TokenHash is the SHA-256 of the session's secret: the secret itself is kept
	// nowhere.
	TokenHash string `json:"-" app:"secret"`
	// UserAgent is the browser or the client that signed in, as it named itself.
	UserAgent string `json:"userAgent" app:"personal"`
	// IP is the address the sign-in came from.
	IP         string    `json:"ip" app:"personal"`
	CreatedAt  time.Time `json:"createdAt" app:"public"`
	LastSeenAt time.Time `json:"lastSeenAt" app:"public"`
	ExpiresAt  time.Time `json:"expiresAt" app:"public"`
}

// NormalizeEmail is an address as the product keeps it, trimmed and in lower
// case, or ErrInvalidEmail.
func NormalizeEmail(s string) (string, error) {
	return normalizeEmail(s)
}
