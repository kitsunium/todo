// Package web serves the todo list's user interface: one page talking to the
// todos, activity and stats APIs.
package web

import (
	"embed"

	"github.com/kitsunium/platform/kit"
)

//go:embed assets
var assets embed.FS

// Service owns the user interface.
var Service = kit.NewService("web", "The user interface: one page talking to the todos, activity and stats APIs.")

// UI is the single page application, served at the root.
var UI = Service.Static("ui", "/", assets, kit.Root("assets"))
