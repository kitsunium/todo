// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package taskstore

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
)

// onShelfTransition runs after every transition of Shelf: a task archived
// by time is announced on the topic "events" like one a member archived,
// so the feed tells it.
func onShelfTransition(ctx context.Context, c app.Change[Record, ShelfState]) error {
	if c.Event != ShelfOnAutoArchive {
		return nil
	}
	return EventsTopic.Publish(ctx, Event{
		ID: app.NewID("event"), Kind: "archive", TaskID: c.Entity.ID, OwnerID: c.Entity.OwnerID,
		Title: c.Entity.Title, Assignee: c.Entity.Assignee, At: app.Now(ctx).UTC(),
	})
}
