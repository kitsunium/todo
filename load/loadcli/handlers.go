// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package loadcli

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/load/driver"
)

// run is the command "run"'s body: its exit status. It drives the load for
// -for at -rate calls a second and prints what it counted: 0 when every
// call was answered, 1 when some failed, 2 when the arguments are wrong.
func run(ctx context.Context, args []string, std app.Stdio) int {
	fs := flag.NewFlagSet("load", flag.ContinueOnError)
	fs.SetOutput(std.Err)
	length := fs.Duration("for", 30*time.Second, "how long the load lasts")
	rate := fs.Int("rate", 5, "calls a second")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	res, err := wired.drive.Execute(ctx, driver.DriveInput{For: *length, Rate: *rate})
	if err != nil {
		_, _ = fmt.Fprintf(std.Err, "load: %v\n", err)
		return 2
	}
	_, _ = fmt.Fprintf(std.Out, "load: %d call(s), %d failed\n", res.Requests, res.Failures)
	if res.Failures > 0 {
		return 1
	}
	return 0
}
