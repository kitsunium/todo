// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:27ca09ed71b55775793a56323d68cbce853561b10ccbc6f7818f3315d9bdf827

// Package httpapi — The server's HTTP API, called over the network.
//
// Its structure is generated from design/ (ADR 0010); its bodies are the
// hand-written twins beside this file.
package httpapi

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/load/driver"
)

// errNotImplemented is what a skeleton answers until its body is written:
// a twin that still returns it is a stub.
var errNotImplemented = errs.New(errs.Pack(64, 5, 2, 255), "NOT_IMPLEMENTED", "not implemented yet", "httpapi: a hand-written body is still a skeleton")

// Client is API@v1 over HTTP, on the address target names.
type Client struct {
	client // its own state, hand-written
}

// NewClient builds Client: its body is the hand-written newClient.
func NewClient(targetSetting *app.Setting[string]) *Client {
	return newClient(targetSetting)
}

// Activity is load/port/API@v1's.
func (a *Client) Activity(ctx context.Context) (int, error) {
	return a.activity(ctx)
}

// Add is load/port/API@v1's.
func (a *Client) Add(ctx context.Context, title string, priority string, assignee string, due time.Time) (string, error) {
	return a.add(ctx, title, priority, assignee, due)
}

// List is load/port/API@v1's.
func (a *Client) List(ctx context.Context) (int, error) {
	return a.list(ctx)
}

// Move is load/port/API@v1's.
func (a *Client) Move(ctx context.Context, id string, event string) error {
	return a.move(ctx, id, event)
}

// Stats is load/port/API@v1's.
func (a *Client) Stats(ctx context.Context) (int, error) {
	return a.stats(ctx)
}

var _ driver.APIV1 = (*Client)(nil)

// Service is httpapi's framework service: what it declares, drawn and run.
var Service = app.NewService("httpapi", "The server's HTTP API, called over the network.")

// TargetSetting — The host and port of the server the load is driven
// against.
var TargetSetting = Service.Setting[string]("target", "127.0.0.1:4000")

// wired is what the handlers' twins reach, set by the role that mounts
// the component (Wire).
var wired struct {
}

// Wire gives the handlers what they use: the role that mounts httpapi calls it
// before its App runs.
func Wire() {

}
