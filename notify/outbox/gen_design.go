// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:c09190927b7a702a68883efedfd3b487bc8023ea307d3b6eeda776c8d6820e96

// Package outbox — The logs of the reminders and of the mails about what
// others did, and the directory they ask.
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

// Acted — An event mailed to a member, by its key.
type Acted struct {
	// Key is the event's id and the member it was mailed to.
	Key string    `json:"key" app:"public"`
	At  time.Time `json:"at" app:"public"`
}

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

// Register is ActionLog@v1 on the store "acted".
type Register struct {
	register // its own state, hand-written
}

// NewRegister builds Register: its body is the hand-written newRegister.
func NewRegister() *Register {
	return newRegister()
}

// Mark is notify/port/ActionLog@v1's.
func (a *Register) Mark(ctx context.Context, key string) error {
	return a.mark(ctx, key)
}

// Seen is notify/port/ActionLog@v1's.
func (a *Register) Seen(ctx context.Context, key string) (bool, error) {
	return a.seen(ctx, key)
}

var _ reminders.ActionLogV1 = (*Register)(nil)

// Service is outbox's framework service: what it declares, drawn and run.
var Service = app.NewService("outbox", "The logs of the reminders and of the mails about what others did, and the directory they ask.")

// RemindedStore — The tasks already reminded of.
var RemindedStore = Service.Store("reminded", func(e Reminded) string { return e.TaskID })

// ActedStore — The events already mailed, each to its member; a key is
// deleted ninety days after it was written.
var ActedStore = Service.Store("acted", func(e Acted) string { return e.Key },
	app.DeleteAfter[Acted](time.Duration(7776000000000000) /* 2160h */, func(e Acted) (time.Time, bool) { return e.At, !e.At.IsZero() }),
)

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts outbox calls it
// before its App runs.
func Wire() {

}
