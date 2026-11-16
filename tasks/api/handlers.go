// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package api

import (
	"context"
	"errors"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/observe/logger"
	"github.com/kitsunium/todo/tasks/planner"
	"github.com/kitsunium/todo/tasks/task"
)

// add is POST /api/tasks's body; what it uses is in wired.
func add(ctx context.Context, in AddRequest) (out task.Task, err error) {
	owner, err := caller(ctx)
	if err != nil {
		return task.Task{}, err
	}
	res, err := wired.addTask.Execute(ctx, planner.AddTaskInput{
		Owner: owner, Title: in.Title, Notes: in.Notes, Priority: in.Priority, Assignee: in.Assignee, Due: in.Due,
	})
	if err != nil {
		return task.Task{}, refusal(err)
	}
	logger.Info(ctx, app.Log(ctx), "task added", logger.String("task", res.Task.ID), logger.String("user", owner), logger.String("priority", string(res.Task.Priority)))
	return res.Task, nil
}

// list is GET /api/tasks's body; what it uses is in wired.
func list(ctx context.Context, in ListRequest) (out TaskList, err error) {
	owner, err := caller(ctx)
	if err != nil {
		return TaskList{}, err
	}
	res, err := wired.listTasks.Execute(ctx, planner.ListTasksInput{Owner: owner, State: in.State})
	if err != nil {
		return TaskList{}, refusal(err)
	}
	return TaskList{Tasks: res.Tasks}, nil
}

// complete is POST /api/tasks/{id}/complete's body; what it uses is in wired.
func complete(ctx context.Context, in ByID) (out task.Task, err error) {
	return move(ctx, in.ID, task.LifecycleOnComplete)
}

// reopen is POST /api/tasks/{id}/reopen's body; what it uses is in wired.
func reopen(ctx context.Context, in ByID) (out task.Task, err error) {
	return move(ctx, in.ID, task.LifecycleOnReopen)
}

// archive is POST /api/tasks/{id}/archive's body; what it uses is in wired.
func archive(ctx context.Context, in ByID) (out task.Task, err error) {
	return move(ctx, in.ID, task.LifecycleOnArchive)
}

// restore is POST /api/tasks/{id}/restore's body; what it uses is in wired.
func restore(ctx context.Context, in ByID) (out task.Task, err error) {
	return move(ctx, in.ID, task.LifecycleOnRestore)
}

// sweep is the job "archive-done"'s body: an error is the run's failure.
func sweep(ctx context.Context) error {
	res, err := wired.archiveDone.Execute(ctx, planner.ArchiveDoneInput{After: wired.archiveAfterSetting.Get()})
	if err != nil {
		return err
	}
	if res.Archived > 0 {
		logger.Info(ctx, app.Log(ctx), "done tasks archived", logger.Int("tasks", res.Archived))
	}
	return nil
}

// count is the query "census"'s body: every member's tasks counted, through
// their use case.
func count(ctx context.Context, in NoInput) (out Census, err error) {
	res, err := wired.countTasks.Execute(ctx, planner.CountTasksInput{})
	if err != nil {
		return Census{}, err
	}
	return Census{Counts: res.Counts, Total: res.Total, Overdue: res.Overdue}, nil
}

// move runs MoveTask for one of the four routes that move a task.
func move(ctx context.Context, id string, event task.LifecycleEvent) (task.Task, error) {
	owner, err := caller(ctx)
	if err != nil {
		return task.Task{}, err
	}
	res, err := wired.moveTask.Execute(ctx, planner.MoveTaskInput{Owner: owner, ID: id, Event: string(event)})
	if err != nil {
		return task.Task{}, refusal(err)
	}
	return res.Task, nil
}

// caller is the member the request comes from. Every route that calls it
// asks for one, so a request without is refused before it gets here: this
// covers a call made in process.
func caller(ctx context.Context) (string, error) {
	uid, ok := app.UserID(ctx)
	if !ok {
		return "", app.Unauthenticated("sign in first")
	}
	return string(uid), nil
}

// refusal is a use case's error as the API answers it: 404 for a task that
// does not exist, 409 for a move its state does not allow, 400 with the
// field at fault for a task that cannot be added; anything else is the
// framework's to answer.
func refusal(err error) error {
	invalid := func(path, rule string, cause error) error {
		e := app.Invalid("the request is invalid")
		e.Violations = []app.Violation{{Path: path, Rule: rule, Message: cause.Error()}}
		return e
	}
	switch {
	case errors.Is(err, task.ErrNotFound):
		return app.NotFound("no such task")
	case errors.Is(err, task.ErrNotAllowed):
		return app.Conflict("the task cannot make this move from the state it is in")
	case errors.Is(err, task.ErrEmptyTitle):
		return invalid("title", "required", err)
	case errors.Is(err, task.ErrTitleTooLong):
		return invalid("title", "maxlen", err)
	case errors.Is(err, task.ErrUnknownPriority):
		return invalid("priority", "oneof", err)
	}
	return err
}
