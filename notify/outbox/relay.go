// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package outbox

import (
	"context"
	"net/smtp"
	"strings"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/notify/reminders"
)

// relay is Relay's own state: write it freely.
type relay struct {
	smtpAddrSetting *app.Setting[string]
	mailFromSetting *app.Setting[string]
}

// newRelay is NewRelay's body.
func newRelay(smtpAddrSetting *app.Setting[string], mailFromSetting *app.Setting[string]) *Relay {
	return &Relay{relay{smtpAddrSetting: smtpAddrSetting, mailFromSetting: mailFromSetting}}
}

// send is Relay.Send's body: the mail is handed to the relay smtp-addr
// names, then kept in the outbox. With no relay named it is kept only,
// Relayed false. A relay that refuses it is the error, and nothing is
// kept: the caller sends it again.
func (a *Relay) send(ctx context.Context, m reminders.Mail) (err error) {
	from, addr := a.mailFromSetting.Get(), a.smtpAddrSetting.Get()
	if addr != "" {
		if err := smtp.SendMail(addr, nil, from, []string{m.To}, message(from, m)); err != nil {
			return err
		}
	}
	return OutboxStore.Put(ctx, Sent{
		ID: app.NewID("mail"), To: m.To, Subject: m.Subject, Body: m.Body,
		At: app.Now(ctx).UTC().Truncate(time.Millisecond), Relayed: addr != "",
	})
}

// message is the mail as SMTP carries it: its headers, each one line
// whatever the mail holds, then its plain-text body.
func message(from string, m reminders.Mail) []byte {
	line := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	var b strings.Builder
	b.WriteString("From: " + line(from) + "\r\n")
	b.WriteString("To: " + line(m.To) + "\r\n")
	b.WriteString("Subject: " + line(m.Subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
	b.WriteString(strings.ReplaceAll(strings.ReplaceAll(m.Body, "\r\n", "\n"), "\n", "\r\n"))
	return []byte(b.String())
}
