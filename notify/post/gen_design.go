// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:2a2e1ae2f3c14afcf45b67e8c18923c85e5975df67c02bb2610e585bf5526ed2

// Package post — The product's mailer, its durable outbox and its transport,
// and the command every component that writes to a member sends through.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package post

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/notify/reminders"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 3, 4, 255), "NOT_IMPLEMENTED", "not implemented yet", "post: a hand-written body is still a skeleton")

// Letter — A mail to one member, as plain text.
type Letter struct {
	To string `json:"to" app:"personal"`
	// Name is the recipient's, beside the address; it may be empty.
	Name    string `json:"name" app:"personal"`
	Subject string `json:"subject" app:"public"`
	Body    string `json:"body" app:"personal"`
}

// Posted — A mail the outbox took.
type Posted struct {
	ID string `json:"id" app:"public"`
}

// Carrier is Mailer@v1 on the mailer "mail".
type Carrier struct {
	carrier // its own state, hand-written
}

// NewCarrier builds Carrier: its body is the hand-written newCarrier.
func NewCarrier() *Carrier {
	return newCarrier()
}

// Send is notify/port/Mailer@v1's.
func (a *Carrier) Send(ctx context.Context, m reminders.Mail) error {
	return a.send(ctx, m)
}

var _ reminders.MailerV1 = (*Carrier)(nil)

// Service is post's framework service: what it declares, drawn and run.
var Service = app.NewService("post", "The product's mailer, its durable outbox and its transport, and the command every component that writes to a member sends through.")

// SendCommand — Puts a mail in the outbox: the one path every mail of the
// product takes.
var SendCommand = Service.Command("send", Send)

// Send handles the command "send": its body is the hand-written send.
func Send(ctx context.Context, in Letter) (Posted, error) {
	return send(ctx, in)
}

// MailMailer is the mailer "mail": a message is kept in its outbox, sent once
// the transaction that wrote it commits, and tried again until it is given up.
var MailMailer = Service.Mailer("mail", app.From("Todo", "hello@todo.localhost"))

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts post calls it
// before its App runs.
func Wire() {

}
