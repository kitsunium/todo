// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package outbox

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
)

// register is Register's own state: the store is the package's, it holds nothing.
type register struct {
}

// newRegister is NewRegister's body.
func newRegister() *Register {
	return &Register{register{}}
}

// mark is Register.Mark's body.
func (a *Register) mark(ctx context.Context, key string) (err error) {
	return ActedStore.Put(ctx, Acted{Key: key, At: app.Now(ctx).UTC().Truncate(time.Millisecond)})
}

// seen is Register.Seen's body. It filters instead of reading the key: an
// event not yet mailed is the usual answer, never a failed read.
func (a *Register) seen(ctx context.Context, key string) (r0 bool, err error) {
	found, err := ActedStore.Filter(ctx, func(r Acted) bool { return r.Key == key })
	return len(found) > 0, err
}
