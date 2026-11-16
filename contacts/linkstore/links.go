// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package linkstore

import (
	"context"
	"errors"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/contacts/link"
)

// links is Links's own state: the store is the package's, it holds nothing.
type links struct {
}

// newLinks is NewLinks's body.
func newLinks() *Links {
	return &Links{links{}}
}

// byPair is Links.ByPair's body: the one link of a pair, through the
// store's index, and whether there is one.
func (a *Links) byPair(ctx context.Context, pair string) (r0 link.Link, r1 bool, err error) {
	found, err := LinksStore.Find(ctx, "by-pair", pair)
	if err != nil || len(found) == 0 {
		return link.Link{}, false, err
	}
	return linkOf(found[0]), true, nil
}

// deleteBody is Links.Delete's body: a link already gone is deleted.
func (a *Links) deleteBody(ctx context.Context, id string) (err error) {
	err = LinksStore.Delete(ctx, id)
	var refused *app.Error
	if errors.As(err, &refused) && refused.Code == app.WireNotFound {
		return nil
	}
	return err
}

// get is Links.Get's body. It filters instead of reading the key: a link
// that is not there is an answer, never a failed read.
func (a *Links) get(ctx context.Context, id string) (r0 link.Link, r1 bool, err error) {
	found, err := LinksStore.Filter(ctx, func(r Record) bool { return r.ID == id })
	if err != nil || len(found) == 0 {
		return link.Link{}, false, err
	}
	return linkOf(found[0]), true, nil
}

// newID is Links.NewID's body: a time-ordered id.
func (a *Links) newID(ctx context.Context) (r0 string, err error) {
	return app.NewID("link"), nil
}

// of is Links.Of's body: the links a member is at an end of.
func (a *Links) of(ctx context.Context, user string) (r0 []link.Link, err error) {
	found, err := LinksStore.Filter(ctx, func(r Record) bool { return r.RequesterID == user || r.AddresseeID == user })
	if err != nil {
		return nil, err
	}
	r0 = make([]link.Link, len(found))
	for i, r := range found {
		r0[i] = linkOf(r)
	}
	return r0, nil
}

// save is Links.Save's body: the link replaces what its id held.
func (a *Links) save(ctx context.Context, l link.Link) (err error) {
	return LinksStore.Put(ctx, recordOf(l))
}

// recordOf is a link as the store keeps it.
func recordOf(l link.Link) Record {
	return Record{
		ID: l.ID, Pair: l.Pair, RequesterID: l.RequesterID, AddresseeID: l.AddresseeID,
		State: WaitingState(l.State), CreatedAt: l.CreatedAt, UpdatedAt: l.UpdatedAt,
	}
}

// linkOf is the link a record keeps.
func linkOf(r Record) link.Link {
	return link.Link{
		ID: r.ID, Pair: r.Pair, RequesterID: r.RequesterID, AddresseeID: r.AddresseeID,
		State: link.RequestsState(r.State), CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

// change is Links.Change's body: the store applies the change to the record it
// holds at that instant, alone among its writers, so two changes made at
// once both stay. A link that is not there is ErrNotFound. The change's own error keeps nothing.
func (a *Links) change(ctx context.Context, id string, change func(*link.Link) error) (r0 link.Link, err error) {
	rec, err := LinksStore.Update(ctx, id, func(r *Record) error {
		v := linkOf(*r)
		if err := change(&v); err != nil {
			return err
		}
		v.ID = r.ID // a change never renames a record
		*r = recordOf(v)
		return nil
	})
	var refused *app.Error
	switch {
	case errors.As(err, &refused) && refused.Code == app.WireNotFound:
		return link.Link{}, link.ErrNotFound
	case err != nil:
		return link.Link{}, err
	}
	return linkOf(rec), nil
}
