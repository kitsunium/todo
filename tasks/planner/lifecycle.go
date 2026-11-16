// What the use cases share: the one way a task moves, and the order a list
// is given in.

package planner

import (
	"cmp"
	"context"
	"errors"
	"time"

	"github.com/kitsunium/todo/tasks/task"
)

// move fires event on t at now and stamps what the move changes: the state,
// UpdatedAt, and DoneAt, set when the task is completed and cleared when it
// opens again.
func move(ctx context.Context, t *task.Task, event task.LifecycleEvent, now time.Time) error {
	to, err := task.NewLifecycle().Fire(ctx, t, t.State, event)
	switch {
	case errors.Is(err, task.ErrLifecycleNoTransition):
		return task.ErrNotAllowed
	case err != nil:
		return err
	}
	t.State, t.UpdatedAt = to, now
	switch to {
	case task.LifecycleDone:
		t.DoneAt = now
	case task.LifecycleOpen:
		t.DoneAt = time.Time{}
	}
	return nil
}

// rank orders the priorities, the most urgent first.
func rank(p task.Priority) int {
	switch p {
	case task.PriorityHigh:
		return 0
	case task.PriorityLow:
		return 2
	}
	return 1
}

// byUrgency orders two tasks: the higher priority first, then the sooner
// due date (a task with none last), then the older one.
func byUrgency(a, b task.Task) int {
	return cmp.Or(
		cmp.Compare(rank(a.Priority), rank(b.Priority)),
		byDue(a, b),
		a.CreatedAt.Compare(b.CreatedAt),
		cmp.Compare(a.ID, b.ID),
	)
}

// byDue orders two tasks by due date, one with none last.
func byDue(a, b task.Task) int {
	switch {
	case a.Due.IsZero() && b.Due.IsZero():
		return 0
	case a.Due.IsZero():
		return 1
	case b.Due.IsZero():
		return -1
	}
	return a.Due.Compare(b.Due)
}
