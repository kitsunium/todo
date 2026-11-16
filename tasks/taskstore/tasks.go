// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package taskstore

import (
	"context"
	"errors"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/tasks/task"
)

// tasks is Tasks's own state: the store is the package's, it holds nothing.
type tasks struct {
}

// newTasks is NewTasks's body.
func newTasks() *Tasks {
	return &Tasks{tasks{}}
}

// get is Tasks.Get's body: a key the store does not hold is ErrNotFound.
func (a *Tasks) get(ctx context.Context, id string) (r0 task.Task, err error) {
	rec, err := TasksStore.Get(ctx, id)
	var refused *app.Error
	switch {
	case errors.As(err, &refused) && refused.Code == app.WireNotFound:
		return task.Task{}, task.ErrNotFound
	case err != nil:
		return task.Task{}, err
	}
	return fromRecord(rec), nil
}

// list is Tasks.List's body.
func (a *Tasks) list(ctx context.Context) (r0 []task.Task, err error) {
	recs, err := TasksStore.List(ctx)
	if err != nil {
		return nil, err
	}
	r0 = make([]task.Task, len(recs))
	for i, rec := range recs {
		r0[i] = fromRecord(rec)
	}
	return r0, nil
}

// newID is Tasks.NewID's body: a time-ordered id.
func (a *Tasks) newID(ctx context.Context) (r0 string, err error) {
	return app.NewID("task"), nil
}

// save is Tasks.Save's body: the task replaces what its id held.
func (a *Tasks) save(ctx context.Context, t task.Task) (err error) {
	return TasksStore.Put(ctx, toRecord(t))
}

// toRecord is a task as the store keeps it.
func toRecord(t task.Task) Record {
	return Record{
		ID: t.ID, Title: t.Title, Notes: t.Notes, Priority: string(t.Priority), Assignee: t.Assignee,
		Due: t.Due, State: string(t.State), CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt, DoneAt: t.DoneAt,
	}
}

// fromRecord is the task a record keeps.
func fromRecord(r Record) task.Task {
	return task.Task{
		ID: r.ID, Title: r.Title, Notes: r.Notes, Priority: task.Priority(r.Priority), Assignee: r.Assignee,
		Due: r.Due, State: task.LifecycleState(r.State), CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt, DoneAt: r.DoneAt,
	}
}
