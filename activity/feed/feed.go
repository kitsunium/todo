// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package feed

import (
	"cmp"
	"context"
	"slices"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/tasks/task"
)

// feed is Feed's own state: the store is the package's, it holds nothing.
type feed struct {
}

// newFeed is NewFeed's body.
func newFeed() *Feed {
	return &Feed{feed{}}
}

// record is Feed.Record's body: one entry, dated as the task was last
// written, then the oldest entries beyond Keep dropped.
func (a *Feed) record(ctx context.Context, kind string, t task.Task) (err error) {
	if err := EntriesStore.Put(ctx, Entry{
		ID: app.NewID("entry"), Kind: kind, TaskID: t.ID, Title: t.Title, Assignee: t.Assignee, At: t.UpdatedAt,
	}); err != nil {
		return err
	}
	all, err := newestFirst(ctx)
	if err != nil {
		return err
	}
	for _, e := range all[min(Keep, len(all)):] {
		if err := EntriesStore.Delete(ctx, e.ID); err != nil {
			return err
		}
	}
	return nil
}

// newestFirst is every entry, the newest first: by date, then by id, which
// is time-ordered.
func newestFirst(ctx context.Context) ([]Entry, error) {
	all, err := EntriesStore.List(ctx)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(all, func(x, y Entry) int { return cmp.Or(y.At.Compare(x.At), cmp.Compare(y.ID, x.ID)) })
	return all, nil
}
