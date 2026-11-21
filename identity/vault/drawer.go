// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// secretBytes is the size of a secret: 256 random bits.
const secretBytes = 32

// drawer is Drawer's own state: it holds nothing.
type drawer struct {
}

// newDrawer is NewDrawer's body.
func newDrawer() *Drawer {
	return &Drawer{drawer{}}
}

// digest is Drawer.Digest's body: the SHA-256 of the secret, in
// hexadecimal. A secret has 256 bits of entropy, so a fast hash is the
// right one: there is nothing to guess.
func (a *Drawer) digest(ctx context.Context, secret string) (r0 string, err error) {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:]), nil
}

// newBody is Drawer.New's body: a secret and its digest.
func (a *Drawer) newBody(ctx context.Context) (r0 string, r1 string, err error) {
	b := make([]byte, secretBytes)
	_, _ = rand.Read(b) // crypto/rand.Read never fails: it crashes the program rather than return predictable bytes.
	r0 = base64.RawURLEncoding.EncodeToString(b)
	r1, err = a.digest(ctx, r0)
	return r0, r1, err
}
