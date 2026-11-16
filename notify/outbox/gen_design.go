// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:620afbb2962f72a1a0c57e91267127ee1982802bbd70150e8fa1d8268bc9f808

// Package outbox — The mails sent, kept, handed to the SMTP relay when one
// is named, their HTTP read, and the log of the reminders.
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

// Reminded — A task its assignee was reminded of.
type Reminded struct {
	TaskID string    `json:"taskId" app:"public"`
	At     time.Time `json:"at" app:"public"`
}

// Sent — A mail the product sent.
type Sent struct {
	ID      string    `json:"id" app:"public"`
	UserID  string    `json:"userId" app:"public"`
	To      string    `json:"to" app:"personal,subject"`
	Subject string    `json:"subject" app:"public"`
	Body    string    `json:"body" app:"public"`
	At      time.Time `json:"at" app:"public"`
	// Relayed says the mail was handed to the relay; false when the deployment
	// names none.
	Relayed bool `json:"relayed" app:"public"`
}

// Mails — The mails sent on a member's behalf, newest first.
type Mails struct {
	Mails []Sent `json:"mails"`
}

// NoInput — A request that says nothing.
type NoInput struct {
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

// Address is notify/port/Directory@v1's.
func (a *Members) Address(ctx context.Context, userID string) (string, bool, error) {
	return a.address(ctx, userID)
}

var _ reminders.DirectoryV1 = (*Members)(nil)

// Relay is Mailer@v1; it keeps the mail in the store "outbox" and hands it to
// the relay smtp-addr names.
type Relay struct {
	relay // its own state, hand-written
}

// NewRelay builds Relay: its body is the hand-written newRelay.
func NewRelay(smtpAddrSetting *app.Setting[string], mailFromSetting *app.Setting[string]) *Relay {
	return newRelay(smtpAddrSetting, mailFromSetting)
}

// Send is notify/port/Mailer@v1's.
func (a *Relay) Send(ctx context.Context, m reminders.Mail) error {
	return a.send(ctx, m)
}

var _ reminders.MailerV1 = (*Relay)(nil)

// Service is outbox's framework service: what it declares, drawn and run.
var Service = app.NewService("outbox", "The mails sent, kept, handed to the SMTP relay when one is named, their HTTP read, and the log of the reminders.")

// MailFromSetting — The address the mails come from.
var MailFromSetting = Service.Setting[string]("mail-from", "todo@localhost")

// SmtpAddrSetting — The SMTP relay's host and port; empty, the mails are
// kept and not relayed.
var SmtpAddrSetting = Service.Setting[string]("smtp-addr", "")

// OutboxStore — Every mail sent, keyed by its id; a mail is deleted thirty
// days after it was sent.
var OutboxStore = Service.Store("outbox", func(e Sent) string { return e.ID },
	app.DeleteAfter[Sent](time.Duration(2592000000000000) /* 720h */, func(e Sent) (time.Time, bool) { return e.At, !e.At.IsZero() }),
)

// RemindedStore — The tasks already reminded of.
var RemindedStore = Service.Store("reminded", func(e Reminded) string { return e.TaskID })

var _ = Service.Endpoint("GET /api/outbox", Kept, app.Auth())

// Kept serves GET /api/outbox: its body is the hand-written kept.
func Kept(ctx context.Context, in NoInput) (Mails, error) {
	return kept(ctx, in)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts outbox calls it
// before its App runs.
func Wire() {

}
