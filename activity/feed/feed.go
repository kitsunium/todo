// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package feed

import (
	"cmp"
	"context"
	"slices"
	"sync"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/tasks/task"
)

// feed is Feed's own state: the store is the package's; mu makes one entry
// at a time read the last sequence number and take the next.
type feed struct {
	mu sync.Mutex
}

// newFeed is NewFeed's body.
func newFeed() *Feed {
	return &Feed{}
}

// record is Feed.Record's body: one entry, numbered after the newest one
// and dated as the task was last written, then the oldest entries of that
// member beyond Keep dropped. The number is what orders the feed: two entries written in
// the same millisecond have the same date, and an id says nothing of which
// came first.
func (a *Feed) record(ctx context.Context, kind string, t task.Task) (err error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	all, err := newestFirst(ctx)
	if err != nil {
		return err
	}
	e := Entry{ID: app.NewID("entry"), Seq: 1, UserID: t.OwnerID, Kind: kind, TaskID: t.ID, Title: t.Title, Assignee: t.Assignee, At: t.UpdatedAt}
	if len(all) > 0 {
		e.Seq = all[0].Seq + 1
	}
	if err := EntriesStore.Put(ctx, e); err != nil {
		return err
	}
	kept := 1 // the entry just written
	for _, old := range all {
		if old.UserID != e.UserID {
			continue
		}
		if kept++; kept <= Keep {
			continue
		}
		if err := EntriesStore.Delete(ctx, old.ID); err != nil {
			return err
		}
	}
	return nil
}

// newestFirst is every entry, the newest first.
func newestFirst(ctx context.Context) ([]Entry, error) {
	all, err := EntriesStore.List(ctx)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(all, func(x, y Entry) int { return cmp.Compare(y.Seq, x.Seq) })
	return all, nil
}
