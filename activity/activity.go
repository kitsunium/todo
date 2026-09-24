// Package activity keeps a feed of everything that happened to the todo list.
package activity

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/todos"
)

// Service owns the activity feed.
var Service = kit.NewService("activity", "A feed of everything that happened to the todo list.")

// Entry is one line of the feed.
type Entry struct {
	ID     string    `json:"id"`
	TodoID string    `json:"todoId"`
	Title  string    `json:"title"`
	Event  string    `json:"event"`
	From   string    `json:"from,omitempty"`
	To     string    `json:"to"`
	At     time.Time `json:"at"`
}

// Feed keeps the entries, keyed by ID.
var Feed = Service.Store("feed", func(e Entry) string { return e.ID })

// The feed listens to every lifecycle change of the todo list.
var _ = Service.Subscribe("record", todos.Changes, Record)

// keep is how many entries the feed keeps: every write persists the whole
// feed, so it must not grow without bound.
const keep = 200

// Record turns a lifecycle change into a feed entry. Delivery is at least
// once, so the entry is keyed by the change's own ID: a redelivered change
// overwrites its entry instead of adding a second one.
func Record(ctx context.Context, c todos.Change) error {
	if err := Feed.Put(ctx, Entry{
		ID:     c.ID,
		TodoID: c.TodoID,
		Title:  c.Title,
		Event:  c.Event,
		From:   string(c.From),
		To:     string(c.To),
		At:     c.At,
	}); err != nil {
		return err
	}
	return prune(ctx)
}

// prune drops the oldest entries beyond keep.
func prune(ctx context.Context) error {
	n, err := Feed.Count(ctx)
	if err != nil || n <= keep {
		return err
	}
	all, err := Feed.List(ctx)
	if err != nil {
		return err
	}
	slices.SortFunc(all, func(a, b Entry) int { return cmp.Compare(a.At.UnixNano(), b.At.UnixNano()) })
	for _, e := range all[:len(all)-keep] {
		if err := Feed.Delete(ctx, e.ID); err != nil {
			return err
		}
	}
	return nil
}

// The feed's API.
var _ = Service.Endpoint("GET /activity", Recent)

// RecentOutput is the latest entries, newest first.
type RecentOutput struct {
	Entries []Entry `json:"entries"`
}

// recentSize is how many entries Recent returns.
const recentSize = 30

// Recent returns the latest entries, newest first.
func Recent(ctx context.Context, _ kit.Empty) (RecentOutput, error) {
	all, err := Feed.List(ctx)
	if err != nil {
		return RecentOutput{}, err
	}
	slices.SortFunc(all, func(a, b Entry) int { return cmp.Compare(b.At.UnixNano(), a.At.UnixNano()) })
	return RecentOutput{Entries: all[:min(len(all), recentSize)]}, nil
}
