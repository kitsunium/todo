// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:d9e6eb32619ad64ff259ee580b4b28ea598fe9a15bc77181628dbbd887d57f92

// Package loadcli — The command that drives the load.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package loadcli

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/load/driver"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 5, 3, 255), "NOT_IMPLEMENTED", "not implemented yet", "loadcli: a hand-written body is still a skeleton")

// Service is loadcli's framework service: what it declares, drawn and run.
var Service = app.NewService("loadcli", "The command that drives the load.")

var _ = Service.CLI("run", "drives a load against the server: -for DURATION (default 30s), -rate N a second (default 5)", Run, app.DefaultCommand())

// Run runs "run": its body is the hand-written run.
func Run(ctx context.Context, args []string, std app.Stdio) int {
	return run(ctx, args, std)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
	drive driver.Drive
}

// Wire gives the handlers what they use: the role that mounts loadcli calls it
// before its App runs.
func Wire(drive driver.Drive) {
	wired.drive = drive
}
