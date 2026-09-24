package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/internal/wire"
)

// Purpose says what a one-time link does.
type Purpose string

// The purposes of a one-time link.
const (
	// PurposeVerify links prove an address and sign its account in.
	PurposeVerify Purpose = "verify"
	// PurposeReset links choose a new password.
	PurposeReset Purpose = "reset"
)

// How long a link works.
const (
	VerifyFor = 24 * time.Hour
	ResetFor  = time.Hour
)

// TokenStatus is where a one-time link stands.
type TokenStatus string

// The states of a one-time link.
const (
	Issued  TokenStatus = "issued"
	Used    TokenStatus = "used"
	Expired TokenStatus = "expired"
)

// Token is a one-time link, as the product remembers it: the SHA-256 of the
// secret the link carries, never the secret. A stolen copy of the store
// opens nothing.
type Token struct {
	ID        string      `json:"id"`
	Hash      string      `json:"hash" kit:"secret"`
	Purpose   Purpose     `json:"purpose"`
	UserID    string      `json:"userId"`
	Status    TokenStatus `json:"status"`
	CreatedAt time.Time   `json:"createdAt"`
	ExpiresAt time.Time   `json:"expiresAt"`
	UsedAt    *time.Time  `json:"usedAt,omitempty"`
}

// Tokens keeps the one-time links, keyed by ID and found by the hash of
// their secret.
var Tokens = Service.Store("tokens", func(t Token) string { return t.ID },
	kit.Unique("hash", func(t Token) string { return t.Hash }))

// TokenLifecycle is the life of a one-time link: used once, or expired by
// its own guard when its deadline passes.
var TokenLifecycle = Service.Workflow("tokens", Tokens, func(t *Token) *TokenStatus { return &t.Status }).
	Initial(Issued).
	On("use", Issued, Used).
	When("expire", Issued, Expired, pastDeadline).
	OnEnter(Used, stampUse)

// pastDeadline holds once a link's deadline passed.
func pastDeadline(t Token, now time.Time) bool { return !now.Before(t.ExpiresAt) }

// stampUse records when a link was used.
func stampUse(ctx context.Context, t *Token) error {
	now := wire.Now(ctx)
	t.UsedAt = &now
	return nil
}

// secretBytes is the entropy of every secret the product hands out: a link's,
// a session's.
const secretBytes = 32

// newSecret returns a fresh secret — 32 random bytes, base64url, what a link
// or a cookie carries — and the digest the product keeps instead.
func newSecret() (secret, digest string) {
	b := make([]byte, secretBytes)
	_, _ = rand.Read(b) // crypto/rand.Read never fails: it crashes the program rather than return predictable bytes.
	secret = base64.RawURLEncoding.EncodeToString(b)
	return secret, digestOf(secret)
}

// digestOf is the SHA-256 of a secret, in hexadecimal. A secret has 256 bits
// of entropy, so a fast hash is the right one: there is nothing to guess.
func digestOf(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// wellFormed reports whether s has the shape of a secret the product made.
func wellFormed(s string) bool {
	b, err := base64.RawURLEncoding.DecodeString(s)
	return err == nil && len(b) == secretBytes
}

// sameDigest compares two digests in constant time.
func sameDigest(a, b string) bool { return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1 }

// errBadLink answers a link that is unknown, used, expired or meant for
// something else — the same answer for all four, so a link reveals nothing.
func errBadLink() error {
	return &kit.Error{Status: http.StatusBadRequest, Code: "invalid_token", Message: "this link is invalid or has expired: ask for a new one"}
}

// issue creates a one-time link for the user and returns its secret, which
// only the mail will carry.
func issue(ctx context.Context, userID string, p Purpose, valid time.Duration) (string, error) {
	secret, digest := newSecret()
	now := wire.Now(ctx)
	_, err := TokenLifecycle.Start(ctx, Token{
		ID: kit.NewID("token"), Hash: digest, Purpose: p, UserID: userID, CreatedAt: now, ExpiresAt: now.Add(valid),
	})
	return secret, err
}

// redeem uses a link: it must exist, serve purpose, be unused and before its
// deadline. It is used up even if what follows fails — a link is single use,
// not single success.
func redeem(ctx context.Context, secret string, p Purpose) (Token, error) {
	if !wellFormed(secret) {
		return Token{}, errBadLink()
	}
	digest := digestOf(secret)
	found, err := Tokens.Find(ctx, "hash", digest)
	if err != nil {
		return Token{}, err
	}
	if len(found) == 0 {
		return Token{}, errBadLink()
	}
	t := found[0]
	if !sameDigest(t.Hash, digest) || t.Purpose != p || t.Status != Issued || !wire.Now(ctx).Before(t.ExpiresAt) {
		return Token{}, errBadLink()
	}
	used, err := TokenLifecycle.Fire(ctx, t.ID, "use")
	if wire.Is(err, kit.CodeConflict) || wire.Is(err, kit.CodeNotFound) {
		return Token{}, errBadLink() // another request used it first
	}
	return used, err
}
