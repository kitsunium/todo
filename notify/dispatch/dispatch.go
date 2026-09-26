// Package dispatch is notify's rules: who gets a mail, and when. They are
// declared on the notify service, beside the notify package rather than in
// it, because they listen to tasks, contacts and groups — which depend on
// identity, which sends its own mails through notify.
//
// A subscription is delivered at least once, so each handler does everything
// that can fail — reading the directory, rendering — before the one thing it
// cannot take back: putting the mail in the outbox, last.
//
// Every mail is written in its recipient's language, which the directory
// (identity.People, over the private UsersAPI) gives with their name; an
// invitation to someone without an account, in the inviter's.
package dispatch

import (
	"context"

	"github.com/kitsunium/todo/contacts"
	"github.com/kitsunium/todo/groups"
	"github.com/kitsunium/todo/identity"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/notify"
	"github.com/kitsunium/todo/tasks"
)

// The mails people get when someone works with them.
//
// fr: Les mails que reçoivent les gens quand quelqu’un travaille avec eux.
var (
	_ = notify.Service.Subscribe("task-mail", tasks.Events, MailTaskEvent)
	_ = notify.Service.Subscribe("contact-mail", contacts.Events, MailContactEvent)
	_ = notify.Service.Subscribe("group-mail", groups.Events, MailGroupEvent)
)

// MailTaskEvent mails a user a task was shared with, or assigned to — by
// someone else.
//
// fr: MailTaskEvent écrit à un utilisateur avec qui une tâche a été partagée,
// ou à qui elle a été assignée — par quelqu’un d’autre.
func MailTaskEvent(ctx context.Context, e tasks.Event) error {
	if (e.Kind != tasks.KindShared && e.Kind != tasks.KindAssigned) || e.UserID == "" || e.UserID == e.ActorID {
		return nil
	}
	people, err := identity.People(ctx, e.ActorID, e.UserID)
	if err != nil {
		return err
	}
	to := people[e.UserID]
	if to.Email == "" {
		return nil // the account is gone
	}
	task := notify.Task{ID: e.TaskID, Title: e.Title, Priority: int(e.Priority), Due: e.Due}
	m := notify.TaskShared(to.Locale, wire.Zone(to.TimeZone), to.Name, people[e.ActorID].Name, task)
	if e.Kind == tasks.KindAssigned {
		m = notify.TaskAssigned(to.Locale, wire.Zone(to.TimeZone), to.Name, people[e.ActorID].Name, task)
	}
	_, err = notify.Deliver(ctx, notify.Recipient{Name: to.Name, Email: to.Email}, m)
	return err
}

// MailContactEvent mails the addressee of a contact request, the requester
// of an accepted one, and the address an invitation to join is for.
//
// fr: MailContactEvent écrit au destinataire d’une demande de contact, au
// demandeur d’une demande acceptée, et à l’adresse que vise une invitation à
// rejoindre le produit.
func MailContactEvent(ctx context.Context, e contacts.Event) error {
	people, err := identity.People(ctx, e.ActorID, e.UserID)
	if err != nil {
		return err
	}
	actor, to := people[e.ActorID], people[e.UserID]
	var m notify.Message
	switch e.Kind {
	case contacts.KindRequested:
		m = notify.ContactRequest(to.Locale, to.Name, actor.Name)
	case contacts.KindAccepted:
		m = notify.ContactAccepted(to.Locale, to.Name, actor.Name)
	case contacts.KindInvited:
		// No account, no language: the inviter's.
		to = identity.Person{UserRef: identity.UserRef{Email: e.Email}, Locale: actor.Locale}
		m = notify.InviteToJoin(actor.Locale, actor.Name, e.Email)
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
//
// fr: MailGroupEvent écrit à un utilisateur invité dans un groupe.
func MailGroupEvent(ctx context.Context, e groups.Event) error {
	if e.Kind != groups.KindInvited || e.UserID == "" {
		return nil
	}
	people, err := identity.People(ctx, e.ActorID, e.UserID)
	if err != nil {
		return err
	}
	to := people[e.UserID]
	if to.Email == "" {
		return nil
	}
	m := notify.GroupInvitation(to.Locale, to.Name, people[e.ActorID].Name, e.GroupID, e.GroupName, e.Color)
	_, err = notify.Deliver(ctx, notify.Recipient{Name: to.Name, Email: to.Email}, m)
	return err
}
