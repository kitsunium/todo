// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:b8c6fb4903e1368054d808d160797f66a4b4f241102b577f4a124db5ec43a883

// Package feed — The feed, written from the tasks' events, and its HTTP
// read.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package feed

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/tasks/taskstore"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 2, 1, 255), "NOT_IMPLEMENTED", "not implemented yet", "feed: a hand-written body is still a skeleton")

const (
	// Keep is how many entries the feed keeps for a member.
	Keep int = 200
)

// Entry — One thing that happened to a task.
type Entry struct {
	ID string `json:"id" app:"public"`
	// Seq orders the entries: each is one above the entry written before it.
	Seq int64 `json:"seq" app:"public"`
	// UserID is the member whose feed the entry is in: the task's owner.
	UserID string `json:"userId" app:"public"`
	// Kind is added, or the lifecycle event that moved the task.
	Kind     string    `json:"kind" app:"public"`
	TaskID   string    `json:"taskId" app:"public"`
	Title    string    `json:"title" app:"public"`
	Assignee string    `json:"assignee" app:"personal,subject"`
	At       time.Time `json:"at" app:"public"`
}

// Page — The latest entries, newest first.
type Page struct {
	Entries []Entry `json:"entries"`
}

// PageRequest — How many entries to read.
type PageRequest struct {
	Limit int `query:"limit"`
}

// Service is feed's framework service: what it declares, drawn and run.
var Service = app.NewService("feed", "The feed, written from the tasks' events, and its HTTP read.")

// EntriesStore — The feed's entries, keyed by the id of the event each one
// tells; an entry is deleted ninety days after it was written.
var EntriesStore = Service.Store("entries", func(e Entry) string { return e.ID },
	app.DeleteAfter[Entry](time.Duration(7776000000000000) /* 2160h */, func(e Entry) (time.Time, bool) { return e.At, !e.At.IsZero() }),
)

var _ = Service.Endpoint("GET /api/activity", Recent, app.Auth())

// Recent serves GET /api/activity: its body is the hand-written recent.
func Recent(ctx context.Context, in PageRequest) (Page, error) {
	return recent(ctx, in)
}

var _ = Service.Subscribe("tasks", taskstore.EventsTopic, Record)

// Record takes each message of the topic "events", at least once: its body is
// the hand-written record, which is idempotent.
func Record(ctx context.Context, msg taskstore.Event) error {
	return record(ctx, msg)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts feed calls it
// before its App runs.
func Wire() {

}
