// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package taskstore

import (
	"context"
	"slices"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/groups/groupsapi"
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
// on the topic "events", with an id of its own, the date the task was last
// written and the members to tell: everyone who sees the task. Whoever listens hears of it later, at least once.
func (a *Announcer) record(ctx context.Context, kind string, t task.Task) (err error) {
	audience, err := audienceOf(ctx, t)
	if err != nil {
		return err
	}
	return EventsTopic.Publish(ctx, Event{
		ID: app.NewID("event"), Kind: kind, TaskID: t.ID, OwnerID: t.OwnerID, Audience: audience, Title: t.Title, At: t.UpdatedAt,
	})
}

// tell is Announcer.Tell's body: what happened to the task is published for
// one member alone.
func (a *Announcer) tell(ctx context.Context, kind string, t task.Task, to string) (err error) {
	return EventsTopic.Publish(ctx, Event{
		ID: app.NewID("event"), Kind: kind, TaskID: t.ID, OwnerID: t.OwnerID, Audience: []string{to}, Title: t.Title, At: t.UpdatedAt,
	})
}

// audienceOf is everyone who sees a task: its owner, the members it is shared
// with, its assignee, and the members of its group, the groups' query "members"
// naming them; each once, the owner first.
func audienceOf(ctx context.Context, t task.Task) ([]string, error) {
	out := []string{t.OwnerID}
	add := func(ids ...string) {
		for _, id := range ids {
			if id != "" && !slices.Contains(out, id) {
				out = append(out, id)
			}
		}
	}
	add(t.SharedWith...)
	add(t.AssigneeID)
	if t.GroupID != "" {
		members, err := groupsapi.MembersQuery.Ask(ctx, groupsapi.WhichGroup{Group: t.GroupID})
		if err != nil {
			return nil, err
		}
		add(members.Users...)
	}
	return out, nil
}
