module github.com/kitsunium/todo

go 1.27.1

require (
	github.com/kitsunium/platform v0.0.0
	github.com/kitsunium/sdk/pkg v0.5.0
)

require (
	github.com/google/pprof v0.0.0-20260802141513-ef3492d7dac3 // indirect
	github.com/kitsunium/sdk/internal/core v0.5.0 // indirect
	github.com/kitsunium/sdk/internal/kernel v0.5.0 // indirect
	github.com/kitsunium/sdk/internal/service v0.5.0 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/tools v0.50.0 // indirect
)

// The platform is developed next to this repository until it is published:
// clone github.com/kitsunium/platform beside it.
replace github.com/kitsunium/platform => ../platform
