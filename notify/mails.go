package notify

import (
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/kitsunium/todo/internal/wire"
)

// Message is the content of one mail, before the design is applied. Every
// mail has one headline, one primary action and a footer saying why the
// recipient got it; render turns it into HTML and plain text. Every word of
// it comes from the catalogue of its language (locales/), French first.
type Message struct {
	// Lang is the language the mail is written in.
	Lang wire.Locale
	// Subject is the subject line.
	Subject string
	// Preheader is the preview line a mail client shows after the subject.
	Preheader string
	// Headline is the title of the mail.
	Headline string
	// Greeting opens the body: "Bonjour Alice,". Empty for none.
	Greeting string
	// Body is the paragraphs before the action.
	Body []string
	// Card highlights what the mail is about: a task, a group.
	Card *Card
	// Action is the one button of the mail.
	Action Action
	// Notes is the small print under the button.
	Notes []string
	// Reason says, in the footer, why the recipient got this mail.
	Reason string

	// err is the first message the catalogue could not render: render
	// refuses the mail rather than send a key in place of a sentence.
	err error
}

// Card is a highlighted block: the task or the group a mail is about.
type Card struct {
	// Label names what it is: "Tâche", "Group".
	Label string
	// Title is its name.
	Title string
	// Detail is one line under the title: the due date and the priority.
	Detail string
	// Color is its edge's color, one of the palette's hexadecimal values.
	Color string
}

// Action is the primary button of a mail.
type Action struct {
	Label string
	URL   string
}

// Task is what a task mail shows of a task.
type Task struct {
	ID       string
	Title    string
	Priority int
	Due      *time.Time
}

// accent is the brand's fox orange.
const accent = "#f26b1d"

// priorityColors follow the web app: 1 urgent … 4 low. Their names are the
// catalogue's priority.1 … priority.4.
var priorityColors = map[int]string{1: "#e5484d", 2: "#f76b15", 3: "#ffb224", 4: "#8e8c99"}

// groupColors maps the web app's group colors to their hexadecimal values.
var groupColors = map[string]string{
	"slate": "#64748b", "red": "#ef4444", "orange": "#f97316", "amber": "#f59e0b", "green": "#22c55e",
	"teal": "#14b8a6", "blue": "#3b82f6", "indigo": "#6366f1", "violet": "#8b5cf6", "pink": "#ec4899",
}

// What a subject line keeps of a name and of a title, so that every subject
// stays under 120 characters in every language.
const (
	nameInSubject  = 30
	titleInSubject = 60
)

// VerifyEmail is the mail that confirms a new account's address.
func VerifyEmail(l wire.Locale, name, token string) Message {
	w := wordsIn(l)
	return w.done(Message{
		Subject:   w.say("verify.subject"),
		Preheader: w.say("verify.preheader"),
		Headline:  w.say("verify.headline"),
		Greeting:  w.greeting(name),
		Body:      []string{w.say("verify.body")},
		Action:    Action{Label: w.say("verify.action"), URL: link("/verify", "token", token)},
		Notes:     []string{w.say("verify.note_expiry"), w.say("verify.note_ignore")},
		Reason:    w.say("verify.reason"),
	})
}

// ResetPassword is the mail whose link chooses a new password.
func ResetPassword(l wire.Locale, name, token string) Message {
	w := wordsIn(l)
	return w.done(Message{
		Subject:   w.say("reset.subject"),
		Preheader: w.say("reset.preheader"),
		Headline:  w.say("reset.headline"),
		Greeting:  w.greeting(name),
		Body:      []string{w.say("reset.body")},
		Action:    Action{Label: w.say("reset.action"), URL: link("/reset", "token", token)},
		Notes:     []string{w.say("reset.note_expiry"), w.say("reset.note_ignore")},
		Reason:    w.say("reset.reason"),
	})
}

// AccountExists answers a sign-up with an address that already has an
// account, so the sign-up itself never tells a stranger who is a user.
func AccountExists(l wire.Locale, name string) Message {
	w := wordsIn(l)
	return w.done(Message{
		Subject:   w.say("exists.subject"),
		Preheader: w.say("exists.preheader"),
		Headline:  w.say("exists.headline"),
		Greeting:  w.greeting(name),
		Body:      []string{w.say("exists.body")},
		Action:    Action{Label: w.say("exists.action"), URL: link("/login")},
		Notes:     []string{w.say("exists.note_forgot", "url", link("/forgot")), w.say("exists.note_ignore")},
		Reason:    w.say("exists.reason"),
	})
}

// ContactRequest tells a user someone wants to add them as a contact.
func ContactRequest(l wire.Locale, to, actor string) Message {
	w := wordsIn(l)
	actor = w.someone(actor)
	return w.done(Message{
		Subject:   w.say("contact_request.subject", "actor", clip(actor, nameInSubject)),
		Preheader: w.say("contact_request.preheader"),
		Headline:  w.say("contact_request.headline", "actor", actor),
		Greeting:  w.greeting(to),
		Body:      []string{w.say("contact_request.body")},
		Action:    Action{Label: w.say("contact_request.action"), URL: link("/app/contacts")},
		Reason:    w.say("contact_request.reason", "actor", actor),
	})
}

// ContactAccepted tells a user their contact request was accepted.
func ContactAccepted(l wire.Locale, to, actor string) Message {
	w := wordsIn(l)
	actor = w.someone(actor)
	return w.done(Message{
		Subject:   w.say("contact_accepted.subject", "actor", clip(actor, nameInSubject)),
		Preheader: w.say("contact_accepted.preheader", "actor", actor),
		Headline:  w.say("contact_accepted.headline", "actor", actor),
		Greeting:  w.greeting(to),
		Body:      []string{w.say("contact_accepted.body", "actor", actor)},
		Action:    Action{Label: w.say("contact_accepted.action"), URL: link("/app/contacts")},
		Reason:    w.say("contact_accepted.reason", "actor", actor),
	})
}

// InviteToJoin invites someone who has no account yet — and so no language:
// it is written in the inviter's.
func InviteToJoin(l wire.Locale, actor, email string) Message {
	w := wordsIn(l)
	actor = w.someone(actor)
	return w.done(Message{
		Subject:   w.say("invite.subject", "actor", clip(actor, nameInSubject)),
		Preheader: w.say("invite.preheader"),
		Headline:  w.say("invite.headline", "actor", actor),
		Greeting:  w.greeting(""),
		Body:      []string{w.say("invite.body", "actor", actor)},
		Action:    Action{Label: w.say("invite.action"), URL: link("/signup", "email", email)},
		Notes:     []string{w.say("invite.note")},
		Reason:    w.say("invite.reason", "actor", actor, "email", email),
	})
}

// GroupInvitation invites a contact into a group.
func GroupInvitation(l wire.Locale, to, actor, groupID, group, color string) Message {
	w := wordsIn(l)
	actor = w.someone(actor)
	hex, ok := groupColors[color]
	if !ok {
		hex = accent
	}
	return w.done(Message{
		Subject:   w.say("group_invitation.subject", "actor", clip(actor, nameInSubject), "group", clip(group, titleInSubject)),
		Preheader: w.say("group_invitation.preheader"),
		Headline:  w.say("group_invitation.headline", "group", group),
		Greeting:  w.greeting(to),
		Body:      []string{w.say("group_invitation.body", "actor", actor)},
		Card:      &Card{Label: w.say("card.group"), Title: group, Color: hex},
		Action:    Action{Label: w.say("group_invitation.action"), URL: link("/app/groups/" + groupID)},
		Reason:    w.say("group_invitation.reason", "actor", actor),
	})
}

// TaskShared tells a user a task was shared with them; its due date is
// written in zone, theirs.
func TaskShared(l wire.Locale, zone *time.Location, to, actor string, t Task) Message {
	w := wordsIn(l).at(zone)
	actor = w.someone(actor)
	return w.done(Message{
		Subject:   w.say("task_shared.subject", "actor", clip(actor, nameInSubject), "title", clip(t.Title, titleInSubject)),
		Preheader: w.say("task_shared.preheader"),
		Headline:  w.say("task_shared.headline", "actor", actor),
		Greeting:  w.greeting(to),
		Body:      []string{w.say("task_shared.body")},
		Card:      w.taskCard(t),
		Action:    Action{Label: w.say("mail.open_task"), URL: link("/app/tasks/" + t.ID)},
		Reason:    w.say("task_shared.reason", "actor", actor),
	})
}

// TaskAssigned tells a user a task was assigned to them; its due date is
// written in zone, theirs.
func TaskAssigned(l wire.Locale, zone *time.Location, to, actor string, t Task) Message {
	w := wordsIn(l).at(zone)
	actor = w.someone(actor)
	return w.done(Message{
		Subject:   w.say("task_assigned.subject", "actor", clip(actor, nameInSubject), "title", clip(t.Title, titleInSubject)),
		Preheader: w.say("task_assigned.preheader"),
		Headline:  w.say("task_assigned.headline", "actor", actor),
		Greeting:  w.greeting(to),
		Body:      []string{w.say("task_assigned.body")},
		Card:      w.taskCard(t),
		Action:    Action{Label: w.say("mail.open_task"), URL: link("/app/tasks/" + t.ID)},
		Reason:    w.say("task_assigned.reason", "actor", actor),
	})
}

// DueSoon reminds everyone a task concerns that it is due in a moment; its
// due date is written in zone, the reader's.
func DueSoon(l wire.Locale, zone *time.Location, to string, t Task, left time.Duration) Message {
	w := wordsIn(l).at(zone)
	when := w.dueIn(left)
	return w.done(Message{
		Subject:   w.say("due_soon.subject", "when", when, "title", clip(t.Title, titleInSubject)),
		Preheader: w.say("due_soon.preheader"),
		Headline:  when,
		Greeting:  w.greeting(to),
		Body:      []string{w.say("due_soon.body")},
		Card:      w.taskCard(t),
		Action:    Action{Label: w.say("mail.open_task"), URL: link("/app/tasks/" + t.ID)},
		Reason:    w.say("due_soon.reason"),
	})
}

// dueIn says how soon a task is due, to the minute: "Échéance dans
// 15 minutes", "Due in 1 minute" — with the plural rules of the language.
func (w *words) dueIn(d time.Duration) string {
	n := int(d.Round(time.Minute) / time.Minute)
	if n <= 0 {
		return w.say("due_soon.when_now")
	}
	return w.count("due_soon.when", n)
}

// taskCard shows a task: its title, its due date and its priority, on the
// priority's color.
func (w *words) taskCard(t Task) *Card {
	var detail []string
	if t.Due != nil {
		detail = append(detail, w.say("card.due", "date", w.date(*t.Due)))
	}
	color, ok := priorityColors[t.Priority]
	if ok {
		detail = append(detail, w.say("priority."+strconv.Itoa(t.Priority)))
	} else {
		color = accent
	}
	return &Card{Label: w.say("card.task"), Title: t.Title, Detail: strings.Join(detail, " · "), Color: color}
}

// clip shortens text for a subject line, on a word boundary when it can,
// and never lets a control character through.
func clip(s string, max int) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, s)
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	cut := string([]rune(s)[:max])
	if i := strings.LastIndexByte(cut, ' '); i > max/2 {
		cut = cut[:i]
	}
	return strings.TrimRight(cut, " ,.;:") + "…"
}
