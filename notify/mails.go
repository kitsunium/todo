package notify

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

// Message is the content of one mail, before the design is applied. Every
// mail has one headline, one primary action and a footer saying why the
// recipient got it; render turns it into HTML and plain text.
type Message struct {
	// Subject is the subject line.
	Subject string
	// Preheader is the preview line a mail client shows after the subject.
	Preheader string
	// Headline is the title of the mail.
	Headline string
	// Greeting opens the body: "Hi Alice,". Empty for none.
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
}

// Card is a highlighted block: the task or the group a mail is about.
type Card struct {
	// Label names what it is: "Task", "Group".
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

// priorityNames and priorityColors follow the web app: 1 urgent … 4 low.
var (
	priorityNames  = map[int]string{1: "Urgent", 2: "High priority", 3: "Medium priority", 4: "Low priority"}
	priorityColors = map[int]string{1: "#e5484d", 2: "#f76b15", 3: "#ffb224", 4: "#8e8c99"}
)

// groupColors maps the web app's group colors to their hexadecimal values.
var groupColors = map[string]string{
	"slate": "#64748b", "red": "#ef4444", "orange": "#f97316", "amber": "#f59e0b", "green": "#22c55e",
	"teal": "#14b8a6", "blue": "#3b82f6", "indigo": "#6366f1", "violet": "#8b5cf6", "pink": "#ec4899",
}

// greeting opens a mail with the recipient's first name.
func greeting(name string) string {
	first, _, _ := strings.Cut(strings.TrimSpace(name), " ")
	if first == "" {
		return "Hi,"
	}
	return "Hi " + first + ","
}

// VerifyEmail is the mail that confirms a new account's address.
func VerifyEmail(name, token string) Message {
	return Message{
		Subject:   "Confirm your email address",
		Preheader: "One click and your Todo account is ready.",
		Headline:  "Confirm your email address",
		Greeting:  greeting(name),
		Body:      []string{"Thanks for signing up for Todo. Confirm that this address is yours, and your account is ready to use."},
		Action:    Action{Label: "Confirm email address", URL: link("/verify", "token", token)},
		Notes: []string{
			"This link works once and expires in 24 hours.",
			"If you did not create an account, ignore this email: nothing happens without the link.",
		},
		Reason: "You received this email because someone signed up for Todo with this address.",
	}
}

// ResetPassword is the mail whose link chooses a new password.
func ResetPassword(name, token string) Message {
	return Message{
		Subject:   "Reset your Todo password",
		Preheader: "Choose a new password. The link expires in one hour.",
		Headline:  "Reset your password",
		Greeting:  greeting(name),
		Body:      []string{"We received a request to reset the password of your Todo account. Choose a new one below; every device signed in to your account will be signed out."},
		Action:    Action{Label: "Choose a new password", URL: link("/reset", "token", token)},
		Notes: []string{
			"This link works once and expires in one hour.",
			"If you did not ask for this, ignore this email: your password stays as it is.",
		},
		Reason: "You received this email because a password reset was requested for this address.",
	}
}

// AccountExists answers a sign-up with an address that already has an
// account, so the sign-up itself never tells a stranger who is a user.
func AccountExists(name string) Message {
	return Message{
		Subject:   "You already have a Todo account",
		Preheader: "Someone tried to sign up with your address.",
		Headline:  "You already have an account",
		Greeting:  greeting(name),
		Body:      []string{"Someone, hopefully you, just tried to create a Todo account with this email address. You already have one, so nothing changed."},
		Action:    Action{Label: "Sign in", URL: link("/login")},
		Notes: []string{
			"Forgot your password? Reset it at " + link("/forgot") + ".",
			"If this was not you, you can safely ignore this email.",
		},
		Reason: "You received this email because a sign-up was attempted with this address.",
	}
}

// ContactRequest tells a user someone wants to add them as a contact.
func ContactRequest(to, actor string) Message {
	return Message{
		Subject:   actor + " wants to add you as a contact",
		Preheader: "Contacts share tasks and invite each other to groups.",
		Headline:  actor + " wants to add you as a contact",
		Greeting:  greeting(to),
		Body:      []string{"Contacts can share tasks with each other and invite each other to their groups. Accept the request to start working together."},
		Action:    Action{Label: "Review the request", URL: link("/app/contacts")},
		Reason:    "You received this email because " + actor + " sent you a contact request on Todo.",
	}
}

// ContactAccepted tells a user their contact request was accepted.
func ContactAccepted(to, actor string) Message {
	return Message{
		Subject:   actor + " accepted your contact request",
		Preheader: "You can now share tasks with " + actor + ".",
		Headline:  "You and " + actor + " are now contacts",
		Greeting:  greeting(to),
		Body:      []string{"You can now share tasks with " + actor + " and invite them to your groups."},
		Action:    Action{Label: "Open your contacts", URL: link("/app/contacts")},
		Reason:    "You received this email because " + actor + " accepted your contact request on Todo.",
	}
}

// InviteToJoin invites someone who has no account yet.
func InviteToJoin(actor, email string) Message {
	return Message{
		Subject:   actor + " invited you to Todo",
		Preheader: "A calm, fast task list for you and the people you work with.",
		Headline:  actor + " invited you to Todo",
		Greeting:  "Hi,",
		Body:      []string{"Todo is a calm, fast task list for you and the people you work with. Create your account with this address, and " + actor + "'s contact request will be waiting for you."},
		Action:    Action{Label: "Create your account", URL: link("/signup", "email", email)},
		Notes:     []string{"Not interested? Ignore this email: you will not hear from us unless someone invites you again."},
		Reason:    "You received this email because " + actor + " invited " + email + " to Todo.",
	}
}

// GroupInvitation invites a contact into a group.
func GroupInvitation(to, actor, groupID, group, color string) Message {
	hex, ok := groupColors[color]
	if !ok {
		hex = accent
	}
	return Message{
		Subject:   actor + " invited you to join " + clip(group, 60),
		Preheader: "Members of a group share one task list.",
		Headline:  "Join " + group + " on Todo",
		Greeting:  greeting(to),
		Body:      []string{actor + " invited you to join a group. Its members share a task list and see each other's progress."},
		Card:      &Card{Label: "Group", Title: group, Color: hex},
		Action:    Action{Label: "View the invitation", URL: link("/app/groups/" + groupID)},
		Reason:    "You received this email because " + actor + " invited you to a group on Todo.",
	}
}

// TaskShared tells a user a task was shared with them.
func TaskShared(to, actor string, t Task) Message {
	return Message{
		Subject:   actor + " shared “" + clip(t.Title, 60) + "” with you",
		Preheader: "You can now see, edit and complete it.",
		Headline:  actor + " shared a task with you",
		Greeting:  greeting(to),
		Body:      []string{"You can now see, edit and complete this task. It is in your Shared with me list."},
		Card:      taskCard(t),
		Action:    Action{Label: "Open the task", URL: link("/app/tasks/" + t.ID)},
		Reason:    "You received this email because " + actor + " shared a task with you on Todo.",
	}
}

// TaskAssigned tells a user a task was assigned to them.
func TaskAssigned(to, actor string, t Task) Message {
	return Message{
		Subject:   actor + " assigned you “" + clip(t.Title, 60) + "”",
		Preheader: "It is in your Assigned to me list.",
		Headline:  actor + " assigned you a task",
		Greeting:  greeting(to),
		Body:      []string{"This task is now yours to do. It is in your Assigned to me list."},
		Card:      taskCard(t),
		Action:    Action{Label: "Open the task", URL: link("/app/tasks/" + t.ID)},
		Reason:    "You received this email because " + actor + " assigned you a task on Todo.",
	}
}

// DueSoon reminds everyone a task concerns that it is due in a moment.
func DueSoon(to string, t Task, in time.Duration) Message {
	when := "Due in " + minutes(in)
	return Message{
		Subject:   when + ": " + clip(t.Title, 60),
		Preheader: "A task you are part of is due soon.",
		Headline:  when,
		Greeting:  greeting(to),
		Body:      []string{"A task you own, are assigned or share is due soon."},
		Card:      taskCard(t),
		Action:    Action{Label: "Open the task", URL: link("/app/tasks/" + t.ID)},
		Reason:    "You received this email because you own, are assigned or share this task on Todo. You get one reminder per due date.",
	}
}

// taskCard shows a task: its title, its due date and its priority, on the
// priority's color.
func taskCard(t Task) *Card {
	var detail []string
	if t.Due != nil {
		detail = append(detail, "Due "+t.Due.UTC().Format("Mon, Jan 2 at 15:04")+" UTC")
	}
	if p, ok := priorityNames[t.Priority]; ok {
		detail = append(detail, p)
	}
	color, ok := priorityColors[t.Priority]
	if !ok {
		color = accent
	}
	return &Card{Label: "Task", Title: t.Title, Detail: strings.Join(detail, " · "), Color: color}
}

// minutes renders a short wait: "15 minutes", "1 minute", "a moment".
func minutes(d time.Duration) string {
	n := int(d.Round(time.Minute) / time.Minute)
	switch {
	case n <= 0:
		return "a moment"
	case n == 1:
		return "1 minute"
	}
	return fmt.Sprintf("%d minutes", n)
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
