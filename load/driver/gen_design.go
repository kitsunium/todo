// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:a531db8fdf9b03e4efb3cfcdc2c2b563c368ecbbcb719b2c42901921e493400b

// Package driver — What a load does, and the API it needs.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package driver

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/pkg/v1/errs"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 5, 1, 255), "NOT_IMPLEMENTED", "not implemented yet", "driver: a hand-written body is still a skeleton")

var (
	// ErrNoRate is "a load needs a rate above zero".
	ErrNoRate = errs.New(errs.Pack(64, 5, 1, 1), "NO_RATE", "a load needs a rate above zero", "load/error/ErrNoRate")
)

// APIV1 is the server's HTTP API, as the load calls it.
type APIV1 interface {
	Activity(ctx context.Context) (int, error)
	Add(ctx context.Context, title string, priority string, due time.Time) (string, error)
	Enter(ctx context.Context, email string, name string, passphrase string) error
	List(ctx context.Context) (int, error)
	Move(ctx context.Context, id string, event string) error
	Stats(ctx context.Context) (int, error)
}

// Drive signs in as the load's own member, then adds tasks, moves them through
// their lifecycle and reads the lists, at a rate, for a duration.
type Drive interface {
	Execute(ctx context.Context, in DriveInput) (DriveOutput, error)
}

// DriveInput is Drive's input: a field added breaks no caller.
type DriveInput struct {
	For  time.Duration
	Rate int
}

// DriveOutput is Drive's output.
type DriveOutput struct {
	Requests int
	Failures int
}

// DriveHandler runs Drive: its body is the hand-written execute.
type DriveHandler struct {
	api APIV1
}

// NewDrive builds Drive on what it calls.
func NewDrive(api APIV1) *DriveHandler {
	return &DriveHandler{api: api}
}

// Execute runs Drive.
func (h *DriveHandler) Execute(ctx context.Context, in DriveInput) (DriveOutput, error) {
	return h.execute(ctx, &in)
}

var _ Drive = (*DriveHandler)(nil)
