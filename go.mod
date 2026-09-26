module github.com/kitsunium/todo

go 1.27.1

require (
	github.com/kitsunium/platform v0.0.0
	github.com/kitsunium/sdk/pkg v0.8.0
)

require (
	github.com/kitsunium/sdk/internal/core v0.8.0 // indirect
	github.com/kitsunium/sdk/internal/kernel v0.8.0 // indirect
	github.com/kitsunium/sdk/internal/service v0.8.0 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/tools v0.50.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// The platform is developed next to this repository until it is published:
// clone github.com/kitsunium/platform beside it.
replace github.com/kitsunium/platform => ../platform
