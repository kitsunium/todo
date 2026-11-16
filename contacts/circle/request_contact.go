// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"context"
	"strings"
	"unicode"

	"github.com/kitsunium/todo/contacts/link"
)

// The kinds of notice a link gives.
const (
	kindRequested = "contact-requested"
	kindAccepted  = "contact-accepted"
)

// execute is RequestContact's body: the member the address belongs to is
// asked, unless they are the asker, already a contact, or a request between
// the two is waiting. A link that was declined, cancelled or left to expire
// is replaced: a pair has one link at most. The addressee is then told who
// asks. An address nobody has is invited to sign up instead, once while the
// first invitation waits.
func (h *RequestContactHandler) execute(ctx context.Context, in *RequestContactInput) (RequestContactOutput, error) {
	email, ok := cleanAddress(in.Email)
	if !ok {
		return RequestContactOutput{}, link.ErrBadAddress
	}
	to, found, err := h.people.ByEmail(ctx, email)
	if err != nil {
		return RequestContactOutput{}, err
	}
	if !found {
		return h.invite(ctx, in.From, email)
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
	return RequestContactOutput{Kind: KindRequest, Link: l}, nil
}

// invite keeps an invitation to an address nobody has and mails it. An
// invitation that cannot be mailed is not kept: the member asks again.
func (h *RequestContactHandler) invite(ctx context.Context, from, email string) (RequestContactOutput, error) {
	mine, err := h.invites.Of(ctx, from)
	if err != nil {
		return RequestContactOutput{}, err
	}
	for _, i := range mine {
		if i.Email == email && i.Status == link.InvitePending {
			return RequestContactOutput{}, link.ErrInvited
		}
	}
	now, err := h.clock.Now(ctx)
	if err != nil {
		return RequestContactOutput{}, err
	}
	id, err := h.invites.NewID(ctx)
	if err != nil {
		return RequestContactOutput{}, err
	}
	i := link.Invite{ID: id, InviterID: from, Email: email, Status: link.InvitePending, CreatedAt: now, UpdatedAt: now}
	if err := h.invites.Save(ctx, i); err != nil {
		return RequestContactOutput{}, err
	}
	actor, _, err := h.people.ByID(ctx, from)
	if err != nil {
		return RequestContactOutput{}, err
	}
	if err := h.inviter.Invite(ctx, email, actor); err != nil {
		if gone := h.invites.Delete(ctx, i.ID); gone != nil {
			return RequestContactOutput{}, gone
		}
		return RequestContactOutput{}, err
	}
	return RequestContactOutput{Kind: KindInvite, Invitation: invitationOf(i)}, nil
}

// cleanAddress is an address as the product keeps it, trimmed and in lower
// case, and whether it is one: a local part, one @ and a domain, 254
// characters at most, with no space or control character. It is the rule an
// account's address follows, so that every address a member may have can be
// asked, and no other invited.
func cleanAddress(s string) (string, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	local, domain, ok := strings.Cut(s, "@")
	if !ok || local == "" || domain == "" || len(s) > 254 || strings.Contains(domain, "@") {
		return "", false
	}
	if strings.IndexFunc(s, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0 {
		return "", false
	}
	return s, true
}

// invitationOf is an invitation as its inviter reads it.
func invitationOf(i link.Invite) Invitation {
	return Invitation{ID: i.ID, Email: i.Email, CreatedAt: i.CreatedAt}
}
