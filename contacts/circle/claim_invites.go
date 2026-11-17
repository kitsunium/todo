// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"context"
	"errors"
	"time"

	"github.com/kitsunium/todo/contacts/link"
)

// execute is ClaimInvites's body: the member's address is read from the
// directory, as the announcement carries the id only; a member the directory
// does not know has nothing to claim. Each invitation waiting for the address
// becomes a request from its inviter to the member, told in the member's
// feed as any request is, then is marked joined. A pair already linked is
// left alone, and so is an inviter who is the member. The request is made
// before the mark, and an invitation already joined is skipped: a message
// delivered twice changes nothing the second time.
func (h *ClaimInvitesHandler) execute(ctx context.Context, in *ClaimInvitesInput) (ClaimInvitesOutput, error) {
	member, known, err := h.people.ByID(ctx, in.UserID)
	if err != nil || !known {
		return ClaimInvitesOutput{}, err
	}
	waiting, err := h.invites.ByEmail(ctx, member.Email)
	if err != nil {
		return ClaimInvitesOutput{}, err
	}
	for _, inv := range waiting {
		if inv.Status != link.InvitePending {
			continue
		}
		now, err := h.clock.Now(ctx)
		if err != nil {
			return ClaimInvitesOutput{}, err
		}
		if inv.InviterID != in.UserID {
			if err := h.request(ctx, inv.InviterID, in.UserID, now); err != nil {
				return ClaimInvitesOutput{}, err
			}
		}
		_, err = h.invites.Change(ctx, inv.ID, func(i *link.Invite) error {
			i.Status, i.UpdatedAt = link.InviteJoined, now
			return nil
		})
		if err != nil && !errors.Is(err, link.ErrNotFound) { // withdrawn meanwhile
			return ClaimInvitesOutput{}, err
		}
	}
	return ClaimInvitesOutput{}, nil
}

// request asks the member to be the inviter's contact, as RequestContact
// does for an address that has an account, and says nothing of a pair that
// is linked or waiting.
func (h *ClaimInvitesHandler) request(ctx context.Context, from, to string, now time.Time) error {
	actor, known, err := h.people.ByID(ctx, from)
	if err != nil || !known { // an inviter whose account is gone asks nobody
		return err
	}
	pair := link.PairOf(from, to)
	old, found, err := h.links.ByPair(ctx, pair)
	if err != nil {
		return err
	}
	if found {
		if old.State == link.RequestsPending || old.State == link.RequestsAccepted {
			return nil
		}
		if err := h.links.Delete(ctx, old.ID); err != nil {
			return err
		}
	}
	id, err := h.links.NewID(ctx)
	if err != nil {
		return err
	}
	l := link.Link{ID: id, Pair: pair, RequesterID: from, AddresseeID: to, State: link.RequestsInitial, CreatedAt: now, UpdatedAt: now}
	if err := h.links.Save(ctx, l); err != nil {
		return err
	}
	return h.notices.Tell(ctx, kindRequested, l, actor)
}
