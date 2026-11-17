// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:6a4c0cc7ef82f45cbf05f7471fb6e2644c468fdaa00e78e6aa4e3dd63779cfee

// Package outbox — The log of the reminders, and the directory they ask.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package outbox

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/notify/reminders"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 3, 2, 255), "NOT_IMPLEMENTED", "not implemented yet", "outbox: a hand-written body is still a skeleton")

// Reminded — A task its owner was reminded of.
type Reminded struct {
	TaskID string    `json:"taskId" app:"public"`
	At     time.Time `json:"at" app:"public"`
}

// Ledger is ReminderLog@v1 on the store "reminded".
type Ledger struct {
	ledger // its own state, hand-written
}

// NewLedger builds Ledger: its body is the hand-written newLedger.
func NewLedger() *Ledger {
	return newLedger()
}

// Mark is notify/port/ReminderLog@v1's.
func (a *Ledger) Mark(ctx context.Context, taskID string) error {
	return a.mark(ctx, taskID)
}

// Seen is notify/port/ReminderLog@v1's.
func (a *Ledger) Seen(ctx context.Context, taskID string) (bool, error) {
	return a.seen(ctx, taskID)
}

var _ reminders.ReminderLogV1 = (*Ledger)(nil)

// Members is Directory@v1 on the identity's query "member".
type Members struct {
	members // its own state, hand-written
}

// NewMembers builds Members: its body is the hand-written newMembers.
func NewMembers() *Members {
	return newMembers()
}

// Recipient is notify/port/Directory@v1's.
func (a *Members) Recipient(ctx context.Context, userID string) (reminders.Recipient, bool, error) {
	return a.recipient(ctx, userID)
}

var _ reminders.DirectoryV1 = (*Members)(nil)

// Service is outbox's framework service: what it declares, drawn and run.
var Service = app.NewService("outbox", "The log of the reminders, and the directory they ask.")

// RemindedStore — The tasks already reminded of.
var RemindedStore = Service.Store("reminded", func(e Reminded) string { return e.TaskID })

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts outbox calls it
// before its App runs.
func Wire() {

}
