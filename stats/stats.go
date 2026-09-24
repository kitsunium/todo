// Package stats keeps the vital signs of the todo list, sampled by a loop of
// the daemon.
package stats

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/todos"
)

// Service owns the samples.
var Service = kit.NewService("stats", "Vital signs of the todo list, sampled by a loop of the daemon.")

// Snapshot is the todo list at one instant.
type Snapshot struct {
	ID             string         `json:"id"`
	At             time.Time      `json:"at"`
	Counts         map[string]int `json:"counts"`
	Total          int            `json:"total"`
	CompletionRate float64        `json:"completionRate"`
}

// Snapshots keeps the latest samples, keyed by a time-ordered ID.
var Snapshots = Service.Store("snapshots", func(s Snapshot) string { return s.ID })

// The sampling loop.
var _ = Service.Every("sample", 10*time.Second, Sample)

// keep is how many samples are kept.
const keep = 60

// Sample asks the todo list for its census and records it. It goes through
// the todos service's private API rather than its store.
func Sample(ctx context.Context) error {
	census, err := todos.CensusAPI.Call(ctx, kit.Empty{})
	if err != nil {
		return err
	}
	counts := make(map[string]int, len(census.Counts))
	for status, n := range census.Counts {
		counts[string(status)] = n
	}
	rate := 0.0
	if finished := counts[string(todos.Done)] + counts[string(todos.Archived)]; census.Total > 0 {
		rate = float64(finished) / float64(census.Total)
	}
	if err := Snapshots.Put(ctx, Snapshot{
		ID:             kit.NewID("snapshot"),
		At:             time.Now().UTC(),
		Counts:         counts,
		Total:          census.Total,
		CompletionRate: rate,
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

// The stats API.
var _ = Service.Endpoint("GET /stats", Current)

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
	slices.SortFunc(all, func(a, b Snapshot) int { return cmp.Compare(a.At.UnixNano(), b.At.UnixNano()) })
	out := CurrentOutput{History: all}
	if len(all) > 0 {
		out.Latest = &all[len(all)-1]
	}
	return out, nil
}
