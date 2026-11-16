// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:20b78385bcc831b04256ba24aa92ec0b4aec39dda784dd51beea005fcb21ccec

// Package site — The web app, a few static files that talk to the HTTP API
// with the session cookie.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package site

import (
	"embed"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 7, 1, 255), "NOT_IMPLEMENTED", "not implemented yet", "site: a hand-written body is still a skeleton")

// Service is site's framework service: what it declares, drawn and run.
var Service = app.NewService("site", "The web app, a few static files that talk to the HTTP API with the session cookie.")

// appFiles are the files of the frontend "app", embedded from dist: the directory
// is hand-written, or a build's output.
//
//go:embed dist
var appFiles embed.FS

// AppFrontend — The web app: one page, its style and its script, written by
// hand, with no build step.
var AppFrontend = Service.Static("app", "/", appFiles, app.Root("dist"))

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts site calls it
// before its App runs.
func Wire() {

}
