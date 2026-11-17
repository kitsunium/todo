// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package post

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kitsunium/sdk/pkg/v1/app/mail"
)

// What a subject line keeps of a name and of a title, so that every subject
// stays short in every language.
const (
	nameInSubject  = 30
	titleInSubject = 60
)

// send is the command "send"'s body: the letter is written and put in the
// outbox.
func send(ctx context.Context, in Letter) (out Posted, err error) {
	return post(ctx, in)
}

// post writes a letter in the recipient's language, as HTML with a plain text
// alternative, and puts it in the mailer's outbox, which delivers it, tries
// again when the relay refuses it, and only captures it when the deployment
// names no relay. The subject and the name are one line each, whatever they
// were given: a line break there would be a header.
func post(ctx context.Context, l Letter) (Posted, error) {
	m, html, text, err := write(l, BaseUrlSetting.Get())
	if err != nil {
		return Posted{}, err
	}
	id, err := MailMailer.Send(ctx, mail.Message{
		To:      []mail.Address{{Name: line(l.Name), Addr: l.To}},
		Subject: m.Subject,
		Text:    text,
		HTML:    html,
	})
	return Posted{ID: id}, err
}

// write is a letter as the message it says and its two renderings, the
// links built on base. A letter whose kind is unknown, or whose parameters or
// path do not fit it, is refused.
func write(l Letter, base string) (m message, html, text string, err error) {
	w := wordsIn(l.Locale, l.TimeZone)
	name := line(l.Name)
	var act *action
	if l.Path != "" {
		target, err := link(base, l.Path)
		if err != nil {
			return message{}, "", "", err
		}
		act = &action{URL: target}
	}
	label := func(key string) { // an action has its label; a mail with none has no button
		if act != nil {
			act.Label = w.say(key)
		}
	}
	someone := func(actor string) string { // whoever did it, by name; a member with none is "a member"
		if actor = line(actor); actor != "" {
			return actor
		}
		return w.say("mail.someone")
	}
	switch l.Kind {
	case KindVerify:
		label("verify.action")
		m = message{
			Subject: w.say("verify.subject"), Preheader: w.say("verify.preheader"), Headline: w.say("verify.headline"),
			Body: []string{w.say("verify.body")}, Notes: []string{w.say("verify.note_ignore")}, Reason: w.say("verify.reason"),
		}
	case KindReset:
		label("reset.action")
		m = message{
			Subject: w.say("reset.subject"), Preheader: w.say("reset.preheader"), Headline: w.say("reset.headline"),
			Body: []string{w.say("reset.body")}, Notes: []string{w.say("reset.note_ignore")}, Reason: w.say("reset.reason"),
		}
	case KindExists:
		label("exists.action")
		m = message{
			Subject: w.say("exists.subject"), Preheader: w.say("exists.preheader"), Headline: w.say("exists.headline"),
			Body: []string{w.say("exists.body"), w.say("exists.body_forgot")}, Notes: []string{w.say("exists.note_ignore")}, Reason: w.say("exists.reason"),
		}
	case KindInvite:
		who := someone(l.Params["actor"])
		label("invite.action")
		m = message{
			Subject: w.say("invite.subject", "actor", clip(who, nameInSubject)), Preheader: w.say("invite.preheader"), Headline: w.say("invite.headline", "actor", who),
			Body:  []string{w.say("invite.body", "actor", who), w.say("invite.body_create")},
			Notes: []string{w.say("invite.note_ignore")}, Reason: w.say("invite.reason", "actor", who, "email", l.To),
		}
	case KindDueSoon:
		title := line(l.Params["title"])
		due, err := time.Parse(time.RFC3339, l.Params["due"])
		if err != nil {
			return message{}, "", "", fmt.Errorf("post: the due date of a reminder: %w", err)
		}
		label("due_soon.action")
		m = message{
			Subject: w.say("due_soon.subject", "title", clip(title, titleInSubject)), Preheader: w.say("due_soon.preheader"), Headline: w.say("due_soon.headline"),
			Body: []string{w.say("due_soon.body", "title", title, "date", w.date(due))}, Reason: w.say("due_soon.reason"),
		}
	case KindContactRequest, KindContactAccepted:
		key := "contact_request"
		if l.Kind == KindContactAccepted {
			key = "contact_accepted"
		}
		who := someone(l.Params["actor"])
		label(key + ".action")
		m = message{
			Subject: w.say(key+".subject", "actor", clip(who, nameInSubject)), Preheader: w.say(key+".preheader", "actor", who), Headline: w.say(key+".headline", "actor", who),
			Body: []string{w.say(key+".body", "actor", who)}, Reason: w.say(key+".reason", "actor", who),
		}
	case KindGroupInvitation:
		who, group := someone(l.Params["actor"]), line(l.Params["group"])
		label("group_invitation.action")
		m = message{
			Subject: w.say("group_invitation.subject", "actor", clip(who, nameInSubject), "group", clip(group, titleInSubject)), Preheader: w.say("group_invitation.preheader"),
			Headline: w.say("group_invitation.headline", "group", group), Body: []string{w.say("group_invitation.body", "actor", who), group},
			Reason: w.say("group_invitation.reason", "actor", who),
		}
	case KindTaskShared, KindTaskAssigned:
		key := "task_shared"
		if l.Kind == KindTaskAssigned {
			key = "task_assigned"
		}
		who, title := someone(l.Params["actor"]), line(l.Params["title"])
		label(key + ".action")
		m = message{
			Subject: w.say(key+".subject", "actor", clip(who, nameInSubject), "title", clip(title, titleInSubject)), Preheader: w.say(key + ".preheader"),
			Headline: w.say(key+".headline", "actor", who), Body: []string{w.say("mail.quoted", "title", title), w.say(key + ".body")},
			Reason: w.say(key+".reason", "actor", who),
		}
	default:
		return message{}, "", "", fmt.Errorf("post: no mail of kind %q", l.Kind)
	}
	m.Greeting, m.Action = w.greeting(name), act
	html, text, err = render(m, w, strings.TrimRight(base, "/"))
	return m, html, text, err
}

// line is a text on one line: its runs of white space, line breaks included,
// are one space.
func line(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// clip is s cut to n characters, with an ellipsis when it was longer.
func clip(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n-1]) + "…"
	}
	return s
}
