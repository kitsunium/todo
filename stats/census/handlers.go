// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package census

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/tasks/api"
)

// current is GET /api/stats's body: the samples kept, oldest first, and the
// latest of them.
func current(ctx context.Context, in NoInput) (out Report, err error) {
	all, err := oldestFirst(ctx)
	if err != nil {
		return Report{}, err
	}
	out = Report{History: all}
	if len(all) > 0 {
		out.Latest = &all[len(all)-1]
	}
	return out, nil
}

// sample is the job "sample"'s body: an error is the run's failure. It
// asks the tasks' API for its census: a component's data is its own, keeps the count, and drops the
// oldest samples beyond Keep.
func sample(ctx context.Context) error {
	count, err := api.CensusQuery.Ask(ctx, api.NoInput{})
	if err != nil {
		return err
	}
	rate := 0.0
	if count.Total > 0 {
		rate = float64(count.Counts["done"]+count.Counts["archived"]) / float64(count.Total)
	}
	if err := SnapshotsStore.Put(ctx, Snapshot{
		ID: app.NewID("snapshot"), At: app.Now(ctx).UTC().Truncate(time.Millisecond),
		Counts: count.Counts, Total: count.Total, Overdue: count.Overdue, CompletionRate: rate,
	}); err != nil {
		return err
	}
	all, err := oldestFirst(ctx)
	if err != nil {
		return err
	}
	for _, s := range all[:max(0, len(all)-Keep)] {
		if err := SnapshotsStore.Delete(ctx, s.ID); err != nil {
			return err
		}
	}
	return nil
}

// oldestFirst is every sample kept, the oldest first.
func oldestFirst(ctx context.Context) ([]Snapshot, error) {
	all, err := SnapshotsStore.List(ctx)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(all, func(a, b Snapshot) int { return cmp.Or(a.At.Compare(b.At), cmp.Compare(a.ID, b.ID)) })
	return all, nil
}
