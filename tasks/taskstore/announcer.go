// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package taskstore

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/tasks/task"
)

// announcer is Announcer's own state: the topic is the package's, it holds nothing.
type announcer struct {
}

// newAnnouncer is NewAnnouncer's body.
func newAnnouncer() *Announcer {
	return &Announcer{announcer{}}
}

// record is Announcer.Record's body: what happened to the task is published
// on the topic "events", with an id of its own and the date the task was
// last written. Whoever listens hears of it later, at least once.
func (a *Announcer) record(ctx context.Context, kind string, t task.Task) (err error) {
	return EventsTopic.Publish(ctx, Event{
		ID: app.NewID("event"), Kind: kind, TaskID: t.ID, OwnerID: t.OwnerID, Title: t.Title, Assignee: t.Assignee, At: t.UpdatedAt,
	})
}
