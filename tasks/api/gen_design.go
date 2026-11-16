// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:0c616a2ddc211ada253896609000dcaab6758455e2991084b86e140dd79690fe

// Package api — The HTTP API of a member's tasks, and the census the rest of
// the product asks.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package api

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/tasks/planner"
	"github.com/kitsunium/todo/tasks/task"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 1, 4, 255), "NOT_IMPLEMENTED", "not implemented yet", "api: a hand-written body is still a skeleton")

// AddRequest — What a member sends to add a task.
type AddRequest struct {
	Title    string    `json:"title" app:"public"`
	Notes    string    `json:"notes" app:"public"`
	Priority string    `json:"priority" app:"public"`
	Assignee string    `json:"assignee" app:"personal"`
	Due      time.Time `json:"due" app:"public"`
}

// ByID — The task a route names.
type ByID struct {
	ID string `path:"id"`
}

// Census — The tasks counted, every member's together.
type Census struct {
	Counts  map[string]int `json:"counts" app:"public"`
	Total   int            `json:"total" app:"public"`
	Overdue int            `json:"overdue" app:"public"`
}

// EditRequest — What a member changes of a task, each field only when it is
// sent.
type EditRequest struct {
	ID       string     `path:"id"`
	Title    *string    `json:"title" app:"public"`
	Notes    *string    `json:"notes" app:"public"`
	Priority *string    `json:"priority" app:"public"`
	Assignee *string    `json:"assignee" app:"personal"`
	Due      *time.Time `json:"due" app:"public"`
}

// ListRequest — Which tasks to list.
type ListRequest struct {
	State string `query:"state"`
}

// NoInput — A question that needs no word.
type NoInput struct {
}

// ShareRequest — The member a task is shared with.
type ShareRequest struct {
	ID     string `path:"id"`
	UserID string `json:"userId" app:"public"`
}

// TaskList — The tasks asked for.
type TaskList struct {
	Tasks []task.Task `json:"tasks"`
}

// UnshareRequest — The member a task is no longer shared with.
type UnshareRequest struct {
	ID     string `path:"id"`
	UserID string `path:"userId"`
}

// Service is api's framework service: what it declares, drawn and run.
var Service = app.NewService("api", "The HTTP API of a member's tasks, and the census the rest of the product asks.")

var _ = Service.Endpoint("POST /api/tasks", Add, app.Auth(), app.RateLimitPerClient(10, 20))

// Add serves POST /api/tasks: its body is the hand-written add.
func Add(ctx context.Context, in AddRequest) (task.Task, error) {
	return add(ctx, in)
}

var _ = Service.Endpoint("GET /api/tasks", List, app.Auth())

// List serves GET /api/tasks: its body is the hand-written list.
func List(ctx context.Context, in ListRequest) (TaskList, error) {
	return list(ctx, in)
}

var _ = Service.Endpoint("GET /api/tasks/{id}", Get, app.Auth())

// Get serves GET /api/tasks/{id}: its body is the hand-written get.
func Get(ctx context.Context, in ByID) (task.Task, error) {
	return get(ctx, in)
}

var _ = Service.Endpoint("PATCH /api/tasks/{id}", Edit, app.Auth(), app.RateLimitPerClient(10, 20))

// Edit serves PATCH /api/tasks/{id}: its body is the hand-written edit.
func Edit(ctx context.Context, in EditRequest) (task.Task, error) {
	return edit(ctx, in)
}

var _ = Service.Endpoint("DELETE /api/tasks/{id}", Delete, app.Auth(), app.RateLimitPerClient(10, 20))

// Delete serves DELETE /api/tasks/{id}: its body is the hand-written deleteBody.
func Delete(ctx context.Context, in ByID) (NoInput, error) {
	return deleteBody(ctx, in)
}

var _ = Service.Endpoint("POST /api/tasks/{id}/share", Share, app.Auth(), app.RateLimitPerClient(10, 20))

// Share serves POST /api/tasks/{id}/share: its body is the hand-written share.
func Share(ctx context.Context, in ShareRequest) (task.Task, error) {
	return share(ctx, in)
}

var _ = Service.Endpoint("DELETE /api/tasks/{id}/share/{userId}", Unshare, app.Auth(), app.RateLimitPerClient(10, 20))

// Unshare serves DELETE /api/tasks/{id}/share/{userId}: its body is the hand-written unshare.
func Unshare(ctx context.Context, in UnshareRequest) (task.Task, error) {
	return unshare(ctx, in)
}

var _ = Service.Endpoint("POST /api/tasks/{id}/complete", Complete, app.Auth(), app.RateLimitPerClient(10, 20))

// Complete serves POST /api/tasks/{id}/complete: its body is the hand-written complete.
func Complete(ctx context.Context, in ByID) (task.Task, error) {
	return complete(ctx, in)
}

var _ = Service.Endpoint("POST /api/tasks/{id}/reopen", Reopen, app.Auth(), app.RateLimitPerClient(10, 20))

// Reopen serves POST /api/tasks/{id}/reopen: its body is the hand-written reopen.
func Reopen(ctx context.Context, in ByID) (task.Task, error) {
	return reopen(ctx, in)
}

var _ = Service.Endpoint("POST /api/tasks/{id}/archive", Archive, app.Auth(), app.RateLimitPerClient(10, 20))

// Archive serves POST /api/tasks/{id}/archive: its body is the hand-written archive.
func Archive(ctx context.Context, in ByID) (task.Task, error) {
	return archive(ctx, in)
}

var _ = Service.Endpoint("POST /api/tasks/{id}/restore", Restore, app.Auth(), app.RateLimitPerClient(10, 20))

// Restore serves POST /api/tasks/{id}/restore: its body is the hand-written restore.
func Restore(ctx context.Context, in ByID) (task.Task, error) {
	return restore(ctx, in)
}

// CensusQuery — The census: how many tasks there are per state, for the
// components that watch the list.
var CensusQuery = Service.Query("census", Count)

// Count handles the query "census": its body is the hand-written count.
func Count(ctx context.Context, in NoInput) (Census, error) {
	return count(ctx, in)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
	addTask     planner.AddTask
	listTasks   planner.ListTasks
	getTask     planner.GetTask
	editTask    planner.EditTask
	deleteTask  planner.DeleteTask
	shareTask   planner.ShareTask
	unshareTask planner.UnshareTask
	moveTask    planner.MoveTask
	countTasks  planner.CountTasks
}

// Wire gives the handlers what they use: the role that mounts api calls it
// before its App runs.
func Wire(addTask planner.AddTask, listTasks planner.ListTasks, getTask planner.GetTask, editTask planner.EditTask, deleteTask planner.DeleteTask, shareTask planner.ShareTask, unshareTask planner.UnshareTask, moveTask planner.MoveTask, countTasks planner.CountTasks) {
	wired.addTask = addTask
	wired.listTasks = listTasks
	wired.getTask = getTask
	wired.editTask = editTask
	wired.deleteTask = deleteTask
	wired.shareTask = shareTask
	wired.unshareTask = unshareTask
	wired.moveTask = moveTask
	wired.countTasks = countTasks
}
