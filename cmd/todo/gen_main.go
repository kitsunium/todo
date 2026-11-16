// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:f78b90bd39d042efd423450d5f8ea59f0f6ddf7d6ea3d6874b6a7bc17cd1eabd

// Command todo — The todo, as one executable, its server and the load it can
// drive against one.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kitsunium/sdk/framework/app"
	_ "github.com/kitsunium/sdk/framework/app/jobs"
	_ "github.com/kitsunium/sdk/framework/app/mail"
	_ "github.com/kitsunium/sdk/framework/app/queue"
	_ "github.com/kitsunium/sdk/framework/app/server"
	_ "github.com/kitsunium/sdk/framework/app/store"
	_ "github.com/kitsunium/sdk/framework/app/workflow"
	loadrole "github.com/kitsunium/todo/cmd/todo/load"
	serverrole "github.com/kitsunium/todo/cmd/todo/server"
)

// Binary is the executable and its process roles (D22).
var Binary = app.NewBinary("todo", "The todo, as one executable, its server and the load it can drive against one.").
	Role("load", loadrole.App, "load").
	Role("server", serverrole.App)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	args := os.Args[1:]
	wire(ctx, args)
	code := Binary.Main(ctx, args)
	stop()
	os.Exit(code)
}

// wire builds what the role the arguments pick needs, before it runs.
func wire(ctx context.Context, args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "load":
			loadrole.Wire(ctx, loadrole.Build{})
			return
		}
	}
	serverrole.Wire(ctx, serverrole.Build{})
}
