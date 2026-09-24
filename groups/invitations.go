package groups

import (
	"context"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/internal/wire"
)

// InvitationStatus is where an invitation into a group stands.
type InvitationStatus string

// The states of an invitation.
const (
	Pending  InvitationStatus = "pending"
	Accepted InvitationStatus = "accepted"
	Declined InvitationStatus = "declined"
	Revoked  InvitationStatus = "revoked"
	Expired  InvitationStatus = "expired"
)

// Invitation asks a user to join a group.
type Invitation struct {
	ID          string           `json:"id"`
	GroupID     string           `json:"groupId"`
	InviterID   string           `json:"inviterId"`
	InviteeID   string           `json:"inviteeId"`
	Status      InvitationStatus `json:"status"`
	CreatedAt   time.Time        `json:"createdAt"`
	RespondedAt *time.Time       `json:"respondedAt,omitempty"`
}

// Invitations keeps the invitations, listed per invitee and per group.
//
// fr: Invitations garde les invitations, listées par invité et par groupe.
var Invitations = Service.Store("invitations", func(i Invitation) string { return i.ID },
	kit.Index("invitee", func(i Invitation) []string { return []string{i.InviteeID} }),
	kit.Index("group", func(i Invitation) []string { return []string{i.GroupID} }))

// InvitationsExpireAfter is how long an invitation waits for an answer.
const InvitationsExpireAfter = 7 * 24 * time.Hour

// InvitationLifecycle is the life of an invitation: the invitee accepts —
// and joins the group — or declines; deleting the group revokes it; a week
// unanswered expires it.
//
// fr: InvitationLifecycle est la vie d’une invitation : l’invité l’accepte — et
// rejoint le groupe — ou la refuse ; supprimer le groupe la révoque ; une
// semaine sans réponse la fait expirer.
var InvitationLifecycle = Service.Workflow("invitations", Invitations, func(i *Invitation) *InvitationStatus { return &i.Status }).
	Initial(Pending).
	On("accept", Pending, Accepted).
	On("decline", Pending, Declined).
	On("revoke", Pending, Revoked).
	After("expire", InvitationsExpireAfter, Pending, Expired).
	OnEnter(Accepted, join).
	OnEnter(Declined, stampAnswer).
	OnTransition(announceInvitation)

// join adds the invitee to the group, as a member, as the invitation is
// accepted. A group deleted meanwhile fails the acceptance.
//
// fr: join ajoute l’invité au groupe, comme membre, au moment où l’invitation
// est acceptée. Un groupe supprimé entre-temps fait échouer l’acceptation.
func join(ctx context.Context, inv *Invitation) error {
	now := wire.Now(ctx)
	inv.RespondedAt = &now
	_, err := Groups.Update(ctx, inv.GroupID, func(g *Group) error {
		if g.RoleOf(inv.InviteeID) == "" {
			g.Members = append(g.Members, Membership{UserID: inv.InviteeID, Role: Member, JoinedAt: now})
			g.UpdatedAt = now
		}
		return nil
	})
	if wire.Is(err, kit.CodeNotFound) {
		return kit.NotFound("this group no longer exists")
	}
	return err
}

// stampAnswer records when an invitation was answered.
//
// fr: stampAnswer note quand une invitation a reçu sa réponse.
func stampAnswer(ctx context.Context, inv *Invitation) error {
	now := wire.Now(ctx)
	inv.RespondedAt = &now
	return nil
}

// announceInvitation tells the invitee they are invited, and the members
// that someone joined.
//
// fr: announceInvitation annonce à l’invité qu’il est invité, et aux membres
// que quelqu’un les a rejoints.
func announceInvitation(ctx context.Context, c kit.Change[Invitation, InvitationStatus]) error {
	inv := c.Entity
	if c.To != Pending && c.To != Accepted {
		return nil
	}
	g, err := Groups.Get(ctx, inv.GroupID)
	if wire.Is(err, kit.CodeNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if c.To == Pending {
		return announce(ctx, g, KindInvited, inv.InviterID, inv.InviteeID, inv.ID, []string{inv.InviteeID})
	}
	return announce(ctx, g, KindJoined, inv.InviteeID, inv.InviteeID, inv.ID, g.memberIDs())
}
