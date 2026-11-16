// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"
	"sync"

	"github.com/kitsunium/sdk/pkg/v1/crypto/password"
)

// stretcher is Stretcher's own state: the hash an unknown address's
// password is checked against, made once.
type stretcher struct {
	once  sync.Once
	decoy string
}

// newStretcher is NewStretcher's body.
func newStretcher() *Stretcher {
	return &Stretcher{}
}

// hash is Stretcher.Hash's body: the SDK's PBKDF2-SHA256, in PHC form.
func (a *Stretcher) hash(ctx context.Context, pw string) (r0 string, err error) {
	return password.Hash(password.PBKDF2SHA256, []byte(pw))
}

// verify is Stretcher.Verify's body. Against an empty hash it does the same
// work on a hash nobody's password makes, and answers false: a caller with
// no account to check spends the time it would have spent with one.
func (a *Stretcher) verify(ctx context.Context, pw string, hash string) (r0 bool, err error) {
	if hash == "" {
		a.once.Do(func() { a.decoy, _ = password.Hash(password.PBKDF2SHA256, []byte("no account has this one")) })
		_, _ = password.Verify([]byte(pw), a.decoy)
		return false, nil
	}
	return password.Verify([]byte(pw), hash)
}
