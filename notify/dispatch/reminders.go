package dispatch

import (
	"context"
	"errors"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/sdk/pkg/v1/logger"
	"github.com/kitsunium/todo/identity"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/notify"
	"github.com/kitsunium/todo/tasks"
)

// Remind is how long before its due date a task's reminder goes out.
const Remind = 15 * time.Minute

// keepSent is how long a sent reminder is kept after its due date.
const keepSent = 24 * time.Hour

// Reminder is a task's next reminder: when it goes out, and whether it did.
type Reminder struct {
	ID     string     `json:"id"`
	TaskID string     `json:"taskId"`
	Title  string     `json:"title"`
	Due    time.Time  `json:"due"`
	At     time.Time  `json:"at"`
	Sent   bool       `json:"sent"`
	SentAt *time.Time `json:"sentAt,omitempty"`
}

// Reminders keeps one reminder per task with a due date, found by task.
var Reminders = notify.Service.Store("reminders", func(r Reminder) string { return r.ID },
	kit.Unique("task", func(r Reminder) string { return r.TaskID }))

// The due dates reach the reminders through the task events.
var _ = notify.Service.Subscribe("track-due", tasks.Events, TrackDue)

// ReminderLoop mails a task's owner, assignee and the users it is shared
// with fifteen minutes before it is due, each in their language. kit runs it: it sleeps until the
// next reminder is due, wakes early when a task changes, and every minute
// regardless — a task event may land after the wake it caused.
var ReminderLoop = notify.Service.Loop("reminders", SendReminders,
	kit.WakeAt(NextReminder), kit.WakeOn(tasks.Events), kit.WakeEvery(time.Minute))

// TrackDue schedules the reminder of a task that has a due date ahead and is
// still to do, and moves it when the due date moves. It never deletes one: a
// reminder checks its task before it goes out, and drops itself if the task
// is done, gone or undated.
func TrackDue(ctx context.Context, e tasks.Event) error {
	if e.Due == nil || e.Kind == tasks.KindDeleted || (e.Status != tasks.Open && e.Status != tasks.Overdue) {
		return nil
	}
	due := e.Due.UTC()
	if !due.After(wire.Now(ctx)) {
		return nil
	}
	found, err := Reminders.Find(ctx, "task", e.TaskID)
	if err != nil {
		return err
	}
	if len(found) == 0 {
		err := Reminders.Insert(ctx, Reminder{ID: kit.NewID("reminder"), TaskID: e.TaskID, Title: e.Title, Due: due, At: due.Add(-Remind)})
		if wire.Is(err, kit.CodeConflict) {
			return nil // a redelivery, or a concurrent event, scheduled it
		}
		return err
	}
	if r := found[0]; r.Due.Equal(due) && r.Title == e.Title {
		return nil
	}
	_, err = Reminders.Update(ctx, found[0].ID, func(r *Reminder) error {
		if !r.Due.Equal(due) {
			r.Due, r.At, r.Sent, r.SentAt = due, due.Add(-Remind), false, nil
		}
		r.Title = e.Title
		return nil
	})
	if wire.Is(err, kit.CodeNotFound) {
		return nil
	}
	return err
}

// NextReminder is when the loop must wake next: the earliest reminder not
// sent yet.
func NextReminder(ctx context.Context) (time.Time, bool) {
	pending, err := Reminders.Filter(ctx, func(r Reminder) bool { return !r.Sent })
	if err != nil || len(pending) == 0 {
		return time.Time{}, false
	}
	next := pending[0].At
	for _, r := range pending[1:] {
		if r.At.Before(next) {
			next = r.At
		}
	}
	return next, true
}

// SendReminders sends every reminder that is due, and forgets the ones sent
// a day ago.
func SendReminders(ctx context.Context, _ kit.Wake) error {
	now := wire.Now(ctx)
	due, err := Reminders.Filter(ctx, func(r Reminder) bool { return !r.Sent && !r.At.After(now) })
	if err != nil {
		return err
	}
	var failed []error
	for _, r := range due {
		if err := remind(ctx, r, now); err != nil {
			failed = append(failed, err)
		}
	}
	old, err := Reminders.Filter(ctx, func(r Reminder) bool { return r.Sent && now.Sub(r.Due) > keepSent })
	if err != nil {
		return errors.Join(append(failed, err)...)
	}
	for _, r := range old {
		if err := Reminders.Delete(ctx, r.ID); err != nil && !wire.Is(err, kit.CodeNotFound) {
			failed = append(failed, err)
		}
	}
	return errors.Join(failed...)
}

// remind sends one reminder to the people its task concerns now, after
// checking the task still wants it.
func remind(ctx context.Context, r Reminder, now time.Time) error {
	aud, err := tasks.AudienceAPI.Call(ctx, tasks.AudienceInput{Task: r.TaskID})
	if err != nil {
		return err
	}
	t := aud.Task
	switch {
	case t == nil || t.Due == nil || (t.Status != tasks.Open && t.Status != tasks.Overdue):
		// Gone, done, or undated: nothing to remind.
		if err := Reminders.Delete(ctx, r.ID); err != nil && !wire.Is(err, kit.CodeNotFound) {
			return err
		}
		return nil
	case !t.Due.Equal(r.Due):
		// Moved, and track-due has not caught up yet: follow the task.
		_, err := Reminders.Update(ctx, r.ID, func(r *Reminder) error {
			r.Due, r.At = t.Due.UTC(), t.Due.UTC().Add(-Remind)
			return nil
		})
		return ignoreGone(err)
	case !now.Before(*t.Due):
		// Too late for "due in 15 minutes": the overdue state says it now.
		return markSent(ctx, r, now)
	}
	people, err := identity.People(ctx, aud.Users...)
	if err != nil {
		return err
	}
	task := notify.Task{ID: t.ID, Title: t.Title, Priority: int(t.Priority), Due: t.Due}
	for _, id := range aud.Users {
		to := people[id]
		if to.Email == "" {
			continue
		}
		if _, err := notify.Deliver(ctx, notify.Recipient{Name: to.Name, Email: to.Email}, notify.DueSoon(to.Locale, to.Name, task, t.Due.Sub(now))); err != nil {
			return err
		}
	}
	logger.Info(ctx, kit.Log(ctx), "reminder sent", logger.String("task", t.ID), logger.Int("recipients", len(aud.Users)))
	return markSent(ctx, r, now)
}

// markSent records that a reminder went out, or will never need to.
func markSent(ctx context.Context, r Reminder, now time.Time) error {
	_, err := Reminders.Update(ctx, r.ID, func(r *Reminder) error {
		r.Sent, r.SentAt = true, &now
		return nil
	})
	return ignoreGone(err)
}

// ignoreGone treats a reminder deleted meanwhile as done with.
func ignoreGone(err error) error {
	if wire.Is(err, kit.CodeNotFound) {
		return nil
	}
	return err
}
