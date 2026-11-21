// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/identity/account"
)

// oneTimeLinks is OneTimeLinks's own state: the store is the package's, it holds nothing.
type oneTimeLinks struct {
}

// newOneTimeLinks is NewOneTimeLinks's body.
func newOneTimeLinks() *OneTimeLinks {
	return &OneTimeLinks{oneTimeLinks{}}
}

// byHash is OneTimeLinks.ByHash's body: the link a digest opens, through
// the store's index, and whether there is one.
func (a *OneTimeLinks) byHash(ctx context.Context, hash string) (r0 account.Link, r1 bool, err error) {
	found, err := LinksStore.Find(ctx, "by-hash", hash)
	if err != nil || len(found) == 0 {
		return account.Link{}, false, err
	}
	r := found[0]
	return account.Link{ID: r.ID, Hash: r.Hash, Purpose: r.Purpose, UserID: r.UserID, CreatedAt: r.CreatedAt, ExpiresAt: r.ExpiresAt}, true, nil
}

// deleteOf is OneTimeLinks.DeleteOf's body: every link of a member for a
// purpose is deleted, none being no error.
func (a *OneTimeLinks) deleteOf(ctx context.Context, userID string, purpose string) (err error) {
	found, err := LinksStore.Filter(ctx, func(r LinkRecord) bool { return r.UserID == userID && r.Purpose == purpose })
	if err != nil {
		return err
	}
	for _, r := range found {
		if err := gone(LinksStore.Delete(ctx, r.ID)); err != nil {
			return err
		}
	}
	return nil
}

// newID is OneTimeLinks.NewID's body: a time-ordered id.
func (a *OneTimeLinks) newID(ctx context.Context) (r0 string, err error) {
	return app.NewID("link"), nil
}

// save is OneTimeLinks.Save's body.
func (a *OneTimeLinks) save(ctx context.Context, l account.Link) (err error) {
	return LinksStore.Put(ctx, LinkRecord{ID: l.ID, Hash: l.Hash, Purpose: l.Purpose, UserID: l.UserID, CreatedAt: l.CreatedAt, ExpiresAt: l.ExpiresAt})
}
