// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package groupstore

import (
	"context"
	"errors"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/groups/team"
)

// invitations is Invitations's own state: the store is the package's, it holds nothing.
type invitations struct {
}

// newInvitations is NewInvitations's body.
func newInvitations() *Invitations {
	return &Invitations{invitations{}}
}

// deleteBody is Invitations.Delete's body: an invitation already gone is deleted.
func (a *Invitations) deleteBody(ctx context.Context, id string) (err error) {
	err = InvitationsStore.Delete(ctx, id)
	var refused *app.Error
	if errors.As(err, &refused) && refused.Code == app.WireNotFound {
		return nil
	}
	return err
}

// get is Invitations.Get's body. It filters instead of reading the key: an
// invitation that is not there is an answer, never a failed read.
func (a *Invitations) get(ctx context.Context, id string) (r0 team.Invitation, r1 bool, err error) {
	found, err := InvitationsStore.Filter(ctx, func(r InvitationRecord) bool { return r.ID == id })
	if err != nil || len(found) == 0 {
		return team.Invitation{}, false, err
	}
	return invitationOf(found[0]), true, nil
}

// newID is Invitations.NewID's body: a time-ordered id.
func (a *Invitations) newID(ctx context.Context) (r0 string, err error) {
	return app.NewID("invitation"), nil
}

// of is Invitations.Of's body: the invitations of an invitee, through the
// store's index.
func (a *Invitations) of(ctx context.Context, user string) (r0 []team.Invitation, err error) {
	return invitationsFound(InvitationsStore.Find(ctx, "by-invitee", user))
}

// ofGroup is Invitations.OfGroup's body: the invitations into a group,
// through the store's index.
func (a *Invitations) ofGroup(ctx context.Context, group string) (r0 []team.Invitation, err error) {
	return invitationsFound(InvitationsStore.Find(ctx, "by-group", group))
}

// save is Invitations.Save's body: the invitation replaces what its id held.
func (a *Invitations) save(ctx context.Context, inv team.Invitation) (err error) {
	return InvitationsStore.Put(ctx, recordOfInvitation(inv))
}

// invitationOf is the invitation a record keeps.
func invitationOf(r InvitationRecord) team.Invitation {
	return team.Invitation{
		ID: r.ID, GroupID: r.GroupID, InviterID: r.InviterID, InviteeID: r.InviteeID,
		State: team.InvitingState(r.State), CreatedAt: r.CreatedAt, RespondedAt: r.RespondedAt,
	}
}

// recordOfInvitation is an invitation as the store keeps it.
func recordOfInvitation(inv team.Invitation) InvitationRecord {
	return InvitationRecord{
		ID: inv.ID, GroupID: inv.GroupID, InviterID: inv.InviterID, InviteeID: inv.InviteeID,
		State: LapsingState(inv.State), CreatedAt: inv.CreatedAt, RespondedAt: inv.RespondedAt,
	}
}

// invitationsFound is the invitations the records of a lookup keep.
func invitationsFound(found []InvitationRecord, err error) ([]team.Invitation, error) {
	if err != nil {
		return nil, err
	}
	out := make([]team.Invitation, len(found))
	for i, r := range found {
		out[i] = invitationOf(r)
	}
	return out, nil
}

// change is Invitations.Change's body: the store applies the change to the record it
// holds at that instant, alone among its writers, so two changes made at
// once both stay. An invitation that is not there is ErrNoInvitation. The change's own error keeps nothing.
func (a *Invitations) change(ctx context.Context, id string, change func(*team.Invitation) error) (r0 team.Invitation, err error) {
	rec, err := InvitationsStore.Update(ctx, id, func(r *InvitationRecord) error {
		v := invitationOf(*r)
		if err := change(&v); err != nil {
			return err
		}
		v.ID = r.ID // a change never renames a record
		*r = recordOfInvitation(v)
		return nil
	})
	var refused *app.Error
	switch {
	case errors.As(err, &refused) && refused.Code == app.WireNotFound:
		return team.Invitation{}, team.ErrNoInvitation
	case err != nil:
		return team.Invitation{}, err
	}
	return invitationOf(rec), nil
}
