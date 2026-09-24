package notify

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// Every mail renders in the one design, as HTML and as plain text, with its
// one button pointing into the web app at TODO_BASE_URL — and whatever a
// user typed stays text.
func TestEveryMailRendersInTheDesign(t *testing.T) {
	t.Setenv("TODO_BASE_URL", "https://todo.example.com/")
	due := time.Date(2026, 9, 25, 17, 0, 0, 0, time.UTC)
	evil := `<script>alert("x")</script> & co`
	task := Task{ID: "task_1", Title: evil, Priority: 1, Due: &due}
	mails := map[string]Message{
		"verify":           VerifyEmail("Alice Martin", "SECRET"),
		"reset":            ResetPassword("Alice", "SECRET"),
		"exists":           AccountExists("Alice"),
		"contact request":  ContactRequest("Bob", evil),
		"contact accepted": ContactAccepted("Bob", "Alice"),
		"invite":           InviteToJoin("Alice", "dan@example.com"),
		"group":            GroupInvitation("Bob", "Alice", "group_1", evil, "violet"),
		"shared":           TaskShared("Bob", "Alice", task),
		"assigned":         TaskAssigned("Bob", "Alice", task),
		"due":              DueSoon("Bob", task, 15*time.Minute),
	}
	for name, m := range mails {
		html, text, err := render(m)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if m.Subject == "" || m.Headline == "" || m.Action.Label == "" || m.Reason == "" {
			t.Errorf("%s: incomplete %+v", name, m)
		}
		if !strings.HasPrefix(m.Action.URL, "https://todo.example.com/") {
			t.Errorf("%s: the button links to %q", name, m.Action.URL)
		}
		if strings.ContainsAny(m.Subject, "\r\n") || utf8.RuneCountInString(m.Subject) > 120 {
			t.Errorf("%s: subject %q", name, m.Subject)
		}
		for _, want := range []string{"<!doctype html>", "#f26b1d", "max-width: 560px", "prefers-color-scheme: dark", m.Action.Label} {
			if !strings.Contains(html, want) {
				t.Errorf("%s: the HTML part lacks %q", name, want)
			}
		}
		if strings.Contains(html, "<script>") {
			t.Errorf("%s: user text reached the HTML unescaped", name)
		}
		if !strings.Contains(text, m.Action.URL) || !strings.Contains(text, m.Headline) {
			t.Errorf("%s: the text part lacks the link or the headline:\n%s", name, text)
		}
		for _, line := range strings.Split(text, "\n") {
			if utf8.RuneCountInString(line) > textWidth && !strings.Contains(line, "://") {
				t.Errorf("%s: a text line is %d wide: %q", name, utf8.RuneCountInString(line), line)
			}
		}
	}
	if u := mails["verify"].Action.URL; u != "https://todo.example.com/verify?token=SECRET" {
		t.Errorf("verify links to %q", u)
	}
	if u := mails["invite"].Action.URL; u != "https://todo.example.com/signup?email=dan%40example.com" {
		t.Errorf("invite links to %q", u)
	}
	if u := mails["group"].Action.URL; u != "https://todo.example.com/app/groups/group_1" {
		t.Errorf("group links to %q", u)
	}
	if d := mails["due"]; d.Subject != "Due in 15 minutes: "+clip(evil, 60) || d.Card.Detail != "Due Fri, Sep 25 at 17:00 UTC · Urgent" {
		t.Errorf("due soon: %q, %q", d.Subject, d.Card.Detail)
	}
}

func TestClipKeepsSubjectsShortAndOnOneLine(t *testing.T) {
	if got := clip("Plan\r\nthe offsite", 60); got != "Plan  the offsite" {
		t.Errorf("clip kept a line break: %q", got)
	}
	long := strings.Repeat("word ", 30)
	if got := clip(long, 60); utf8.RuneCountInString(got) > 61 || !strings.HasSuffix(got, "…") {
		t.Errorf("clip(%d runes) = %q", utf8.RuneCountInString(long), got)
	}
}
