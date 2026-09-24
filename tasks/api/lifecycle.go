// Package api is the task list's lifecycle and its HTTP API, declared on the
// tasks service. It lives beside package tasks rather than in it because it
// asks groups, contacts and activity — and groups and activity ask tasks
// questions back: tasks holds what they import, this package what imports
// them.
package api

import (
	"context"
	"slices"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/groups"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/tasks"
)

// ArchiveAfter is how long a done task stays on the lists before its own
// timer archives it: TODO_ARCHIVE_AFTER, a day by default.
var ArchiveAfter = kit.EnvDuration("TODO_ARCHIVE_AFTER", 24*time.Hour)

// Lifecycle is the life of every task. People complete, reopen, archive and
// restore tasks, and move an overdue one to a later date; the workflow's own
// loop marks a task overdue the moment its due date passes, and archives a
// done task once it has been done for ArchiveAfter.
var Lifecycle = tasks.Service.Workflow("lifecycle", tasks.Tasks, func(t *tasks.Task) *tasks.Status { return &t.Status }).
	Initial(tasks.Open).
	On("complete", tasks.Open, tasks.Done).
	On("complete", tasks.Overdue, tasks.Done).
	On("reopen", tasks.Done, tasks.Open).
	On("archive", tasks.Done, tasks.Archived).
	On("restore", tasks.Archived, tasks.Open).
	On("reschedule", tasks.Overdue, tasks.Open).
	When("overdue", tasks.Open, tasks.Overdue, pastDue).
	After("auto-archive", ArchiveAfter, tasks.Done, tasks.Archived).
	OnEnter(tasks.Done, stampCompletion).
	OnEnter(tasks.Open, clearCompletion).
	OnTransition(announceTransition)

// pastDue holds once a task's due date has passed.
func pastDue(t tasks.Task, now time.Time) bool { return t.Due != nil && now.After(*t.Due) }

// stampCompletion records when a task was done, and by whom: the user whose
// request completed it.
func stampCompletion(ctx context.Context, t *tasks.Task) error {
	now := wire.Now(ctx)
	t.CompletedAt, t.UpdatedAt = &now, now
	if uid, ok := kit.UserID(ctx); ok {
		t.CompletedBy = string(uid)
	}
	return nil
}

// clearCompletion forgets the completion of a task that is to do again.
func clearCompletion(_ context.Context, t *tasks.Task) error {
	t.CompletedAt, t.CompletedBy = nil, ""
	return nil
}

// transitionKinds names the events the transitions announce. A reschedule
// announces nothing of its own: the edit that caused it does.
var transitionKinds = map[string]tasks.Kind{
	"create":       tasks.KindCreated,
	"complete":     tasks.KindCompleted,
	"reopen":       tasks.KindReopened,
	"archive":      tasks.KindArchived,
	"auto-archive": tasks.KindArchived,
	"restore":      tasks.KindRestored,
	"overdue":      tasks.KindOverdue,
}

// announceTransition publishes every transition, whoever fired it — a user,
// the overdue guard, the auto-archive timer — to everyone who sees the task.
func announceTransition(ctx context.Context, c kit.Change[tasks.Task, tasks.Status]) error {
	kind, ok := transitionKinds[c.Event]
	if !ok {
		return nil
	}
	actor, _ := kit.UserID(ctx)
	e := newEvent(ctx, c.Entity, kind, string(actor))
	if c.Event != "create" {
		e.From, e.To = c.From, c.To
	}
	return publish(ctx, c.Entity, e)
}

// newEvent describes what just happened to t.
func newEvent(ctx context.Context, t tasks.Task, kind tasks.Kind, actor string) tasks.Event {
	return tasks.Event{
		ID: kit.NewID("event"), TaskID: t.ID, Title: t.Title, Kind: kind, ActorID: actor, GroupID: t.GroupID,
		At: wire.Now(ctx), Status: t.Status, Due: t.Due, Priority: t.Priority,
	}
}

// publish announces e to everyone who sees t, the actor excepted.
func publish(ctx context.Context, t tasks.Task, e tasks.Event) error {
	rs, err := recipients(ctx, t, e.ActorID)
	if err != nil {
		return err
	}
	e.Recipients = rs
	return tasks.Events.Publish(ctx, e)
}

// recipients is everyone who sees t — its audience and the members of its
// group — but the actor.
func recipients(ctx context.Context, t tasks.Task, actor string) ([]string, error) {
	users := t.Audience()
	if t.GroupID != "" {
		members, err := groups.MembersAPI.Call(ctx, groups.GroupQuery{Group: t.GroupID})
		if err != nil {
			return nil, err
		}
		for _, m := range members.Members {
			users = append(users, m.UserID)
		}
	}
	out := []string{}
	for _, u := range users {
		if u != "" && u != actor && !slices.Contains(out, u) {
			out = append(out, u)
		}
	}
	return out, nil
}
