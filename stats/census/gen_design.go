// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:27181b66e729f1fb8db820d65217c17c9cc9e0bc6f9b395ba00ff28fca67de63

// Package census — The job that samples the task list, and the HTTP read of
// its samples.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package census

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 4, 1, 255), "NOT_IMPLEMENTED", "not implemented yet", "census: a hand-written body is still a skeleton")

const (
	// Keep is how many samples are kept: an hour, at one every thirty seconds.
	Keep int = 120
)

// Snapshot — The task list at one instant.
type Snapshot struct {
	ID string    `json:"id" app:"public"`
	At time.Time `json:"at" app:"public"`
	// Counts has every state of a task.
	Counts  map[string]int `json:"counts" app:"public"`
	Total   int            `json:"total" app:"public"`
	Overdue int            `json:"overdue" app:"public"`
	// CompletionRate is the share of tasks finished: done or archived.
	CompletionRate float64 `json:"completionRate" app:"public"`
}

// NoInput — A request that says nothing.
type NoInput struct {
}

// Report — The latest sample and the recent history, oldest first.
type Report struct {
	Latest  *Snapshot  `json:"latest"`
	History []Snapshot `json:"history"`
}

// Service is census's framework service: what it declares, drawn and run.
var Service = app.NewService("census", "The job that samples the task list, and the HTTP read of its samples.")

// SnapshotsStore — The latest samples, keyed by a time-ordered id.
var SnapshotsStore = Service.Store("snapshots", func(e Snapshot) string { return e.ID })

var _ = Service.Endpoint("GET /api/stats", Current)

// Current serves GET /api/stats: its body is the hand-written current.
func Current(ctx context.Context, in NoInput) (Report, error) {
	return current(ctx, in)
}

var _ = Service.Every("sample", 30*time.Second, Sample)

// Sample — Counts the tasks and keeps the count.
func Sample(ctx context.Context) error {
	return sample(ctx)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts census calls it
// before its App runs.
func Wire() {

}
