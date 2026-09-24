// Package stats keeps the vital signs of the task list, sampled by a job of
// the daemon through the tasks service's private census.
package stats

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/tasks"
)

// Service owns the samples.
var Service = kit.NewService("stats", "Vital signs of the task list, sampled by a job of the daemon.")

// Snapshot is the task list at one instant.
type Snapshot struct {
	ID string    `json:"id"`
	At time.Time `json:"at"`
	// Counts has every status of a task.
	Counts map[string]int `json:"counts"`
	Total  int            `json:"total"`
	// CompletionRate is the share of tasks finished: done or archived.
	CompletionRate float64 `json:"completionRate"`
	// OpenByPriority counts the tasks left to do per priority.
	OpenByPriority map[string]int `json:"openByPriority"`
	// Shared counts the tasks shared with someone; Grouped those on a
	// group's list.
	Shared  int `json:"shared"`
	Grouped int `json:"grouped"`
}

// Snapshots keeps the latest samples, keyed by a time-ordered ID.
var Snapshots = Service.Store("snapshots", func(s Snapshot) string { return s.ID })

// The sampling job: a piece of the daemon's internal loop.
var _ = Service.Every("sample", 30*time.Second, Sample)

// keep is how many samples are kept: an hour, at one every 30 seconds.
const keep = 120

// Sample asks the tasks service for its census and records it. It goes
// through the census endpoint rather than the store: a service's data is
// its own.
func Sample(ctx context.Context) error {
	census, err := tasks.CensusAPI.Call(ctx, kit.Empty{})
	if err != nil {
		return err
	}
	counts := make(map[string]int, len(census.Counts))
	for status, n := range census.Counts {
		counts[string(status)] = n
	}
	rate := 0.0
	if census.Total > 0 {
		rate = float64(counts[string(tasks.Done)]+counts[string(tasks.Archived)]) / float64(census.Total)
	}
	if err := Snapshots.Put(ctx, Snapshot{
		ID: kit.NewID("snapshot"), At: wire.Now(ctx), Counts: counts, Total: census.Total, CompletionRate: rate,
		OpenByPriority: census.OpenByPriority, Shared: census.Shared, Grouped: census.Grouped,
	}); err != nil {
		return err
	}
	return prune(ctx)
}

// prune drops the oldest samples beyond keep.
func prune(ctx context.Context) error {
	all, err := Snapshots.List(ctx)
	if err != nil {
		return err
	}
	for _, s := range all[:max(0, len(all)-keep)] {
		if err := Snapshots.Delete(ctx, s.ID); err != nil {
			return err
		}
	}
	return nil
}

// The stats API: public, like a status page.
var _ = Service.Endpoint("GET /api/stats", Current)

// CurrentOutput is the latest sample and the recent history, oldest first.
type CurrentOutput struct {
	Latest  *Snapshot  `json:"latest"`
	History []Snapshot `json:"history"`
}

// Current returns the latest sample and the recent history.
func Current(ctx context.Context, _ kit.Empty) (CurrentOutput, error) {
	all, err := Snapshots.List(ctx)
	if err != nil {
		return CurrentOutput{}, err
	}
	slices.SortFunc(all, func(a, b Snapshot) int { return cmp.Or(a.At.Compare(b.At), cmp.Compare(a.ID, b.ID)) })
	out := CurrentOutput{History: all}
	if len(all) > 0 {
		out.Latest = &all[len(all)-1]
	}
	return out, nil
}
