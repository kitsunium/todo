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
	"github.com/kitsunium/todo/contacts/linkstore"
	"github.com/kitsunium/todo/groups/groupstore"
	"github.com/kitsunium/todo/tasks/taskstore"
)

// pageSize is how many entries a read answers when it asks for no number.
const pageSize = 50

// recent is GET /api/activity's body: the caller's newest entries, Limit of them at
// most, fifty when the request names no number.
func recent(ctx context.Context, in PageRequest) (out Page, err error) {
	uid, ok := app.UserID(ctx)
	if !ok {
		return Page{}, app.Unauthenticated("sign in first")
	}
	all, err := newestFirst(ctx)
	if err != nil {
		return Page{}, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = pageSize
	}
	out.Entries = []Entry{}
	for _, e := range all {
		if e.UserID == string(uid) && len(out.Entries) < limit {
			out.Entries = append(out.Entries, e)
		}
	}
	return out, nil
}

// writing makes one entry at a time read the last sequence number and take
// the next.
var writing sync.Mutex

// record is the subscription "tasks"'s body: one entry a message and a
// member it is told to: everyone who sees the task.
func record(ctx context.Context, msg taskstore.Event) error {
	audience := msg.Audience
	if len(audience) == 0 {
		audience = []string{msg.OwnerID}
	}
	for _, user := range audience {
		e := Entry{ID: msg.ID + "." + user, UserID: user, Kind: msg.Kind, TaskID: msg.TaskID, Title: msg.Title, At: msg.At}
		if err := write(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

// recordContact is the subscription "contacts"'s body: one entry a message,
// in the feed of the member it is told to, titled with who did it.
func recordContact(ctx context.Context, msg linkstore.Event) error {
	return write(ctx, Entry{ID: msg.ID, UserID: msg.UserID, Kind: msg.Kind, Title: msg.ActorName, At: msg.At})
}

// recordGroup is the subscription "groups"'s body: one entry a message, in
// the feed of the member it is told to, titled with the group's name.
func recordGroup(ctx context.Context, msg groupstore.Event) error {
	return write(ctx, Entry{ID: msg.ID, UserID: msg.UserID, Kind: msg.Kind, Title: msg.GroupName, At: msg.At})
}

// write keeps an entry, keyed by its event's id (and its reader, when an
// event has several), so a message delivered twice writes nothing the
// second time. The entry is numbered after the
// newest one, and the number is what orders the feed: two entries written
// in the same millisecond have the same date, and an id says nothing of
// which came first. The member's oldest entries beyond Keep are then
// dropped.
func write(ctx context.Context, e Entry) error {
	writing.Lock()
	defer writing.Unlock()
	all, err := newestFirst(ctx)
	if err != nil {
		return err
	}
	e.Seq = 1
	if len(all) > 0 {
		e.Seq = all[0].Seq + 1
	}
	kept := 1 // the entry about to be written
	for _, old := range all {
		if old.ID == e.ID {
			return nil // delivered before
		}
	}
	if err := EntriesStore.Put(ctx, e); err != nil {
		return err
	}
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
