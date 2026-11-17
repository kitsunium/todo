// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:6a4c0cc7ef82f45cbf05f7471fb6e2644c468fdaa00e78e6aa4e3dd63779cfee

// Package load — Drives a load against a running server over its HTTP API,
// then exits.
package load

import (
	"context"
	"embed"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/framework/model"
	"github.com/kitsunium/todo/load/driver"
	"github.com/kitsunium/todo/load/httpapi"
	"github.com/kitsunium/todo/load/loadcli"
)

// configFiles are role load's configuration files, embedded from config: they
// are hand-written.
//
//go:embed config
var configFiles embed.FS

// App is role load: the framework App of httpapi, loadcli, profile cli.
var App = app.NewApp("load", httpapi.Service, loadcli.Service).With(app.Profile(model.ProfileCLI), app.ConfigFiles(configFiles))

// Build is what the binary's build says: its variables, set by -ldflags -X.
type Build struct {
}

// Wire builds what role load's handlers use, and gives it to them.
func Wire(ctx context.Context, build Build) {
	client := httpapi.NewClient(httpapi.TargetSetting)
	drive := driver.NewDrive(client)
	httpapi.Wire()
	loadcli.Wire(drive)
}
