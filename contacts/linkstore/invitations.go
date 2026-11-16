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

// invitations is Invitations's own state: the store is the package's, it holds nothing.
type invitations struct {
}

// newInvitations is NewInvitations's body.
func newInvitations() *Invitations {
	return &Invitations{invitations{}}
}

// byEmail is Invitations.ByEmail's body: every invitation waiting for an
// address, through the store's index.
func (self *Invitations) byEmail(ctx context.Context, email string) (r0 []link.Invite, err error) {
	found, err := InvitesStore.Find(ctx, "by-email", email)
	if err != nil {
		return nil, err
	}
	return invitesOf(found), nil
}

// change is Invitations.Change's body: the store applies the change to the
// record it holds at that instant, alone among its writers. An invitation
// that is not there is ErrNotFound; the change's own error keeps nothing.
func (self *Invitations) change(ctx context.Context, id string, change func(*link.Invite) error) (r0 link.Invite, err error) {
	rec, err := InvitesStore.Update(ctx, id, func(r *InviteRecord) error {
		v := inviteOf(*r)
		if err := change(&v); err != nil {
			return err
		}
		v.ID = r.ID // a change never renames a record
		*r = inviteRecordOf(v)
		return nil
	})
	var refused *app.Error
	switch {
	case errors.As(err, &refused) && refused.Code == app.WireNotFound:
		return link.Invite{}, link.ErrNotFound
	case err != nil:
		return link.Invite{}, err
	}
	return inviteOf(rec), nil
}

// deleteBody is Invitations.Delete's body: an invitation already gone is deleted.
func (self *Invitations) deleteBody(ctx context.Context, id string) (err error) {
	err = InvitesStore.Delete(ctx, id)
	var refused *app.Error
	if errors.As(err, &refused) && refused.Code == app.WireNotFound {
		return nil
	}
	return err
}

// get is Invitations.Get's body. It filters instead of reading the key: an
// invitation that is not there is an answer, never a failed read.
func (self *Invitations) get(ctx context.Context, id string) (r0 link.Invite, r1 bool, err error) {
	found, err := InvitesStore.Filter(ctx, func(r InviteRecord) bool { return r.ID == id })
	if err != nil || len(found) == 0 {
		return link.Invite{}, false, err
	}
	return inviteOf(found[0]), true, nil
}

// newID is Invitations.NewID's body: a time-ordered id.
func (self *Invitations) newID(ctx context.Context) (r0 string, err error) {
	return app.NewID("invite"), nil
}

// of is Invitations.Of's body: the invitations a member sent.
func (self *Invitations) of(ctx context.Context, user string) (r0 []link.Invite, err error) {
	found, err := InvitesStore.Filter(ctx, func(r InviteRecord) bool { return r.InviterID == user })
	if err != nil {
		return nil, err
	}
	return invitesOf(found), nil
}

// save is Invitations.Save's body: the invitation replaces what its id held.
func (self *Invitations) save(ctx context.Context, i link.Invite) (err error) {
	return InvitesStore.Put(ctx, inviteRecordOf(i))
}

// invitesOf is the invitations some records keep.
func invitesOf(rs []InviteRecord) []link.Invite {
	out := make([]link.Invite, len(rs))
	for i, r := range rs {
		out[i] = inviteOf(r)
	}
	return out
}

// inviteOf is the invitation a record keeps.
func inviteOf(r InviteRecord) link.Invite {
	return link.Invite{ID: r.ID, InviterID: r.InviterID, Email: r.Email, Status: r.Status, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

// inviteRecordOf is an invitation as the store keeps it.
func inviteRecordOf(i link.Invite) InviteRecord {
	return InviteRecord{ID: i.ID, InviterID: i.InviterID, Email: i.Email, Status: i.Status, CreatedAt: i.CreatedAt, UpdatedAt: i.UpdatedAt}
}
