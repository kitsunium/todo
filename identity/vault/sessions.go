// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"
	"errors"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/identity/account"
)

// sessions is Sessions's own state: the store is the package's, it holds nothing.
type sessions struct {
}

// newSessions is NewSessions's body.
func newSessions() *Sessions {
	return &Sessions{sessions{}}
}

// byTokenHash is Sessions.ByTokenHash's body: the session a digest opens,
// through the store's index, and whether there is one.
func (a *Sessions) byTokenHash(ctx context.Context, hash string) (r0 account.Session, r1 bool, err error) {
	found, err := SessionsStore.Find(ctx, "by-token", hash)
	if err != nil || len(found) == 0 {
		return account.Session{}, false, err
	}
	r := found[0]
	return account.Session{
		ID: r.ID, UserID: r.UserID, TokenHash: r.TokenHash, CreatedAt: r.CreatedAt, LastSeenAt: r.LastSeenAt, ExpiresAt: r.ExpiresAt,
	}, true, nil
}

// deleteBody is Sessions.Delete's body: a session already gone is deleted.
func (a *Sessions) deleteBody(ctx context.Context, id string) (err error) {
	err = SessionsStore.Delete(ctx, id)
	var refused *app.Error
	if errors.As(err, &refused) && refused.Code == app.WireNotFound {
		return nil
	}
	return err
}

// newID is Sessions.NewID's body: a time-ordered id.
func (a *Sessions) newID(ctx context.Context) (r0 string, err error) {
	return app.NewID("session"), nil
}

// save is Sessions.Save's body: the session replaces what its id held.
func (a *Sessions) save(ctx context.Context, s account.Session) (err error) {
	return SessionsStore.Put(ctx, SessionRecord{
		ID: s.ID, UserID: s.UserID, TokenHash: s.TokenHash, CreatedAt: s.CreatedAt, LastSeenAt: s.LastSeenAt, ExpiresAt: s.ExpiresAt,
	})
}
