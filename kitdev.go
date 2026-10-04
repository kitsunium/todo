//go:build kitdev

// What the todo's dev build links beside the release's: kit dev builds the
// product with -tags kitdev. The framework serves the Studio's API at /_kit/
// — the graph and its live events, traces, the source, the mailbox and the
// other dev tools — only in a program that imports its studio subsystem; a
// release build links none of it, and production refuses /_kit/ either way.

package main

import _ "github.com/kitsunium/sdk/framework/kit/studio"
