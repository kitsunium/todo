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
	l, found, err := h.links.Get(ctx, in.ID)
	if err != nil {
		return AnswerRequestOutput{}, err
	}
	if _, mine := l.Other(in.User); !found || !mine {
		return AnswerRequestOutput{}, link.ErrNotFound
	}
	event := link.RequestsEvent(in.Event)
	if byAddressee := event != link.RequestsOnCancel; byAddressee != (in.User == l.AddresseeID) {
		return AnswerRequestOutput{}, link.ErrNotAllowed
	}
	to, err := link.NewRequests().Fire(ctx, &l, l.State, event)
	switch {
	case errors.Is(err, link.ErrRequestsNoTransition):
		return AnswerRequestOutput{}, link.ErrNotAllowed
	case err != nil:
		return AnswerRequestOutput{}, err
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return AnswerRequestOutput{}, err
	}
	l.State, l.UpdatedAt = to, now
	if err := h.links.Save(ctx, l); err != nil {
		return AnswerRequestOutput{}, err
	}
	if to == link.RequestsAccepted {
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
