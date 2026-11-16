// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:620afbb2962f72a1a0c57e91267127ee1982802bbd70150e8fa1d8268bc9f808

// Package api — The HTTP API of a member's tasks, the census the rest of the
// product asks, and the job that archives what was done long ago.
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

// ListRequest — Which tasks to list.
type ListRequest struct {
	State string `query:"state"`
}

// NoInput — A question that needs no word.
type NoInput struct {
}

// TaskList — The tasks asked for.
type TaskList struct {
	Tasks []task.Task `json:"tasks"`
}

// Service is api's framework service: what it declares, drawn and run.
var Service = app.NewService("api", "The HTTP API of a member's tasks, the census the rest of the product asks, and the job that archives what was done long ago.")

// ArchiveAfterSetting — How long a done task stays listed before it is
// archived.
var ArchiveAfterSetting = Service.Setting[time.Duration]("archive-after", 24*time.Hour)

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

var _ = Service.Every("archive-done", time.Minute, Sweep)

// Sweep — Archives the tasks done for longer than archive-after.
func Sweep(ctx context.Context) error {
	return sweep(ctx)
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
	addTask             planner.AddTask
	listTasks           planner.ListTasks
	moveTask            planner.MoveTask
	archiveDone         planner.ArchiveDone
	archiveAfterSetting *app.Setting[time.Duration]
	countTasks          planner.CountTasks
}

// Wire gives the handlers what they use: the role that mounts api calls it
// before its App runs.
func Wire(addTask planner.AddTask, listTasks planner.ListTasks, moveTask planner.MoveTask, archiveDone planner.ArchiveDone, archiveAfterSetting *app.Setting[time.Duration], countTasks planner.CountTasks) {
	wired.addTask = addTask
	wired.listTasks = listTasks
	wired.moveTask = moveTask
	wired.archiveDone = archiveDone
	wired.archiveAfterSetting = archiveAfterSetting
	wired.countTasks = countTasks
}
