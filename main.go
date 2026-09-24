// Command todo is the reference product of kit: a task list for people who
// work together — accounts, contacts, groups, shared tasks, mail — whose
// architecture is its own live diagram.
//
//	kit dev                         # http://localhost:4000 — the app
//	                                # http://localhost:4000/_kit/ — its diagram
//	go run . graph                  # the product graph, as JSON
//	KIT_DATA_DIR=/data ./todo       # production: data on disk, no Studio
package main

import (
	"context"
	"os"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/activity"
	"github.com/kitsunium/todo/contacts"
	"github.com/kitsunium/todo/groups"
	"github.com/kitsunium/todo/identity"
	"github.com/kitsunium/todo/notify"
	"github.com/kitsunium/todo/stats"
	"github.com/kitsunium/todo/tasks"
	"github.com/kitsunium/todo/web"

	// Two services declare part of their building blocks in a second
	// package, because they call services that call them back: tasks/api
	// (the lifecycle and the HTTP API of tasks) and notify/dispatch (who
	// notify mails, when). Importing them is what declares those blocks.
	_ "github.com/kitsunium/todo/notify/dispatch"
	_ "github.com/kitsunium/todo/tasks/api"
)

// services are the bounded contexts of the product, in the order a request
// usually meets them.
func services() []*kit.Service {
	return []*kit.Service{
		identity.Service, contacts.Service, groups.Service, tasks.Service,
		notify.Service, activity.Service, stats.Service, web.Service,
	}
}

// App is the whole product: eight services, one binary.
var App = kit.NewApp("todo", services()...)

func main() {
	os.Exit(App.Main(context.Background(), os.Args[1:]))
}
