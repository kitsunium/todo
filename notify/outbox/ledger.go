// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package outbox

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
)

// ledger is Ledger's own state: the store is the package's, it holds nothing.
type ledger struct {
}

// newLedger is NewLedger's body.
func newLedger() *Ledger {
	return &Ledger{ledger{}}
}

// mark is Ledger.Mark's body.
func (a *Ledger) mark(ctx context.Context, taskID string) (err error) {
	return RemindedStore.Put(ctx, Reminded{TaskID: taskID, At: app.Now(ctx).UTC().Truncate(time.Millisecond)})
}

// seen is Ledger.Seen's body. It filters instead of reading the key: a
// task not yet reminded of is the usual answer, never a failed read.
func (a *Ledger) seen(ctx context.Context, taskID string) (r0 bool, err error) {
	found, err := RemindedStore.Filter(ctx, func(r Reminded) bool { return r.TaskID == taskID })
	return len(found) > 0, err
}
