// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:c64b5f4ee68d0e0de734834bbe3d83de5623419c1acb369cb084a2d557d20f34

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

// Mail — A mail, as plain text.
type Mail struct {
	To      string `json:"to" app:"personal"`
	Subject string `json:"subject" app:"public"`
	Body    string `json:"body" app:"personal"`
}

// DirectoryV1 gives the address a member is written to.
type DirectoryV1 interface {
	Address(ctx context.Context, userID string) (string, bool, error)
}

// MailerV1 sends a mail.
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
