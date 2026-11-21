// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"errors"
	"testing"
	"time"

	"github.com/kitsunium/todo/contacts/link"
)

// givenRequestContactAsksAMember builds RequestContact for "asks a member": the use case, its input, and
// what its output must be.
func givenRequestContactAsksAMember(t *testing.T) (RequestContact, RequestContactInput, func(*testing.T, RequestContactOutput)) {
	t.Helper()
	w := newWorld()
	return w.request(), RequestContactInput{From: "user_ada", Email: "bob@example.com"}, func(t *testing.T, out RequestContactOutput) {
		t.Helper()
		l := out.Link
		if l.RequesterID != "user_ada" || l.AddresseeID != "user_bob" || l.State != link.RequestsPending || l.Pair != link.PairOf("user_bob", "user_ada") {
			t.Errorf("the request: %+v", l)
		}
		if kept := w.links[l.ID]; kept != l {
			t.Errorf("the store keeps %+v, want %+v", kept, l)
		}
		if len(w.notices) != 1 || w.notices[0] != "contact-requested by user_ada on "+l.ID {
			t.Errorf("the notices: %q", w.notices)
		}
	}
}

// givenRequestContactAsksAgainAfterARefusal builds RequestContact for "asks again after a refusal": the use case, its input, and
// what its output must be.
func givenRequestContactAsksAgainAfterARefusal(t *testing.T) (RequestContact, RequestContactInput, func(*testing.T, RequestContactOutput)) {
	t.Helper()
	w := newWorld()
	old := w.linked("user_bob", "user_ada", link.RequestsDeclined)
	return w.request(), RequestContactInput{From: "user_ada", Email: "bob@example.com"}, func(t *testing.T, out RequestContactOutput) {
		t.Helper()
		if _, still := w.links[old.ID]; still || len(w.links) != 1 || out.Link.State != link.RequestsPending || out.Link.RequesterID != "user_ada" {
			t.Errorf("the pair's links: %+v, want the new request alone", w.links)
		}
	}
}

// givenRequestContactInvitesAnAddressNobodyHas builds RequestContact for "invites an address nobody has": the use case, its input, and
// what its output must be.
func givenRequestContactInvitesAnAddressNobodyHas(t *testing.T) (RequestContact, RequestContactInput, func(*testing.T, RequestContactOutput)) {
	t.Helper()
	w := newWorld()
	return w.request(), RequestContactInput{From: "user_ada", Email: " Zoe@Example.com "}, func(t *testing.T, out RequestContactOutput) {
		t.Helper()
		i := out.Invitation
		if out.Kind != KindInvite || i.Email != "zoe@example.com" || i.ID == "" || !i.CreatedAt.Equal(epoch.Add(time.Hour)) {
			t.Errorf("the answer: %+v", out)
		}
		kept := w.invites[i.ID]
		if kept.InviterID != "user_ada" || kept.Email != "zoe@example.com" || kept.Status != link.InvitePending || len(w.invites) != 1 {
			t.Errorf("the store keeps %+v", w.invites)
		}
		if len(w.mailed) != 1 || w.mailed[0] != "zoe@example.com by user_ada" || len(w.links) != 0 || len(w.notices) != 0 {
			t.Errorf("mailed %q, links %+v, notices %q", w.mailed, w.links, w.notices)
		}
	}
}

// givenRequestContactInvitesAgainAfterACancellation builds RequestContact for "invites again after a cancellation": the use case, its input, and
// what its output must be.
func givenRequestContactInvitesAgainAfterACancellation(t *testing.T) (RequestContact, RequestContactInput, func(*testing.T, RequestContactOutput)) {
	t.Helper()
	w := newWorld()
	// Bob's invitation to the same address, and Ada's own that joined long ago, bar nothing.
	w.invited("user_bob", "zoe@example.com")
	joined := w.invited("user_ada", "zoe@example.com")
	joined.Status = link.InviteJoined
	w.invites[joined.ID] = joined
	return w.request(), RequestContactInput{From: "user_ada", Email: "zoe@example.com"}, func(t *testing.T, out RequestContactOutput) {
		t.Helper()
		if out.Kind != KindInvite || len(w.invites) != 3 || len(w.mailed) != 1 {
			t.Errorf("the answer %+v, the invitations %+v, mailed %q", out, w.invites, w.mailed)
		}
	}
}

// givenRequestContactRefusesAnAddressThatIsNone builds RequestContact for "refuses an address that is none": the use case and the input it
// refuses with link.ErrBadAddress.
func givenRequestContactRefusesAnAddressThatIsNone(t *testing.T) (RequestContact, RequestContactInput) {
	t.Helper()
	return newWorld().request(), RequestContactInput{From: "user_ada", Email: "not an address"}
}

// givenRequestContactRefusesAnAddressAlreadyInvited builds RequestContact for "refuses an address already invited": the use case and the input it
// refuses with link.ErrInvited.
func givenRequestContactRefusesAnAddressAlreadyInvited(t *testing.T) (RequestContact, RequestContactInput) {
	t.Helper()
	w := newWorld()
	w.invited("user_ada", "zoe@example.com")
	return w.request(), RequestContactInput{From: "user_ada", Email: "zoe@example.com"}
}

// givenRequestContactRefusesOneself builds RequestContact for "refuses oneself": the use case and the input it
// refuses with link.ErrSelf.
func givenRequestContactRefusesOneself(t *testing.T) (RequestContact, RequestContactInput) {
	t.Helper()
	return newWorld().request(), RequestContactInput{From: "user_ada", Email: "ada@example.com"}
}

// givenRequestContactRefusesAPairAlreadyLinked builds RequestContact for "refuses a pair already linked": the use case and the input it
// refuses with link.ErrAlready.
func givenRequestContactRefusesAPairAlreadyLinked(t *testing.T) (RequestContact, RequestContactInput) {
	t.Helper()
	w := newWorld()
	w.linked("user_bob", "user_ada", link.RequestsPending)
	return w.request(), RequestContactInput{From: "user_ada", Email: "bob@example.com"}
}

// An invitation that cannot be mailed is not kept, so the member asks again
// and is not told they already did.
func TestRequestContactKeepsNoInvitationItCouldNotMail(t *testing.T) {
	w := newWorld()
	w.mailErr = errors.New("relay down")
	if _, err := w.request().Execute(t.Context(), RequestContactInput{From: "user_ada", Email: "zoe@example.com"}); !errors.Is(err, w.mailErr) {
		t.Fatalf("error %v, want the mailer's", err)
	}
	if len(w.invites) != 0 {
		t.Errorf("the invitations: %+v, want none", w.invites)
	}
}
