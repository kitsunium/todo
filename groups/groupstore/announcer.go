// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package groupstore

import (
	"context"
	"slices"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/groups/roster"
	"github.com/kitsunium/todo/groups/team"
)

// announcer is Announcer's own state: the topic is the package's, it holds nothing.
type announcer struct {
}

// newAnnouncer is NewAnnouncer's body.
func newAnnouncer() *Announcer {
	return &Announcer{announcer{}}
}

// tell is Announcer.Tell's body: what the actor did in the group is published
// on the topic "events", one event for each member told. The actor is never
// told of what they did themselves, and nobody twice.
func (a *Announcer) tell(ctx context.Context, kind string, g team.Group, actor roster.Person, to []string) (err error) {
	at := app.Now(ctx).UTC().Truncate(time.Millisecond)
	var told []string
	for _, id := range to {
		if id == "" || id == actor.ID || slices.Contains(told, id) {
			continue
		}
		told = append(told, id)
		if err := EventsTopic.Publish(ctx, Event{
			ID: app.NewID("event"), Kind: kind, GroupID: g.ID, GroupName: g.Name,
			UserID: id, ActorID: actor.ID, ActorName: actor.Name, At: at,
		}); err != nil {
			return err
		}
	}
	return nil
}
