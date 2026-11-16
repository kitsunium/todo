// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:92f401eee68f7ed970fa67dd3972c2db2225ff89c62bf5694e285eed507fe8ce

// Package feed — The feed, written from the events of the tasks, of the
// contacts and of the groups, and its HTTP read.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package feed

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/contacts/linkstore"
	"github.com/kitsunium/todo/groups/groupstore"
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
	// UserID is the member whose feed the entry is in: whoever the event is told
	// to, and the subject of the entry.
	UserID string `json:"userId" app:"personal,subject"`
	// Kind is added or the lifecycle event that moved a task, contact-requested or
	// contact-accepted for a contact, group-invited, group-joined, group-left,
	// group-removed, group-role or group-deleted for a group.
	Kind   string `json:"kind" app:"public"`
	TaskID string `json:"taskId" app:"public"`
	// Title is the task's title, the name of the member a contact entry is about,
	// or the name of the group a group entry is about.
	Title string    `json:"title" app:"personal"`
	At    time.Time `json:"at" app:"public"`
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
var Service = app.NewService("feed", "The feed, written from the events of the tasks, of the contacts and of the groups, and its HTTP read.")

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

var _ = Service.Subscribe("contacts", linkstore.EventsTopic, RecordContact)

// RecordContact takes each message of the topic "events", at least once: its
// body is the hand-written recordContact, which is idempotent.
func RecordContact(ctx context.Context, msg linkstore.Event) error {
	return recordContact(ctx, msg)
}

var _ = Service.Subscribe("groups", groupstore.EventsTopic, RecordGroup)

// RecordGroup takes each message of the topic "events", at least once: its
// body is the hand-written recordGroup, which is idempotent.
func RecordGroup(ctx context.Context, msg groupstore.Event) error {
	return recordGroup(ctx, msg)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts feed calls it
// before its App runs.
func Wire() {

}
