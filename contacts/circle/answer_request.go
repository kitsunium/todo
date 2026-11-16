// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"context"
	"errors"

	"github.com/kitsunium/todo/contacts/link"
)

// execute is AnswerRequest's body. A link the member is at no end of is not
// found. Its addressee accepts or declines it, its requester cancels it,
// and only while it is pending: anything else is not allowed. The requester
// is told when their request is accepted.
func (h *AnswerRequestHandler) execute(ctx context.Context, in *AnswerRequestInput) (AnswerRequestOutput, error) {
	event := link.RequestsEvent(in.Event)
	now, err := h.clock.Now(ctx)
	if err != nil {
		return AnswerRequestOutput{}, err
	}
	// The link is answered as it is at that instant: an acceptance and a
	// cancellation made at once do not both stand.
	l, err := h.links.Change(ctx, in.ID, func(l *link.Link) error {
		if _, mine := l.Other(in.User); !mine {
			return link.ErrNotFound
		}
		if byAddressee := event != link.RequestsOnCancel; byAddressee != (in.User == l.AddresseeID) {
			return link.ErrNotAllowed
		}
		to, err := link.NewRequests().Fire(ctx, l, l.State, event)
		switch {
		case errors.Is(err, link.ErrRequestsNoTransition):
			return link.ErrNotAllowed
		case err != nil:
			return err
		}
		l.State, l.UpdatedAt = to, now
		return nil
	})
	if err != nil {
		return AnswerRequestOutput{}, err
	}
	if l.State == link.RequestsAccepted {
		actor, _, err := h.people.ByID(ctx, in.User)
		if err != nil {
			return AnswerRequestOutput{}, err
		}
		if err := h.notices.Tell(ctx, kindAccepted, l, actor); err != nil {
			return AnswerRequestOutput{}, err
		}
	}
	return AnswerRequestOutput{Link: l}, nil
}
