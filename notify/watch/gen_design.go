// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:a531db8fdf9b03e4efb3cfcdc2c2b563c368ecbbcb719b2c42901921e493400b

// Package watch — The job that sends the reminders.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package watch

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/notify/reminders"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 3, 3, 255), "NOT_IMPLEMENTED", "not implemented yet", "watch: a hand-written body is still a skeleton")

// Service is watch's framework service: what it declares, drawn and run.
var Service = app.NewService("watch", "The job that sends the reminders.")

// RemindWithinSetting — How long before its due date a task's owner is
// reminded.
var RemindWithinSetting = Service.Setting[time.Duration]("remind-within", 24*time.Hour)

var _ = Service.Every("remind-due", time.Minute, Remind)

// Remind — Reminds the owners of the tasks due within remind-within.
func Remind(ctx context.Context) error {
	return remind(ctx)
}

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
	remindDue           reminders.RemindDue
	remindWithinSetting *app.Setting[time.Duration]
}

// Wire gives the handlers what they use: the role that mounts watch calls it
// before its App runs.
func Wire(remindDue reminders.RemindDue, remindWithinSetting *app.Setting[time.Duration]) {
	wired.remindDue = remindDue
	wired.remindWithinSetting = remindWithinSetting
}
