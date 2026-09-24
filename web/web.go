// Package web serves the todo's user interface: a single page application
// written in web/src, built by Vite into web/dist and embedded in the binary.
//
//	cd web && npm ci && npm run build    # refresh web/dist
//	cd web && npm run dev:mock           # develop without the backend
package web

import (
	"embed"

	"github.com/kitsunium/platform/kit"
)

// dist is the committed production build. Vite never emits a file whose
// name starts with "." or "_", which //go:embed would silently drop.
//
//go:embed dist
var dist embed.FS

// Service owns the user interface.
var Service = kit.NewService("web", "The user interface: a single page application talking to the product's APIs.\n\nfr: L’interface utilisateur : une application monopage qui parle aux API du produit.")

// App is the single page application, served at the root. A path with no
// file extension that matches no file serves index.html, so the client-side
// routes (/login, /app/today, /app/tasks/{id}…) work on a reload.
//
// fr: App est l’application monopage, servie à la racine. Un chemin sans
// extension qui ne correspond à aucun fichier sert index.html, pour que les
// routes côté client (/login, /app/today, /app/tasks/{id}…) fonctionnent au
// rechargement de la page.
var App = Service.Static("app", "/", dist, kit.Root("dist"))
