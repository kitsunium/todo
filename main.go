// Command todo is the first product built with kit: a todo list whose
// architecture is its own diagram.
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
	"github.com/kitsunium/todo/stats"
	"github.com/kitsunium/todo/todos"
	"github.com/kitsunium/todo/web"
)

// App is the whole product: four services, one binary.
var App = kit.NewApp("todo", todos.Service, activity.Service, stats.Service, web.Service)

func main() {
	os.Exit(App.Main(context.Background(), os.Args[1:]))
}
