// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"cmp"
	"context"
	"slices"
	"strings"

	"github.com/kitsunium/todo/contacts/link"
)

// execute is ListContacts's body: the member's accepted links, the pending
// ones they received and the pending ones they sent, each with the member
// at its other end, by name. A link whose other member is gone is left out.
// The invitations still waiting follow, the newest first.
func (h *ListContactsHandler) execute(ctx context.Context, in *ListContactsInput) (ListContactsOutput, error) {
	links, err := h.links.Of(ctx, in.User)
	if err != nil {
		return ListContactsOutput{}, err
	}
	out := ListContactsOutput{Contacts: []Contact{}, Incoming: []Contact{}, Outgoing: []Contact{}, Invites: []Invitation{}}
	for _, l := range links {
		other, mine := l.Other(in.User)
		if !mine {
			continue
		}
		who, found, err := h.people.ByID(ctx, other)
		if err != nil {
			return ListContactsOutput{}, err
		}
		if !found {
			continue
		}
		c := Contact{ID: l.ID, User: who, Since: l.UpdatedAt}
		switch {
		case l.State == link.RequestsAccepted:
			out.Contacts = append(out.Contacts, c)
		case l.State == link.RequestsPending && l.AddresseeID == in.User:
			out.Incoming = append(out.Incoming, c)
		case l.State == link.RequestsPending:
			out.Outgoing = append(out.Outgoing, c)
		}
	}
	for _, list := range [][]Contact{out.Contacts, out.Incoming, out.Outgoing} {
		slices.SortFunc(list, func(a, b Contact) int {
			return cmp.Or(cmp.Compare(strings.ToLower(a.User.Name), strings.ToLower(b.User.Name)), cmp.Compare(a.ID, b.ID))
		})
	}
	sent, err := h.invites.Of(ctx, in.User)
	if err != nil {
		return ListContactsOutput{}, err
	}
	for _, i := range sent {
		if i.Status == link.InvitePending {
			out.Invites = append(out.Invites, invitationOf(i))
		}
	}
	slices.SortFunc(out.Invites, func(a, b Invitation) int {
		return cmp.Or(b.CreatedAt.Compare(a.CreatedAt), cmp.Compare(b.ID, a.ID))
	})
	return out, nil
}
