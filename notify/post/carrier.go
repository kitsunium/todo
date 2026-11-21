// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package post

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/kitsunium/todo/notify/reminders"
)

// carrier is Carrier's own state: the mailer is the package's, it holds nothing.
type carrier struct {
}

// newCarrier is NewCarrier's body.
func newCarrier() *Carrier {
	return &Carrier{carrier{}}
}

// send is Carrier.Send's body: the reminder goes to the outbox like every
// other mail, in its recipient's language, its action the web app's root.
func (a *Carrier) send(ctx context.Context, m reminders.Mail) (err error) {
	_, err = post(ctx, Letter{
		To: m.To.Address, Name: m.To.Name, Locale: m.To.Locale, TimeZone: m.To.TimeZone, Kind: KindDueSoon, Path: "/",
		Params: map[string]string{"title": m.Title, "due": m.Due.UTC().Format(time.RFC3339)},
	})
	return err
}

// notify is Carrier.Notify's body: what a member did is written to the member
// it concerns, in their language and their zone, through the same outbox. A
// contact request leads to the contacts' page, an invitation to the groups',
// a task to itself.
func (a *Carrier) notify(ctx context.Context, n reminders.Notice) (err error) {
	l := Letter{To: n.To.Address, Name: n.To.Name, Locale: n.To.Locale, TimeZone: n.To.TimeZone, Params: map[string]string{"actor": n.Actor}}
	switch n.Kind {
	case reminders.OnContactRequested:
		l.Kind, l.Path = KindContactRequest, "/contacts"
	case reminders.OnContactAccepted:
		l.Kind, l.Path = KindContactAccepted, "/contacts"
	case reminders.OnGroupInvited:
		l.Kind, l.Path, l.Params["group"] = KindGroupInvitation, "/groups", n.Subject
	case reminders.OnShared:
		l.Kind, l.Path, l.Params["title"] = KindTaskShared, "/task/"+url.PathEscape(n.SubjectID), n.Subject
	case reminders.OnAssigned:
		l.Kind, l.Path, l.Params["title"] = KindTaskAssigned, "/task/"+url.PathEscape(n.SubjectID), n.Subject
	default:
		return fmt.Errorf("post: no mail for what a member did, kind %q", n.Kind)
	}
	_, err = post(ctx, l)
	return err
}
