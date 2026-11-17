// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:32913ff1f1c2e4c8f3bc85573da565471d0c6a3bf2b70f14abb926ad227fee32

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

// Mail — A reminder to write, a task due soon, for the mailer to say in the
// recipient's language.
type Mail struct {
	To    Recipient `json:"to"`
	Title string    `json:"title" app:"personal"`
	Due   time.Time `json:"due" app:"public"`
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

// DirectoryV1 gives who a member is written to, and how they read.
type DirectoryV1 interface {
	Recipient(ctx context.Context, userID string) (Recipient, bool, error)
}

// MailerV1 sends a reminder.
type MailerV1 interface {
	Send(ctx context.Context, m Mail) error
}

// ReminderLogV1 remembers which tasks were reminded of, so each is once.
type ReminderLogV1 interface {
	Mark(ctx context.Context, taskID string) error
	Seen(ctx context.Context, taskID string) (bool, error)
}

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
