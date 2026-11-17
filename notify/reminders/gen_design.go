// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:c09190927b7a702a68883efedfd3b487bc8023ea307d3b6eeda776c8d6820e96

// Package reminders — Who is reminded of what, and what that needs.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package reminders

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/tasks/planner"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 3, 1, 255), "NOT_IMPLEMENTED", "not implemented yet", "reminders: a hand-written body is still a skeleton")

const (
	// OnAssigned is the kind of the event of a task assigned to a member by
	// another: the assignee is written to.
	OnAssigned string = "assigned"
	// OnContactAccepted is the kind of the event of a contact request accepted:
	// the requester is written to.
	OnContactAccepted string = "contact-accepted"
	// OnContactRequested is the kind of the event of a contact request received:
	// the addressee is written to.
	OnContactRequested string = "contact-requested"
	// OnGroupInvited is the kind of the event of a group invitation received: the
	// invitee is written to.
	OnGroupInvited string = "group-invited"
	// OnShared is the kind of the event of a task shared with a member: that
	// member is written to.
	OnShared string = "shared"
)

// Mail — A reminder to write, a task due soon, for the mailer to say in the
// recipient's language.
type Mail struct {
	To    Recipient `json:"to"`
	Title string    `json:"title" app:"personal"`
	Due   time.Time `json:"due" app:"public"`
}

// Notice — What a member did that another is written to about, for the
// mailer to say in the recipient's language: one of the On kinds, the actor's
// name, the task or the group.
type Notice struct {
	To Recipient `json:"to"`
	// Kind is one of the On constants.
	Kind string `json:"kind" app:"public"`
	// Actor is the name of the member who did it.
	Actor string `json:"actor" app:"personal"`
	// Subject is the title of the task or the name of the group; empty for a
	// contact.
	Subject string `json:"subject" app:"personal"`
	// SubjectID is the id of the task, which the action leads to; empty otherwise.
	SubjectID string `json:"subjectId" app:"public"`
}

// Recipient — Who a reminder is written to, and how they read it.
type Recipient struct {
	Address string `json:"address" app:"personal"`
	Name    string `json:"name" app:"personal"`
	// Locale is fr or en; empty reads French.
	Locale string `json:"locale" app:"public"`
	// TimeZone is the IANA zone the due date is written in; empty reads UTC.
	TimeZone string `json:"timeZone" app:"personal"`
}

// ActionLogV1 remembers which events were mailed to which member, so each is
// once.
type ActionLogV1 interface {
	Mark(ctx context.Context, key string) error
	Seen(ctx context.Context, key string) (bool, error)
}

// DirectoryV1 gives who a member is written to, and how they read.
type DirectoryV1 interface {
	Recipient(ctx context.Context, userID string) (Recipient, bool, error)
}

// MailerV1 sends a reminder.
type MailerV1 interface {
	Send(ctx context.Context, m Mail) error
}

// NotifierV1 sends a member what another did.
type NotifierV1 interface {
	Notify(ctx context.Context, n Notice) error
}

// ReminderLogV1 remembers which tasks were reminded of, so each is once.
type ReminderLogV1 interface {
	Mark(ctx context.Context, taskID string) error
	Seen(ctx context.Context, taskID string) (bool, error)
}

// MailAction mails a member what another did to them or with them: a contact
// request received or accepted, a group invitation, a task shared or assigned;
// once for each event and member, and never about what a member did to
// themselves.
type MailAction interface {
	Execute(ctx context.Context, in MailActionInput) (MailActionOutput, error)
}

// MailActionInput is MailAction's input: a field added breaks no caller.
type MailActionInput struct {
	EventID   string
	Kind      string
	ActorID   string
	UserID    string
	Subject   string
	SubjectID string
}

// MailActionOutput is MailAction's output.
type MailActionOutput struct {
	Sent int
}

// MailActionHandler runs MailAction: its body is the hand-written execute.
type MailActionHandler struct {
	actionLog ActionLogV1
	directory DirectoryV1
	notifier  NotifierV1
}

// NewMailAction builds MailAction on what it calls.
func NewMailAction(actionLog ActionLogV1, directory DirectoryV1, notifier NotifierV1) *MailActionHandler {
	return &MailActionHandler{actionLog: actionLog, directory: directory, notifier: notifier}
}

// Execute runs MailAction.
func (h *MailActionHandler) Execute(ctx context.Context, in MailActionInput) (MailActionOutput, error) {
	return h.execute(ctx, &in)
}

var _ MailAction = (*MailActionHandler)(nil)

// RemindDue mails the owner of each open task due within a delay, once per
// task.
type RemindDue interface {
	Execute(ctx context.Context, in RemindDueInput) (RemindDueOutput, error)
}

// RemindDueInput is RemindDue's input: a field added breaks no caller.
type RemindDueInput struct {
	Within time.Duration
}

// RemindDueOutput is RemindDue's output.
type RemindDueOutput struct {
	Sent int
}

// RemindDueHandler runs RemindDue: its body is the hand-written execute.
type RemindDueHandler struct {
	dueSoon     planner.DueSoon
	reminderLog ReminderLogV1
	directory   DirectoryV1
	mailer      MailerV1
}

// NewRemindDue builds RemindDue on what it calls.
func NewRemindDue(dueSoon planner.DueSoon, reminderLog ReminderLogV1, directory DirectoryV1, mailer MailerV1) *RemindDueHandler {
	return &RemindDueHandler{dueSoon: dueSoon, reminderLog: reminderLog, directory: directory, mailer: mailer}
}

// Execute runs RemindDue.
func (h *RemindDueHandler) Execute(ctx context.Context, in RemindDueInput) (RemindDueOutput, error) {
	return h.execute(ctx, &in)
}

var _ RemindDue = (*RemindDueHandler)(nil)
