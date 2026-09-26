module github.com/kitsunium/todo

go 1.27.1

require (
	github.com/kitsunium/platform v0.0.0
	github.com/kitsunium/sdk/pkg v0.6.0
)

require (
	github.com/fxamacker/cbor/v2 v2.9.3 // indirect
	github.com/google/pprof v0.0.0-20260802141513-ef3492d7dac3 // indirect
	github.com/kitsunium/sdk/internal/core v0.6.0 // indirect
	github.com/kitsunium/sdk/internal/kernel v0.6.0 // indirect
	github.com/kitsunium/sdk/internal/service v0.6.0 // indirect
	github.com/pelletier/go-toml/v2 v2.4.3 // indirect
	github.com/vmihailenco/msgpack/v5 v5.4.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.mongodb.org/mongo-driver v1.17.9 // indirect
	golang.org/x/mod v0.41.0 // indirect
	golang.org/x/sync v0.23.0 // indirect
	golang.org/x/tools v0.50.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// The platform is developed next to this repository until it is published:
// clone github.com/kitsunium/platform beside it.
replace github.com/kitsunium/platform => ../platform
