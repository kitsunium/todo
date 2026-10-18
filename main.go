// Command todo is a task list for people who work together — accounts,
// contacts, groups, shared tasks, mail — whose architecture is its own live
// diagram.
//
//	TODO_ENV=dev go run .           # http://localhost:4000 — the app
//	go run . graph                  # the product graph, as JSON
//	TODO_DATA_DIR=/data ./todo      # production: data on disk
package main

import (
	"context"
	"os"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/activity"
	"github.com/kitsunium/todo/contacts"
	"github.com/kitsunium/todo/groups"
	"github.com/kitsunium/todo/identity"
	"github.com/kitsunium/todo/notify"
	"github.com/kitsunium/todo/stats"
	"github.com/kitsunium/todo/tasks"
	"github.com/kitsunium/todo/web"

	// The todo serves HTTP — the server profile, the default. The framework
	// links its HTTP engine only into a program that imports this package,
	// and refuses to start a server that does not (profile.server-missing).
	_ "github.com/kitsunium/sdk/framework/app/server"

	// The subsystems the services declare, each linked only by its import
	// (the start refuses a declaration whose subsystem is not imported):
	// the stores, the topics and their subscriptions, the workflows, the
	// mailer's outbox and transport, and the stats job.
	_ "github.com/kitsunium/sdk/framework/app/jobs"
	_ "github.com/kitsunium/sdk/framework/app/mail"
	_ "github.com/kitsunium/sdk/framework/app/queue"
	_ "github.com/kitsunium/sdk/framework/app/store"
	_ "github.com/kitsunium/sdk/framework/app/workflow"

	// Two services declare part of their building blocks in a second
	// package, because they call services that call them back: tasks/api
	// (the lifecycle and the HTTP API of tasks) and notify/dispatch (who
	// notify mails, when). Importing them is what declares those blocks.
	_ "github.com/kitsunium/todo/notify/dispatch"
	_ "github.com/kitsunium/todo/tasks/api"
)

// services are the bounded contexts of the product, in the order a request
// usually meets them.
func services() []*app.Service {
	return []*app.Service{
		identity.Service, contacts.Service, groups.Service, tasks.Service,
		notify.Service, activity.Service, stats.Service, web.Service,
	}
}

// App is the whole product: eight services, one binary.
var App = app.NewApp("todo", services()...)

func main() {
	os.Exit(App.Main(context.Background(), os.Args[1:]))
}
