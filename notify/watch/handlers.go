// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package watch

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/observe/logger"
	"github.com/kitsunium/todo/contacts/linkstore"
	"github.com/kitsunium/todo/groups/groupstore"
	"github.com/kitsunium/todo/notify/reminders"
	"github.com/kitsunium/todo/tasks/taskstore"
)

// remind is the job "remind-due"'s body: an error is the run's failure.
func remind(ctx context.Context) error {
	res, err := wired.remindDue.Execute(ctx, reminders.RemindDueInput{Within: wired.remindWithinSetting.Get()})
	if res.Sent > 0 {
		logger.Info(ctx, app.Log(ctx), "reminders sent", logger.Int("mails", res.Sent))
	}
	return err
}

// mailTask is the subscription "tasks"'s body: a task shared with a member or
// assigned to one is mailed to that member. The event's id keys the mail.
func mailTask(ctx context.Context, msg taskstore.Event) error {
	_, err := wired.mailAction.Execute(ctx, reminders.MailActionInput{
		EventID: msg.ID, Kind: msg.Kind, ActorID: msg.ActorID, UserID: msg.UserID, Subject: msg.Title, SubjectID: msg.TaskID,
	})
	return err
}

// mailContact is the subscription "contacts"'s body: a request received is
// mailed to its addressee, one accepted to its requester.
func mailContact(ctx context.Context, msg linkstore.Event) error {
	_, err := wired.mailAction.Execute(ctx, reminders.MailActionInput{EventID: msg.ID, Kind: msg.Kind, ActorID: msg.ActorID, UserID: msg.UserID})
	return err
}

// mailGroup is the subscription "groups"'s body: an invitation received is
// mailed to the invitee, naming the group.
func mailGroup(ctx context.Context, msg groupstore.Event) error {
	_, err := wired.mailAction.Execute(ctx, reminders.MailActionInput{
		EventID: msg.ID, Kind: msg.Kind, ActorID: msg.ActorID, UserID: msg.UserID, Subject: msg.GroupName, SubjectID: msg.GroupID,
	})
	return err
}
