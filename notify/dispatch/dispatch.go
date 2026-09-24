// Package dispatch is notify's rules: who gets a mail, and when. They are
// declared on the notify service, beside the notify package rather than in
// it, because they listen to tasks, contacts and groups — which depend on
// identity, which sends its own mails through notify.
//
// A subscription is delivered at least once, so each handler does everything
// that can fail — reading the directory, rendering — before the one thing it
// cannot take back: putting the mail in the outbox, last.
package dispatch

import (
	"context"

	"github.com/kitsunium/todo/contacts"
	"github.com/kitsunium/todo/groups"
	"github.com/kitsunium/todo/identity"
	"github.com/kitsunium/todo/notify"
	"github.com/kitsunium/todo/tasks"
)

// The mails people get when someone works with them.
var (
	_ = notify.Service.Subscribe("task-mail", tasks.Events, MailTaskEvent)
	_ = notify.Service.Subscribe("contact-mail", contacts.Events, MailContactEvent)
	_ = notify.Service.Subscribe("group-mail", groups.Events, MailGroupEvent)
)

// MailTaskEvent mails a user a task was shared with, or assigned to — by
// someone else.
func MailTaskEvent(ctx context.Context, e tasks.Event) error {
	if (e.Kind != tasks.KindShared && e.Kind != tasks.KindAssigned) || e.UserID == "" || e.UserID == e.ActorID {
		return nil
	}
	names, err := identity.Directory(ctx, e.ActorID, e.UserID)
	if err != nil {
		return err
	}
	to := names[e.UserID]
	if to.Email == "" {
		return nil // the account is gone
	}
	task := notify.Task{ID: e.TaskID, Title: e.Title, Priority: int(e.Priority), Due: e.Due}
	m := notify.TaskShared(to.Name, names[e.ActorID].Name, task)
	if e.Kind == tasks.KindAssigned {
		m = notify.TaskAssigned(to.Name, names[e.ActorID].Name, task)
	}
	_, err = notify.Deliver(ctx, notify.Recipient{Name: to.Name, Email: to.Email}, m)
	return err
}

// MailContactEvent mails the addressee of a contact request, the requester
// of an accepted one, and the address an invitation to join is for.
func MailContactEvent(ctx context.Context, e contacts.Event) error {
	names, err := identity.Directory(ctx, e.ActorID, e.UserID)
	if err != nil {
		return err
	}
	actor, to := names[e.ActorID].Name, names[e.UserID]
	var m notify.Message
	switch e.Kind {
	case contacts.KindRequested:
		m = notify.ContactRequest(to.Name, actor)
	case contacts.KindAccepted:
		m = notify.ContactAccepted(to.Name, actor)
	case contacts.KindInvited:
		to = identity.UserRef{Email: e.Email}
		m = notify.InviteToJoin(actor, e.Email)
	default:
		return nil
	}
	if to.Email == "" {
		return nil
	}
	_, err = notify.Deliver(ctx, notify.Recipient{Name: to.Name, Email: to.Email}, m)
	return err
}

// MailGroupEvent mails a user invited into a group.
func MailGroupEvent(ctx context.Context, e groups.Event) error {
	if e.Kind != groups.KindInvited || e.UserID == "" {
		return nil
	}
	names, err := identity.Directory(ctx, e.ActorID, e.UserID)
	if err != nil {
		return err
	}
	to := names[e.UserID]
	if to.Email == "" {
		return nil
	}
	m := notify.GroupInvitation(to.Name, names[e.ActorID].Name, e.GroupID, e.GroupName, e.Color)
	_, err = notify.Deliver(ctx, notify.Recipient{Name: to.Name, Email: to.Email}, m)
	return err
}
