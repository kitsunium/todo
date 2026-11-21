// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package linkstore

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/contacts/circle"
	"github.com/kitsunium/todo/contacts/link"
)

// announcer is Announcer's own state: the topic is the package's, it holds nothing.
type announcer struct {
}

// newAnnouncer is NewAnnouncer's body.
func newAnnouncer() *Announcer {
	return &Announcer{announcer{}}
}

// tell is Announcer.Tell's body: what the actor did to the link is
// published on the topic "events", for the member at its other end.
func (a *Announcer) tell(ctx context.Context, kind string, l link.Link, actor circle.Person) (err error) {
	to, _ := l.Other(actor.ID)
	return EventsTopic.Publish(ctx, Event{
		ID: app.NewID("event"), Kind: kind, LinkID: l.ID, UserID: to, ActorID: actor.ID, ActorName: actor.Name, At: l.UpdatedAt,
	})
}
