// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:b8c6fb4903e1368054d808d160797f66a4b4f241102b577f4a124db5ec43a883

// Package vault — The accounts and the sessions in the framework's stores,
// the passwords stretched, the secrets drawn.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package vault

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/identity/access"
	"github.com/kitsunium/todo/identity/account"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 6, 3, 255), "NOT_IMPLEMENTED", "not implemented yet", "vault: a hand-written body is still a skeleton")

// AccountRecord — An account as the store keeps it, field for field.
type AccountRecord struct {
	ID           string    `json:"id" app:"public"`
	Email        string    `json:"email" app:"personal,subject"`
	Name         string    `json:"name" app:"personal"`
	PasswordHash string    `json:"passwordHash" app:"secret"`
	FailedLogins int       `json:"failedLogins" app:"public"`
	LockedUntil  time.Time `json:"lockedUntil" app:"public"`
	CreatedAt    time.Time `json:"createdAt" app:"public"`
	UpdatedAt    time.Time `json:"updatedAt" app:"public"`
}

// SessionRecord — A session as the store keeps it, field for field.
type SessionRecord struct {
	ID         string    `json:"id" app:"public"`
	UserID     string    `json:"userId" app:"public"`
	TokenHash  string    `json:"tokenHash" app:"secret"`
	CreatedAt  time.Time `json:"createdAt" app:"public"`
	LastSeenAt time.Time `json:"lastSeenAt" app:"public"`
	ExpiresAt  time.Time `json:"expiresAt" app:"public"`
}

// Accounts is Accounts@v1 on the store "accounts".
type Accounts struct {
	accounts // its own state, hand-written
}

// NewAccounts builds Accounts: its body is the hand-written newAccounts.
func NewAccounts() *Accounts {
	return newAccounts()
}

// ByEmail is identity/port/Accounts@v1's.
func (self *Accounts) ByEmail(ctx context.Context, email string) (account.Account, bool, error) {
	return self.byEmail(ctx, email)
}

// Get is identity/port/Accounts@v1's.
func (self *Accounts) Get(ctx context.Context, id string) (account.Account, bool, error) {
	return self.get(ctx, id)
}

// NewID is identity/port/Accounts@v1's.
func (self *Accounts) NewID(ctx context.Context) (string, error) {
	return self.newID(ctx)
}

// Save is identity/port/Accounts@v1's.
func (self *Accounts) Save(ctx context.Context, a account.Account) error {
	return self.save(ctx, a)
}

var _ access.AccountsV1 = (*Accounts)(nil)

// Drawer is Secrets@v1, 256 random bits a secret, SHA-256 its digest.
type Drawer struct {
	drawer // its own state, hand-written
}

// NewDrawer builds Drawer: its body is the hand-written newDrawer.
func NewDrawer() *Drawer {
	return newDrawer()
}

// Digest is identity/port/Secrets@v1's.
func (a *Drawer) Digest(ctx context.Context, secret string) (string, error) {
	return a.digest(ctx, secret)
}

// New is identity/port/Secrets@v1's.
func (a *Drawer) New(ctx context.Context) (string, string, error) {
	return a.newBody(ctx)
}

var _ access.SecretsV1 = (*Drawer)(nil)

// Sessions is Sessions@v1 on the store "sessions".
type Sessions struct {
	sessions // its own state, hand-written
}

// NewSessions builds Sessions: its body is the hand-written newSessions.
func NewSessions() *Sessions {
	return newSessions()
}

// ByTokenHash is identity/port/Sessions@v1's.
func (self *Sessions) ByTokenHash(ctx context.Context, hash string) (account.Session, bool, error) {
	return self.byTokenHash(ctx, hash)
}

// Delete is identity/port/Sessions@v1's.
func (self *Sessions) Delete(ctx context.Context, id string) error {
	return self.deleteBody(ctx, id)
}

// NewID is identity/port/Sessions@v1's.
func (self *Sessions) NewID(ctx context.Context) (string, error) {
	return self.newID(ctx)
}

// Save is identity/port/Sessions@v1's.
func (self *Sessions) Save(ctx context.Context, s account.Session) error {
	return self.save(ctx, s)
}

var _ access.SessionsV1 = (*Sessions)(nil)

// Stretcher is Passwords@v1 on the SDK's PBKDF2-SHA256.
type Stretcher struct {
	stretcher // its own state, hand-written
}

// NewStretcher builds Stretcher: its body is the hand-written newStretcher.
func NewStretcher() *Stretcher {
	return newStretcher()
}

// Hash is identity/port/Passwords@v1's.
func (a *Stretcher) Hash(ctx context.Context, password string) (string, error) {
	return a.hash(ctx, password)
}

// Verify is identity/port/Passwords@v1's.
func (a *Stretcher) Verify(ctx context.Context, password string, hash string) (bool, error) {
	return a.verify(ctx, password, hash)
}

var _ access.PasswordsV1 = (*Stretcher)(nil)

// WallClock is Clock@v1 on the framework's clock, UTC, to the millisecond.
type WallClock struct {
	wallClock // its own state, hand-written
}

// NewWallClock builds WallClock: its body is the hand-written newWallClock.
func NewWallClock() *WallClock {
	return newWallClock()
}

// Now is identity/port/Clock@v1's.
func (a *WallClock) Now(ctx context.Context) (time.Time, error) {
	return a.now(ctx)
}

var _ access.ClockV1 = (*WallClock)(nil)

// Service is vault's framework service: what it declares, drawn and run.
var Service = app.NewService("vault", "The accounts and the sessions in the framework's stores, the passwords stretched, the secrets drawn.")

// AccountsStore — Every account, keyed by its id; one untouched for three
// years is erased.
var AccountsStore = Service.Store("accounts", func(e AccountRecord) string { return e.ID },
	app.EraseAfter[AccountRecord](time.Duration(94608000000000000) /* 26280h */, func(e AccountRecord) (time.Time, bool) { return e.UpdatedAt, !e.UpdatedAt.IsZero() }),
	app.Unique[AccountRecord]("by-email", func(e AccountRecord) string { return e.Email }),
)

// SessionsStore — Every session, keyed by its id; one is deleted a day after
// it expired.
var SessionsStore = Service.Store("sessions", func(e SessionRecord) string { return e.ID },
	app.DeleteAfter[SessionRecord](time.Duration(86400000000000) /* 24h */, func(e SessionRecord) (time.Time, bool) { return e.ExpiresAt, !e.ExpiresAt.IsZero() }),
	app.Unique[SessionRecord]("by-token", func(e SessionRecord) string { return e.TokenHash }),
)

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts vault calls it
// before its App runs.
func Wire() {

}
