// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"context"

	"github.com/kitsunium/todo/contacts/link"
)

// The kinds of notice a link gives.
const (
	kindRequested = "contact-requested"
	kindAccepted  = "contact-accepted"
)

// execute is RequestContact's body: the member the address belongs to is
// asked, unless they are the asker, already a contact, or a request between
// the two is waiting. A link that was declined or cancelled is replaced: a
// pair has one link at most. The addressee is then told who asks.
func (h *RequestContactHandler) execute(ctx context.Context, in *RequestContactInput) (RequestContactOutput, error) {
	to, found, err := h.people.ByEmail(ctx, in.Email)
	if err != nil {
		return RequestContactOutput{}, err
	}
	if !found {
		return RequestContactOutput{}, link.ErrUnknownMember
	}
	if to.ID == in.From {
		return RequestContactOutput{}, link.ErrSelf
	}
	pair := link.PairOf(in.From, to.ID)
	old, found, err := h.links.ByPair(ctx, pair)
	if err != nil {
		return RequestContactOutput{}, err
	}
	if found {
		if old.State == link.RequestsPending || old.State == link.RequestsAccepted {
			return RequestContactOutput{}, link.ErrAlready
		}
		if err := h.links.Delete(ctx, old.ID); err != nil {
			return RequestContactOutput{}, err
		}
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return RequestContactOutput{}, err
	}
	id, err := h.links.NewID(ctx)
	if err != nil {
		return RequestContactOutput{}, err
	}
	l := link.Link{ID: id, Pair: pair, RequesterID: in.From, AddresseeID: to.ID, State: link.RequestsInitial, CreatedAt: now, UpdatedAt: now}
	if err := h.links.Save(ctx, l); err != nil {
		return RequestContactOutput{}, err
	}
	actor, _, err := h.people.ByID(ctx, in.From)
	if err != nil {
		return RequestContactOutput{}, err
	}
	if err := h.notices.Tell(ctx, kindRequested, l, actor); err != nil {
		return RequestContactOutput{}, err
	}
	return RequestContactOutput{Link: l}, nil
}
