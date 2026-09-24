package identity

import (
	"context"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/sdk/pkg/v1/logger"
	"github.com/kitsunium/todo/internal/wire"
)

// SessionReaper is identity's own event loop, written by hand: a ticker and
// a select, the way a Go program waits. kit starts it with the app, cancels
// it on shutdown, restarts it after a failure, and reads its select to draw
// what it waits on.
var SessionReaper = Service.Go("session-reaper", ReapSessions)

// reapEvery is the reaper's period.
const reapEvery = time.Minute

// spentFor is how long a used or expired link is kept, for the record.
const spentFor = 24 * time.Hour

// ReapSessions deletes, every minute, the sessions that expired and the
// one-time links spent more than a day ago. An expired session already
// signs no one in; the reaper keeps the store from growing forever.
func ReapSessions(ctx context.Context) error {
	t := time.NewTicker(reapEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			if err := Reap(ctx); err != nil {
				return err // kit restarts the loop after a backoff
			}
		}
	}
}

// Reap is one pass of the reaper.
func Reap(ctx context.Context) error {
	now := wire.Now(ctx)
	expired, err := Sessions.Filter(ctx, func(s Session) bool { return !now.Before(s.ExpiresAt) })
	if err != nil {
		return err
	}
	for _, s := range expired {
		if err := Sessions.Delete(ctx, s.ID); err != nil && !wire.Is(err, kit.CodeNotFound) {
			return err
		}
	}
	spent, err := Tokens.Filter(ctx, func(t Token) bool {
		return t.Status != Issued && now.Sub(t.ExpiresAt) > spentFor
	})
	if err != nil {
		return err
	}
	for _, t := range spent {
		if err := Tokens.Delete(ctx, t.ID); err != nil && !wire.Is(err, kit.CodeNotFound) {
			return err
		}
	}
	if len(expired)+len(spent) > 0 {
		logger.Info(ctx, kit.Log(ctx), "reaped", logger.Int("sessions", len(expired)), logger.Int("links", len(spent)))
	}
	return nil
}
