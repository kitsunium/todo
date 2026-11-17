// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:390717aeedfce4c9881bfea93cf567e5dfb6e0f4b99777661b165bb23d396f17

// Package watch — The job that sends the reminders, and the subscriptions
// that mail what others did.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package watch

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/contacts/linkstore"
	"github.com/kitsunium/todo/groups/groupstore"
	"github.com/kitsunium/todo/notify/reminders"
	"github.com/kitsunium/todo/tasks/taskstore"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 3, 3, 255), "NOT_IMPLEMENTED", "not implemented yet", "watch: a hand-written body is still a skeleton")

// Service is watch's framework service: what it declares, drawn and run.
var Service = app.NewService("watch", "The job that sends the reminders, and the subscriptions that mail what others did.")

// RemindWithinSetting — How long before its due date a task's owner is
// reminded.
var RemindWithinSetting = Service.Setting[time.Duration]("remind-within", 24*time.Hour)

var _ = Service.Every("remind-due", time.Minute, Remind)

// Remind — Reminds the owners of the tasks due within remind-within.
func Remind(ctx context.Context) error {
	return remind(ctx)
}

var _ = Service.Subscribe("tasks", taskstore.EventsTopic, MailTask)

// MailTask — A task shared with a member or assigned to one: the member is
// mailed.
func MailTask(ctx context.Context, msg taskstore.Event) error {
	return mailTask(ctx, msg)
}

var _ = Service.Subscribe("contacts", linkstore.EventsTopic, MailContact)

// MailContact — A contact request received or accepted: the addressee or the
// requester is mailed.
func MailContact(ctx context.Context, msg linkstore.Event) error {
	return mailContact(ctx, msg)
}

var _ = Service.Subscribe("groups", groupstore.EventsTopic, MailGroup)

// MailGroup — A group invitation received: the invitee is mailed.
func MailGroup(ctx context.Context, msg groupstore.Event) error {
	return mailGroup(ctx, msg)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
	remindDue           reminders.RemindDue
	remindWithinSetting *app.Setting[time.Duration]
	mailAction          reminders.MailAction
}

// Wire gives the handlers what they use: the role that mounts watch calls it
// before its App runs.
func Wire(remindDue reminders.RemindDue, remindWithinSetting *app.Setting[time.Duration], mailAction reminders.MailAction) {
	wired.remindDue = remindDue
	wired.remindWithinSetting = remindWithinSetting
	wired.mailAction = mailAction
}
