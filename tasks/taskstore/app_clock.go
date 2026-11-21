// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package taskstore

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
)

// appClock is AppClock's own state: the clock is the framework's.
type appClock struct {
}

// newAppClock is NewAppClock's body.
func newAppClock() *AppClock {
	return &AppClock{appClock{}}
}

// now is AppClock.Now's body: the framework's time, which a test moves, in
// UTC and to the millisecond, as every timestamp the product keeps.
func (a *AppClock) now(ctx context.Context) (r0 time.Time, err error) {
	return app.Now(ctx).UTC().Truncate(time.Millisecond), nil
}
