package server_test

import (
	"errors"
	"testing"
	"time"

	"github.com/kitsunium/todo/contacts/link"
	"github.com/kitsunium/todo/contacts/linkstore"
)

// The store itself keeps one waiting invitation by member and address: a
// second one written straight to it, past the use case that looks first, is
// refused by the unique index. An invitation that joined frees the place, and
// another member's, or another address, never held it.
func TestTheStoreKeepsOneWaitingInvitationByMemberAndAddress(t *testing.T) {
	start(t)
	ctx := t.Context()
	invites := linkstore.NewInvitations()
	invite := func(id, from, email, status string) error {
		return invites.Save(ctx, link.Invite{ID: id, InviterID: from, Email: email, Status: status, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	}
	if err := invite("invite_1", "user_ada", "zoe@example.com", link.InvitePending); err != nil {
		t.Fatalf("the first invitation: %v", err)
	}
	if err := invite("invite_2", "user_ada", "zoe@example.com", link.InvitePending); !errors.Is(err, link.ErrInvited) {
		t.Fatalf("a second waiting invitation of the same member to the same address: %v, want link.ErrInvited", err)
	}
	if _, found, err := invites.Get(ctx, "invite_2"); err != nil || found {
		t.Errorf("the refused invitation: found %v (%v), want it not kept", found, err)
	}
	if err := invite("invite_3", "user_bob", "zoe@example.com", link.InvitePending); err != nil {
		t.Errorf("another member's invitation to the address: %v", err)
	}
	if err := invite("invite_4", "user_ada", "kim@example.com", link.InvitePending); err != nil {
		t.Errorf("the member's invitation to another address: %v", err)
	}
	// Saved again as it is, an invitation keeps its own place.
	if err := invite("invite_1", "user_ada", "zoe@example.com", link.InvitePending); err != nil {
		t.Errorf("the first invitation saved again: %v", err)
	}
	// Once it joined, its place is free for a new one.
	if _, err := invites.Change(ctx, "invite_1", func(i *link.Invite) error {
		i.Status = link.InviteJoined
		return nil
	}); err != nil {
		t.Fatalf("joining the first invitation: %v", err)
	}
	if err := invite("invite_5", "user_ada", "zoe@example.com", link.InvitePending); err != nil {
		t.Errorf("a new invitation after the first joined: %v", err)
	}
}
