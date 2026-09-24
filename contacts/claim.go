package contacts

import (
	"context"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/identity"
	"github.com/kitsunium/todo/internal/wire"
)

// The invitations by email wait for identity to say their address now has a
// verified account.
var _ = Service.Subscribe("claim-invites", identity.AccountEvents, ClaimInvites)

// ClaimInvites turns the invitations waiting for a newly verified address
// into contact requests from whoever sent them. Delivery is at least once:
// an invitation already claimed is skipped, and a request that already
// exists is left as it is.
func ClaimInvites(ctx context.Context, e identity.AccountEvent) error {
	if e.Kind != identity.Verified {
		return nil
	}
	invites, err := Invites.Find(ctx, "email", e.Email)
	if err != nil {
		return err
	}
	for _, inv := range invites {
		if inv.Status != Invited {
			continue
		}
		if inv.InviterID != e.UserID {
			if _, err := request(ctx, inv.InviterID, e.UserID); err != nil && !wire.Is(err, kit.CodeConflict) {
				return err
			}
		}
		now := wire.Now(ctx)
		if _, err := Invites.Update(ctx, inv.ID, func(i *Invite) error {
			i.Status, i.JoinedAt = Joined, &now
			return nil
		}); err != nil && !wire.Is(err, kit.CodeNotFound) {
			return err
		}
	}
	return nil
}
