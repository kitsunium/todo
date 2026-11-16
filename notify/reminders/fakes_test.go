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

// dueTask is an open task due in an hour, assigned to assignee.
func dueTask(id, assignee string) task.Task {
	return task.Task{ID: id, OwnerID: "user_ada", Title: "ship\nthe  release", Assignee: assignee, State: task.LifecycleOpen, Due: epoch.Add(time.Hour)}
}
