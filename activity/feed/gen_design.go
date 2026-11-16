// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:370a82d053439cd8e17f0439915150946927d383504b60ee20d20c234deff80b

// Package feed — The feed, written from what the tasks' use cases record,
// and its HTTP read.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package feed

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
var errNotImplemented = errs.New(errs.Pack(64, 2, 1, 255), "NOT_IMPLEMENTED", "not implemented yet", "feed: a hand-written body is still a skeleton")

const (
	// Keep is how many entries the feed keeps.
	Keep int = 200
)

// Entry — One thing that happened to a task.
type Entry struct {
	ID string `json:"id" app:"public"`
	// Seq orders the entries: each is one above the entry written before it.
	Seq int64 `json:"seq" app:"public"`
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

// Feed is Journal@v1 on the store "entries"; it keeps the latest Keep entries.
type Feed struct {
	feed // its own state, hand-written
}

// NewFeed builds Feed: its body is the hand-written newFeed.
func NewFeed() *Feed {
	return newFeed()
}

// Record is tasks/port/Journal@v1's.
func (a *Feed) Record(ctx context.Context, kind string, t task.Task) error {
	return a.record(ctx, kind, t)
}

var _ planner.JournalV1 = (*Feed)(nil)

// Service is feed's framework service: what it declares, drawn and run.
var Service = app.NewService("feed", "The feed, written from what the tasks' use cases record, and its HTTP read.")

// EntriesStore — The feed's entries, keyed by their id; an entry is deleted
// ninety days after it was written.
var EntriesStore = Service.Store("entries", func(e Entry) string { return e.ID },
	app.DeleteAfter[Entry](time.Duration(7776000000000000) /* 2160h */, func(e Entry) (time.Time, bool) { return e.At, !e.At.IsZero() }),
)

var _ = Service.Endpoint("GET /api/activity", Recent)

// Recent serves GET /api/activity: its body is the hand-written recent.
func Recent(ctx context.Context, in PageRequest) (Page, error) {
	return recent(ctx, in)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts feed calls it
// before its App runs.
func Wire() {

}
