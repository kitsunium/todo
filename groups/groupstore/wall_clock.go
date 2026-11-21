// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package groupstore

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
)

// wallClock is WallClock's own state: the clock is the framework's.
type wallClock struct {
}

// newWallClock is NewWallClock's body.
func newWallClock() *WallClock {
	return &WallClock{wallClock{}}
}

// now is WallClock.Now's body: the framework's time, which a test moves, in
// UTC and to the millisecond.
func (a *WallClock) now(ctx context.Context) (r0 time.Time, err error) {
	return app.Now(ctx).UTC().Truncate(time.Millisecond), nil
}
