// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:fff3fa1ce1d5d0ffb043be2f4bd6a1b7fa06f149490a7503c5b7744197cb6684

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

const (
	// KindDueSoon is the reminder of a task due soon; its params are title and
	// due, the due date in RFC 3339.
	KindDueSoon string = "due-soon"
	// KindExists is the mail that tells an address it already has an account.
	KindExists string = "exists"
	// KindInvite is the mail that invites an address nobody has; its params are
	// actor, the inviter's name.
	KindInvite string = "invite"
	// KindReset is the mail whose link chooses a new password: its path carries
	// the link's secret.
	KindReset string = "reset"
	// KindVerify is the mail whose link proves an address: its path carries the
	// link's secret.
	KindVerify string = "verify"
)

// Letter — A mail to one member, as a kind and its parameters; the mailer
// writes it in the member's language, as HTML and as plain text.
type Letter struct {
	To string `json:"to" app:"personal"`
	// Name is the recipient's, beside the address; it may be empty.
	Name string `json:"name" app:"personal"`
	// Locale is the recipient's language, fr or en; anything else reads French.
	Locale string `json:"locale" app:"public"`
	// TimeZone is the IANA zone the recipient's dates are written in; empty or
	// unknown reads UTC.
	TimeZone string `json:"timeZone" app:"personal"`
	// Kind is which mail it is, one of the Kind constants.
	Kind string `json:"kind" app:"public"`
	// Params are the values the kind's words and card are made of.
	Params map[string]string `json:"params,omitempty" app:"personal"`
	// Path is the action's path in the web app, with its query when it carries a
	// secret; the link is base-url and this path. Empty for no action.
	Path string `json:"path,omitempty" app:"secret"`
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

// BaseUrlSetting — Where members reach the product: every link of a mail is
// this address and a path.
var BaseUrlSetting = Service.Setting[string]("base-url", "http://localhost:4000")

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
