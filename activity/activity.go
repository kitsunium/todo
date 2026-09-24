// Package activity is what happened to me: one feed per user, written from
// the events of tasks, contacts and groups, read and marked read by its
// owner.
package activity

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/groups"
	"github.com/kitsunium/todo/identity"
)

// Service owns the feeds.
var Service = kit.NewService("activity", "What happened to me: a feed per user, written from the events of tasks, contacts and groups.")

// TaskRef is the task an entry is about.
type TaskRef struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Entry is one line of one user's feed.
type Entry struct {
	// ID is the event's ID and the recipient's: a redelivered event finds
	// its entries already written.
	ID     string `json:"id"`
	UserID string `json:"userId"`
	// Kind is the event, with its source: task.shared, contact.accepted…
	Kind  string            `json:"kind"`
	Actor *identity.UserRef `json:"actor,omitempty"`
	Task  *TaskRef          `json:"task,omitempty"`
	Group *groups.GroupRef  `json:"group,omitempty"`
	// Text is the entry as one sentence, written for its recipient.
	Text string    `json:"text"`
	At   time.Time `json:"at"`
	Read bool      `json:"read"`
}

// Entries keeps every feed, listed per user.
var Entries = Service.Store("entries", func(e Entry) string { return e.ID },
	kit.Index("user", func(e Entry) []string { return []string{e.UserID} }))

// Keep is how many entries a feed keeps: the oldest go first.
const Keep = 200

// The feed API, and the unread count the task list shows in its sidebar.
var (
	_ = Service.Endpoint("GET /api/activity", Feed, kit.Auth())
	_ = Service.Endpoint("POST /api/activity/read", MarkRead, kit.Auth(), kit.RateLimitPerClient(10, 20))

	// UnreadAPI counts a user's unread entries, for the task counts.
	UnreadAPI = Service.Endpoint("GET /internal/activity/unread", Unread, kit.Private())
)

// FeedEntry is an entry as its owner reads it.
type FeedEntry struct {
	ID    string            `json:"id"`
	Kind  string            `json:"kind"`
	Actor *identity.UserRef `json:"actor,omitempty"`
	Task  *TaskRef          `json:"task,omitempty"`
	Group *groups.GroupRef  `json:"group,omitempty"`
	Text  string            `json:"text"`
	At    time.Time         `json:"at"`
	Read  bool              `json:"read"`
}

// FeedOutput is the caller's feed, newest first.
type FeedOutput struct {
	Entries []FeedEntry `json:"entries"`
}

// me is the signed-in user.
func me(ctx context.Context) (string, error) {
	uid, ok := kit.UserID(ctx)
	if !ok {
		return "", kit.Unauthenticated("sign in first")
	}
	return string(uid), nil
}

// feedOf returns a user's entries, newest first.
func feedOf(ctx context.Context, uid string) ([]Entry, error) {
	entries, err := Entries.Find(ctx, "user", uid)
	if err != nil {
		return nil, err
	}
	slices.SortFunc(entries, func(a, b Entry) int { return cmp.Or(b.At.Compare(a.At), cmp.Compare(b.ID, a.ID)) })
	return entries, nil
}

// Feed returns the caller's feed, newest first.
func Feed(ctx context.Context, _ kit.Empty) (FeedOutput, error) {
	uid, err := me(ctx)
	if err != nil {
		return FeedOutput{}, err
	}
	entries, err := feedOf(ctx, uid)
	if err != nil {
		return FeedOutput{}, err
	}
	out := FeedOutput{Entries: make([]FeedEntry, 0, len(entries))}
	for _, e := range entries {
		out.Entries = append(out.Entries, FeedEntry{ID: e.ID, Kind: e.Kind, Actor: e.Actor, Task: e.Task, Group: e.Group, Text: e.Text, At: e.At, Read: e.Read})
	}
	return out, nil
}

// MarkRead marks the caller's whole feed read.
func MarkRead(ctx context.Context, _ kit.Empty) (kit.Empty, error) {
	uid, err := me(ctx)
	if err != nil {
		return kit.Empty{}, err
	}
	entries, err := Entries.Find(ctx, "user", uid)
	if err != nil {
		return kit.Empty{}, err
	}
	for _, e := range entries {
		if e.Read {
			continue
		}
		if _, err := Entries.Update(ctx, e.ID, func(e *Entry) error { e.Read = true; return nil }); err != nil {
			return kit.Empty{}, err
		}
	}
	return kit.Empty{}, nil
}

// UnreadQuery names a user.
type UnreadQuery struct {
	User string `query:"user"`
}

// UnreadOutput counts a user's unread entries.
type UnreadOutput struct {
	Unread int `json:"unread"`
}

// Unread counts a user's unread entries.
func Unread(ctx context.Context, in UnreadQuery) (UnreadOutput, error) {
	entries, err := Entries.Find(ctx, "user", in.User)
	if err != nil {
		return UnreadOutput{}, err
	}
	n := 0
	for _, e := range entries {
		if !e.Read {
			n++
		}
	}
	return UnreadOutput{Unread: n}, nil
}
