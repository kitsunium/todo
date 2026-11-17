// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package reminders

import (
	"context"
)

// execute is MailAction's body: the member an event is about is mailed of
// what the actor did, once for each event and member. What the product
// does not mail (another kind of event, a member who did it to themselves),
// an event already mailed, and an actor or a member the directory does not
// know are all a success that sends nothing. The mail leaves before the
// event is marked, so one that fails is tried again at the next delivery.
func (h *MailActionHandler) execute(ctx context.Context, in *MailActionInput) (MailActionOutput, error) {
	switch in.Kind {
	case OnContactRequested, OnContactAccepted, OnGroupInvited, OnShared, OnAssigned:
	default:
		return MailActionOutput{}, nil
	}
	if in.UserID == "" || in.UserID == in.ActorID {
		return MailActionOutput{}, nil
	}
	key := in.EventID + "." + in.UserID
	seen, err := h.actionLog.Seen(ctx, key)
	if err != nil || seen {
		return MailActionOutput{}, err
	}
	actor, found, err := h.directory.Recipient(ctx, in.ActorID)
	if err != nil || !found {
		return MailActionOutput{}, err
	}
	to, found, err := h.directory.Recipient(ctx, in.UserID)
	if err != nil || !found {
		return MailActionOutput{}, err
	}
	n := Notice{To: to, Kind: in.Kind, Actor: actor.Name, Subject: in.Subject, SubjectID: in.SubjectID}
	if err := h.notifier.Notify(ctx, n); err != nil {
		return MailActionOutput{}, err
	}
	if err := h.actionLog.Mark(ctx, key); err != nil {
		return MailActionOutput{}, err
	}
	return MailActionOutput{Sent: 1}, nil
}
