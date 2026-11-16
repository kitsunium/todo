// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:b86d0cd87931d5d99bbe4360b3b902a3632eea398c69f6b97c0556cdb2a00c3c

// Package vault — The accounts, the sessions and the one-time links in the
// framework's stores, the mails of an account, the directory the rest of the
// product asks, the passwords stretched, the secrets drawn.
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
	Verified     bool      `json:"verified" app:"public"`
	FailedLogins int       `json:"failedLogins" app:"public"`
	LockedUntil  time.Time `json:"lockedUntil" app:"public"`
	CreatedAt    time.Time `json:"createdAt" app:"public"`
	UpdatedAt    time.Time `json:"updatedAt" app:"public"`
}

// LinkRecord — A one-time link as the store keeps it, field for field.
type LinkRecord struct {
	ID        string    `json:"id" app:"public"`
	Hash      string    `json:"hash" app:"secret"`
	Purpose   string    `json:"purpose" app:"public"`
	UserID    string    `json:"userId" app:"public"`
	CreatedAt time.Time `json:"createdAt" app:"public"`
	ExpiresAt time.Time `json:"expiresAt" app:"public"`
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

// ByEmail — Which address is asked about.
type ByEmail struct {
	Email string `json:"email" app:"personal"`
}

// ByID — Which member is asked about.
type ByID struct {
	ID string `json:"id" app:"public"`
}

// Member — A member as the rest of the product may know them; Found is false
// when there is none.
type Member struct {
	Found bool   `json:"found" app:"public"`
	ID    string `json:"id" app:"public"`
	Email string `json:"email" app:"personal"`
	Name  string `json:"name" app:"personal"`
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

// OneTimeLinks is Links@v1 on the store "links".
type OneTimeLinks struct {
	oneTimeLinks // its own state, hand-written
}

// NewOneTimeLinks builds OneTimeLinks: its body is the hand-written newOneTimeLinks.
func NewOneTimeLinks() *OneTimeLinks {
	return newOneTimeLinks()
}

// ByHash is identity/port/Links@v1's.
func (a *OneTimeLinks) ByHash(ctx context.Context, hash string) (account.Link, bool, error) {
	return a.byHash(ctx, hash)
}

// DeleteOf is identity/port/Links@v1's.
func (a *OneTimeLinks) DeleteOf(ctx context.Context, userID string, purpose string) error {
	return a.deleteOf(ctx, userID, purpose)
}

// NewID is identity/port/Links@v1's.
func (a *OneTimeLinks) NewID(ctx context.Context) (string, error) {
	return a.newID(ctx)
}

// Save is identity/port/Links@v1's.
func (a *OneTimeLinks) Save(ctx context.Context, l account.Link) error {
	return a.save(ctx, l)
}

var _ access.LinksV1 = (*OneTimeLinks)(nil)

// Postman is Courier@v1; it writes the mails of an account, in French, and
// sends them through the product's mailer.
type Postman struct {
	postman // its own state, hand-written
}

// NewPostman builds Postman: its body is the hand-written newPostman.
func NewPostman(baseUrlSetting *app.Setting[string]) *Postman {
	return newPostman(baseUrlSetting)
}

// AlreadyMember is identity/port/Courier@v1's.
func (a *Postman) AlreadyMember(ctx context.Context, to account.Account) error {
	return a.alreadyMember(ctx, to)
}

// Reset is identity/port/Courier@v1's.
func (a *Postman) Reset(ctx context.Context, to account.Account, secret string) error {
	return a.reset(ctx, to, secret)
}

// Verification is identity/port/Courier@v1's.
func (a *Postman) Verification(ctx context.Context, to account.Account, secret string) error {
	return a.verification(ctx, to, secret)
}

var _ access.CourierV1 = (*Postman)(nil)

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

// DeleteOf is identity/port/Sessions@v1's.
func (self *Sessions) DeleteOf(ctx context.Context, userID string, keep string) error {
	return self.deleteOf(ctx, userID, keep)
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
var Service = app.NewService("vault", "The accounts, the sessions and the one-time links in the framework's stores, the mails of an account, the directory the rest of the product asks, the passwords stretched, the secrets drawn.")

// BaseUrlSetting — Where members reach the product, for the links of its
// mails.
var BaseUrlSetting = Service.Setting[string]("base-url", "http://localhost:4000")

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

// LinksStore — Every one-time link, keyed by its id; one is deleted a day
// after it expired.
var LinksStore = Service.Store("links", func(e LinkRecord) string { return e.ID },
	app.DeleteAfter[LinkRecord](time.Duration(86400000000000) /* 24h */, func(e LinkRecord) (time.Time, bool) { return e.ExpiresAt, !e.ExpiresAt.IsZero() }),
	app.Unique[LinkRecord]("by-hash", func(e LinkRecord) string { return e.Hash }),
)

// MemberQuery — The directory: a member's address and name by their id, for
// the components that write to them.
var MemberQuery = Service.Query("member", Directory)

// Directory handles the query "member": its body is the hand-written directory.
func Directory(ctx context.Context, in ByID) (Member, error) {
	return directory(ctx, in)
}

// MemberByEmailQuery — The directory, by address: whose account an address
// is.
var MemberByEmailQuery = Service.Query("member-by-email", Lookup)

// Lookup handles the query "member-by-email": its body is the hand-written lookup.
func Lookup(ctx context.Context, in ByEmail) (Member, error) {
	return lookup(ctx, in)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts vault calls it
// before its App runs.
func Wire() {

}
