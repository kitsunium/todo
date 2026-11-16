// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package post

import (
	"context"
	"strings"

	"github.com/kitsunium/sdk/pkg/v1/app/mail"
)

// send is the command "send"'s body: the letter is put in the outbox.
func send(ctx context.Context, in Letter) (out Posted, err error) {
	return post(ctx, in)
}

// post puts a letter in the mailer's outbox, which delivers it, tries again
// when the relay refuses it, and only captures it when the deployment names
// no relay. The subject and the name are one line each, whatever they were
// given: a line break there would be a header.
func post(ctx context.Context, l Letter) (Posted, error) {
	line := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	id, err := MailMailer.Send(ctx, mail.Message{
		To:      []mail.Address{{Name: line(l.Name), Addr: l.To}},
		Subject: line(l.Subject),
		Text:    l.Body,
	})
	return Posted{ID: id}, err
}
