package activity

import (
	"cmp"
	"context"
	"slices"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/contacts"
	"github.com/kitsunium/todo/groups"
	"github.com/kitsunium/todo/identity"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/tasks"
)

// The feeds are written from the events of the three services whose work
// involves other people.
var (
	_ = Service.Subscribe("tasks", tasks.Events, RecordTaskEvent)
	_ = Service.Subscribe("contacts", contacts.Events, RecordContactEvent)
	_ = Service.Subscribe("groups", groups.Events, RecordGroupEvent)
)

// quoted puts a title between typographic quotes.
func quoted(title string) string { return "“" + title + "”" }

// RecordTaskEvent writes a task's event into the feed of everyone who sees
// the task, but whoever did it. The auto-archiving of a done task is left
// out: nobody did it, nobody needs to read it.
func RecordTaskEvent(ctx context.Context, e tasks.Event) error {
	if e.Kind == tasks.KindArchived && e.ActorID == "" {
		return nil
	}
	if len(e.Recipients) == 0 {
		return nil
	}
	names, err := identity.Directory(ctx, e.ActorID, e.UserID)
	if err != nil {
		return err
	}
	var group *groups.GroupRef
	if e.GroupID != "" {
		found, err := groups.BatchAPI.Call(ctx, groups.IDs{IDs: []string{e.GroupID}})
		if err != nil {
			return err
		}
		if len(found.Groups) == 1 {
			group = &found.Groups[0]
		}
	}
	actor, user, title := names[e.ActorID].Name, names[e.UserID].Name, quoted(e.Title)
	for _, r := range e.Recipients {
		var text string
		switch e.Kind {
		case tasks.KindCreated:
			text = actor + " created " + title + "."
			if group != nil {
				text = actor + " added " + title + " to " + group.Name + "."
			}
		case tasks.KindUpdated:
			text = actor + " updated " + title + "."
		case tasks.KindCompleted:
			text = actor + " completed " + title + "."
		case tasks.KindReopened:
			text = actor + " reopened " + title + "."
		case tasks.KindArchived:
			text = actor + " archived " + title + "."
		case tasks.KindRestored:
			text = actor + " restored " + title + "."
		case tasks.KindOverdue:
			text = title + " is overdue."
		case tasks.KindShared:
			text = actor + " shared " + title + " with " + you(r, e.UserID, user) + "."
		case tasks.KindUnshared:
			text = actor + " stopped sharing " + title + " with " + you(r, e.UserID, user) + "."
			if e.ActorID == e.UserID {
				text = actor + " left " + title + "."
			}
		case tasks.KindAssigned:
			text = actor + " assigned " + title + " to " + you(r, e.UserID, user) + "."
		case tasks.KindDeleted:
			text = actor + " deleted " + title + "."
		default:
			continue
		}
		task := &TaskRef{ID: e.TaskID, Title: e.Title}
		if err := write(ctx, Entry{ID: e.ID + "." + r, UserID: r, Kind: "task." + string(e.Kind),
			Actor: ref(names, e.ActorID), Task: task, Group: group, Text: text, At: e.At}); err != nil {
			return err
		}
	}
	return nil
}

// RecordContactEvent tells the addressee of a request about it, and the
// requester of an accepted one. An invitation by email has no feed to go
// to: its invitee has no account yet.
func RecordContactEvent(ctx context.Context, e contacts.Event) error {
	var text string
	switch e.Kind {
	case contacts.KindRequested:
		text = " wants to add you as a contact."
	case contacts.KindAccepted:
		text = " accepted your contact request."
	default:
		return nil
	}
	names, err := identity.Directory(ctx, e.ActorID)
	if err != nil {
		return err
	}
	return write(ctx, Entry{ID: e.ID + "." + e.UserID, UserID: e.UserID, Kind: "contact." + e.Kind,
		Actor: ref(names, e.ActorID), Text: names[e.ActorID].Name + text, At: e.At})
}

// RecordGroupEvent writes a group's comings and goings into its members'
// feeds, and an invitation into the invitee's.
func RecordGroupEvent(ctx context.Context, e groups.Event) error {
	names, err := identity.Directory(ctx, e.ActorID, e.UserID)
	if err != nil {
		return err
	}
	group := &groups.GroupRef{ID: e.GroupID, Name: e.GroupName, Color: e.Color}
	actor, user := names[e.ActorID].Name, names[e.UserID].Name
	for _, r := range e.Recipients {
		var text string
		switch e.Kind {
		case groups.KindInvited:
			text = actor + " invited you to join " + e.GroupName + "."
		case groups.KindJoined:
			text = actor + " joined " + e.GroupName + "."
		case groups.KindLeft:
			text = actor + " left " + e.GroupName + "."
		case groups.KindRemoved:
			text = actor + " removed " + you(r, e.UserID, user) + " from " + e.GroupName + "."
		case groups.KindDeleted:
			text = actor + " deleted the group " + e.GroupName + "."
		default:
			continue
		}
		if err := write(ctx, Entry{ID: e.ID + "." + r, UserID: r, Kind: "group." + e.Kind,
			Actor: ref(names, e.ActorID), Group: group, Text: text, At: e.At}); err != nil {
			return err
		}
	}
	return nil
}

// you names a user for a reader: "you" when the reader is that user.
func you(reader, user, name string) string {
	if reader == user {
		return "you"
	}
	return name
}

// ref is a user of the directory, or nil for none.
func ref(names map[string]identity.UserRef, id string) *identity.UserRef {
	if id == "" {
		return nil
	}
	u := names[id]
	return &u
}

// write adds an entry to its user's feed, once — a redelivered event leaves
// the entry, read or not, as it is — and trims the feed to Keep entries.
func write(ctx context.Context, e Entry) error {
	if err := Entries.Insert(ctx, e); err != nil {
		if wire.Is(err, kit.CodeConflict) {
			return nil
		}
		return err
	}
	feed, err := Entries.Find(ctx, "user", e.UserID)
	if err != nil || len(feed) <= Keep {
		return err
	}
	slices.SortFunc(feed, func(a, b Entry) int { return cmp.Or(a.At.Compare(b.At), cmp.Compare(a.ID, b.ID)) })
	for _, old := range feed[:len(feed)-Keep] {
		if err := Entries.Delete(ctx, old.ID); err != nil && !wire.Is(err, kit.CodeNotFound) {
			return err
		}
	}
	return nil
}
