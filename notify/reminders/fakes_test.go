// The ports RemindDue's tests run on.

package reminders

import (
	"context"
	"time"

	"github.com/kitsunium/todo/tasks/planner"
	"github.com/kitsunium/todo/tasks/task"
)

// epoch is the instant every test starts at.
var epoch = time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)

// dueSoon is planner.DueSoon answering the tasks it holds.
type dueSoon []task.Task

func (d dueSoon) Execute(context.Context, planner.DueSoonInput) (planner.DueSoonOutput, error) {
	return planner.DueSoonOutput{Tasks: d}, nil
}

// memLog is ReminderLogV1 in memory.
type memLog map[string]bool

func (l memLog) Seen(_ context.Context, id string) (bool, error) { return l[id], nil }

func (l memLog) Mark(_ context.Context, id string) error {
	l[id] = true
	return nil
}

// memMailer is MailerV1 that keeps what it was given.
type memMailer struct {
	sent []Mail
}

func (m *memMailer) Send(_ context.Context, mail Mail) error {
	m.sent = append(m.sent, mail)
	return nil
}

// directory is DirectoryV1 over the recipients it holds, by member.
type directory map[string]Recipient

func (d directory) Recipient(_ context.Context, userID string) (Recipient, bool, error) {
	to, ok := d[userID]
	return to, ok, nil
}

// everyone is the directory of the tests: Ada, who reads English in Paris,
// Bob, who reads French, and nobody else.
var everyone = directory{
	"user_ada": {Address: "ada@example.com", Name: "Ada", Locale: "en", TimeZone: "Europe/Paris"},
	"user_bob": {Address: "bob@example.com", Name: "Bob", Locale: "fr", TimeZone: "Europe/Paris"},
}

// memNotifier is NotifierV1 that keeps what it was given.
type memNotifier struct {
	sent []Notice
}

func (m *memNotifier) Notify(_ context.Context, n Notice) error {
	m.sent = append(m.sent, n)
	return nil
}

// dueTask is an open task of owner due in an hour.
func dueTask(id, owner string) task.Task {
	return task.Task{ID: id, OwnerID: owner, Title: "ship\nthe  release", State: task.LifecycleOpen, Due: epoch.Add(time.Hour)}
}
