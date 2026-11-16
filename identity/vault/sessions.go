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
	return sessionOf(found[0]), true, nil
}

// of is Sessions.Of's body: every session of a member.
func (a *Sessions) of(ctx context.Context, userID string) (r0 []account.Session, err error) {
	found, err := SessionsStore.Filter(ctx, func(r SessionRecord) bool { return r.UserID == userID })
	if err != nil {
		return nil, err
	}
	r0 = make([]account.Session, len(found))
	for i, r := range found {
		r0[i] = sessionOf(r)
	}
	return r0, nil
}

// sessionOf is the session a record keeps.
func sessionOf(r SessionRecord) account.Session {
	return account.Session{
		ID: r.ID, UserID: r.UserID, TokenHash: r.TokenHash, UserAgent: r.UserAgent, IP: r.IP,
		CreatedAt: r.CreatedAt, LastSeenAt: r.LastSeenAt, ExpiresAt: r.ExpiresAt,
	}
}

// deleteBody is Sessions.Delete's body: a session already gone is deleted.
func (a *Sessions) deleteBody(ctx context.Context, id string) (err error) {
	return gone(SessionsStore.Delete(ctx, id))
}

// deleteOf is Sessions.DeleteOf's body: every session of a member but keep
// ends ("" keeps none).
func (a *Sessions) deleteOf(ctx context.Context, userID string, keep string) (err error) {
	found, err := SessionsStore.Filter(ctx, func(r SessionRecord) bool { return r.UserID == userID && r.ID != keep })
	if err != nil {
		return err
	}
	for _, r := range found {
		if err := gone(SessionsStore.Delete(ctx, r.ID)); err != nil {
			return err
		}
	}
	return nil
}

// gone is the error of a delete, a record already gone being none.
func gone(err error) error {
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
		ID: s.ID, UserID: s.UserID, TokenHash: s.TokenHash, UserAgent: s.UserAgent, IP: s.IP,
		CreatedAt: s.CreatedAt, LastSeenAt: s.LastSeenAt, ExpiresAt: s.ExpiresAt,
	})
}
