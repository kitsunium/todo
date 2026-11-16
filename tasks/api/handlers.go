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
		Owner: owner, Title: in.Title, Notes: in.Notes, Priority: in.Priority, GroupID: in.GroupID, AssigneeID: in.AssigneeID, Due: in.Due,
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
	res, err := wired.listTasks.Execute(ctx, planner.ListTasksInput{Owner: owner, State: in.State, View: in.View, Group: in.Group, TZ: in.TZ})
	if err != nil {
		return TaskList{}, refusal(err)
	}
	return TaskList{Tasks: res.Tasks}, nil
}

// counts is GET /api/tasks/counts's body; what it uses is in wired.
func counts(ctx context.Context, in CountsRequest) (out ViewCounts, err error) {
	me, err := caller(ctx)
	if err != nil {
		return ViewCounts{}, err
	}
	res, err := wired.countViews.Execute(ctx, planner.CountViewsInput{User: me, TZ: in.TZ})
	if err != nil {
		return ViewCounts{}, refusal(err)
	}
	return ViewCounts{
		Inbox: res.Inbox, Today: res.Today, Upcoming: res.Upcoming, Shared: res.Shared, Assigned: res.Assigned,
		Overdue: res.Overdue, CompletedThisWeek: res.CompletedThisWeek, Groups: res.Groups,
	}, nil
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

// get is GET /api/tasks/{id}'s body; what it uses is in wired.
func get(ctx context.Context, in ByID) (out task.Task, err error) {
	me, err := caller(ctx)
	if err != nil {
		return task.Task{}, err
	}
	res, err := wired.getTask.Execute(ctx, planner.GetTaskInput{User: me, ID: in.ID})
	if err != nil {
		return task.Task{}, refusal(err)
	}
	return res.Task, nil
}

// edit is PATCH /api/tasks/{id}'s body; what it uses is in wired.
func edit(ctx context.Context, in EditRequest) (out task.Task, err error) {
	me, err := caller(ctx)
	if err != nil {
		return task.Task{}, err
	}
	res, err := wired.editTask.Execute(ctx, planner.EditTaskInput{
		User: me, ID: in.ID, Title: in.Title, Notes: in.Notes, Priority: in.Priority, AssigneeID: in.AssigneeID, GroupID: in.GroupID, Due: in.Due,
	})
	if err != nil {
		return task.Task{}, refusal(err)
	}
	return res.Task, nil
}

// deleteBody is DELETE /api/tasks/{id}'s body; what it uses is in wired.
func deleteBody(ctx context.Context, in ByID) (out NoInput, err error) {
	me, err := caller(ctx)
	if err != nil {
		return NoInput{}, err
	}
	if _, err := wired.deleteTask.Execute(ctx, planner.DeleteTaskInput{User: me, ID: in.ID}); err != nil {
		return NoInput{}, refusal(err)
	}
	return NoInput{}, nil
}

// share is POST /api/tasks/{id}/share's body; what it uses is in wired.
func share(ctx context.Context, in ShareRequest) (out task.Task, err error) {
	me, err := caller(ctx)
	if err != nil {
		return task.Task{}, err
	}
	res, err := wired.shareTask.Execute(ctx, planner.ShareTaskInput{User: me, ID: in.ID, With: in.UserID})
	if err != nil {
		return task.Task{}, refusal(err)
	}
	return res.Task, nil
}

// unshare is DELETE /api/tasks/{id}/share/{userId}'s body; what it uses is in wired.
func unshare(ctx context.Context, in UnshareRequest) (out task.Task, err error) {
	me, err := caller(ctx)
	if err != nil {
		return task.Task{}, err
	}
	res, err := wired.unshareTask.Execute(ctx, planner.UnshareTaskInput{User: me, ID: in.ID, With: in.UserID})
	if err != nil {
		return task.Task{}, refusal(err)
	}
	return res.Task, nil
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

// refusal is a use case's error as the API answers it: 404 for a task or a
// group that does not exist or that the caller does not see, 403 for what only
// its owner does (or an owner or an admin of its group), 409 for sharing with
// who is no contact, 409 for a move its state does not allow, 400 with the
// field at fault for a task that cannot be added or a view that cannot be
// listed; anything else is the framework's to answer.
func refusal(err error) error {
	invalid := func(path, rule string, cause error) error {
		e := app.Invalid("the request is invalid")
		e.Violations = []app.Violation{{Path: path, Rule: rule, Message: cause.Error()}}
		return e
	}
	switch {
	case errors.Is(err, task.ErrNotFound):
		return app.NotFound("no such task")
	case errors.Is(err, task.ErrNoGroup):
		return app.NotFound("no such group")
	case errors.Is(err, task.ErrNotOwner):
		return app.Forbidden("only its owner, or an owner or an admin of its group, does this to a task")
	case errors.Is(err, task.ErrNotContact):
		return app.Conflict("a task is shared with a contact of its owner")
	case errors.Is(err, task.ErrNotAllowed):
		return app.Conflict("the task cannot make this move from the state it is in")
	case errors.Is(err, task.ErrEmptyTitle):
		return invalid("title", "required", err)
	case errors.Is(err, task.ErrTitleTooLong):
		return invalid("title", "maxlen", err)
	case errors.Is(err, task.ErrUnknownPriority):
		return invalid("priority", "oneof", err)
	case errors.Is(err, task.ErrNotMember):
		return invalid("groupId", "member", err)
	case errors.Is(err, task.ErrNotAssignable):
		return invalid("assigneeId", "audience", err)
	case errors.Is(err, task.ErrUnknownView):
		return invalid("view", "oneof", err)
	case errors.Is(err, task.ErrGroupNeeded):
		return invalid("group", "required", err)
	case errors.Is(err, task.ErrBadZone):
		return invalid("tz", "timezone", err)
	}
	return err
}
