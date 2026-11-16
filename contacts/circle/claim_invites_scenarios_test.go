// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"context"
	"testing"
	"time"

	"github.com/kitsunium/todo/contacts/link"
)

// zoe is the member whose address was just proven.
var zoe = ClaimInvitesInput{UserID: "user_zoe", Email: "zoe@example.com"}

// givenClaimInvitesTurnsInvitationsIntoRequests builds ClaimInvites for "turns invitations into requests": the use case, its input, and
// what its output must be.
func givenClaimInvitesTurnsInvitationsIntoRequests(t *testing.T) (ClaimInvites, ClaimInvitesInput, func(*testing.T, ClaimInvitesOutput)) {
	t.Helper()
	w := newWorld()
	ada, bob := w.invited("user_ada", "zoe@example.com"), w.invited("user_bob", "zoe@example.com")
	other := w.invited("user_cy", "kim@example.com") // for another address
	return w.claim(), zoe, func(t *testing.T, _ ClaimInvitesOutput) {
		t.Helper()
		for from, inv := range map[string]link.Invite{"user_ada": ada, "user_bob": bob} {
			l, found, _ := w.ByPair(context.Background(), link.PairOf(from, "user_zoe"))
			if !found || l.RequesterID != from || l.AddresseeID != "user_zoe" || l.State != link.RequestsPending || !l.CreatedAt.Equal(epoch.Add(time.Hour)) {
				t.Errorf("the request of %s: %+v", from, l)
			}
			if got := w.invites[inv.ID]; got.Status != link.InviteJoined {
				t.Errorf("the invitation of %s: %+v, want joined", from, got)
			}
		}
		if len(w.links) != 2 || len(w.notices) != 2 {
			t.Errorf("links %+v, notices %q, want two of each", w.links, w.notices)
		}
		if w.invites[other.ID].Status != link.InvitePending {
			t.Errorf("the invitation to another address is %+v, want it left alone", w.invites[other.ID])
		}
	}
}

// givenClaimInvitesLeavesAPairAlreadyLinkedAlone builds ClaimInvites for "leaves a pair already linked alone": the use case, its input, and
// what its output must be.
func givenClaimInvitesLeavesAPairAlreadyLinkedAlone(t *testing.T) (ClaimInvites, ClaimInvitesInput, func(*testing.T, ClaimInvitesOutput)) {
	t.Helper()
	w := newWorld()
	w.people["user_zoe"] = Person{ID: "user_zoe", Email: "zoe@example.com", Name: "Zoe"}
	linked := w.linked("user_ada", "user_zoe", link.RequestsAccepted)
	inv := w.invited("user_ada", "zoe@example.com")
	return w.claim(), zoe, func(t *testing.T, _ ClaimInvitesOutput) {
		t.Helper()
		if kept := w.links[linked.ID]; len(w.links) != 1 || kept != linked || len(w.notices) != 0 {
			t.Errorf("links %+v, notices %q, want the accepted link alone", w.links, w.notices)
		}
		if w.invites[inv.ID].Status != link.InviteJoined {
			t.Errorf("the invitation: %+v, want joined all the same", w.invites[inv.ID])
		}
	}
}

// givenClaimInvitesSkipsAnInviterWhoIsTheMember builds ClaimInvites for "skips an inviter who is the member": the use case, its input, and
// what its output must be.
func givenClaimInvitesSkipsAnInviterWhoIsTheMember(t *testing.T) (ClaimInvites, ClaimInvitesInput, func(*testing.T, ClaimInvitesOutput)) {
	t.Helper()
	w := newWorld()
	inv := w.invited("user_zoe", "zoe@example.com")
	return w.claim(), zoe, func(t *testing.T, _ ClaimInvitesOutput) {
		t.Helper()
		if len(w.links) != 0 || len(w.notices) != 0 || w.invites[inv.ID].Status != link.InviteJoined {
			t.Errorf("links %+v, invitation %+v, want no request and the invitation joined", w.links, w.invites[inv.ID])
		}
	}
}

// givenClaimInvitesClaimsTwiceAsOnce builds ClaimInvites for "claims twice as once": the use case, its input, and
// what its output must be.
func givenClaimInvitesClaimsTwiceAsOnce(t *testing.T) (ClaimInvites, ClaimInvitesInput, func(*testing.T, ClaimInvitesOutput)) {
	t.Helper()
	w := newWorld()
	w.invited("user_ada", "zoe@example.com")
	uc := w.claim()
	return uc, zoe, func(t *testing.T, _ ClaimInvitesOutput) {
		t.Helper()
		before := len(w.links)
		for _, l := range w.links { // the request is answered between the two deliveries
			l.State = link.RequestsDeclined
			w.links[l.ID] = l
		}
		if _, err := uc.Execute(t.Context(), zoe); err != nil {
			t.Fatal(err)
		}
		if len(w.links) != before || before != 1 || len(w.notices) != 1 {
			t.Errorf("links %+v, notices %q, want the one request and its notice", w.links, w.notices)
		}
		for _, l := range w.links {
			if l.State != link.RequestsDeclined {
				t.Errorf("the second delivery changed the link: %+v", l)
			}
		}
	}
}

// givenClaimInvitesIgnoresAnAddressNobodyInvited builds ClaimInvites for "ignores an address nobody invited": the use case, its input, and
// what its output must be.
func givenClaimInvitesIgnoresAnAddressNobodyInvited(t *testing.T) (ClaimInvites, ClaimInvitesInput, func(*testing.T, ClaimInvitesOutput)) {
	t.Helper()
	w := newWorld()
	w.invited("user_ada", "kim@example.com")
	return w.claim(), zoe, func(t *testing.T, _ ClaimInvitesOutput) {
		t.Helper()
		if len(w.links) != 0 || len(w.notices) != 0 {
			t.Errorf("links %+v, notices %q, want nothing", w.links, w.notices)
		}
	}
}
