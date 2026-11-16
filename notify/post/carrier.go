// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package post

import (
	"context"

	"github.com/kitsunium/todo/notify/reminders"
)

// carrier is Carrier's own state: the mailer is the package's, it holds nothing.
type carrier struct {
}

// newCarrier is NewCarrier's body.
func newCarrier() *Carrier {
	return &Carrier{carrier{}}
}

// send is Carrier.Send's body: the mail goes to the outbox like every other.
func (a *Carrier) send(ctx context.Context, m reminders.Mail) (err error) {
	_, err = post(ctx, Letter{To: m.To, Subject: m.Subject, Body: m.Body})
	return err
}
