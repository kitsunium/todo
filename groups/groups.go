// Package groups is teams: a group has a name, a color, members with roles,
// and a shared task list (the tasks service keeps the tasks; a task names
// its group). Members join by invitation, and only a contact can be invited.
package groups

import (
	"cmp"
	"context"
	"slices"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/internal/wire"
)

// Service owns groups, their members and the invitations into them.
var Service = kit.NewService("groups", "Teams: groups sharing a task list, their members' roles, and the invitations that bring members in.\n\nfr: Les équipes : des groupes qui partagent une liste de tâches, les rôles de leurs membres, et les invitations qui y font entrer de nouveaux membres.")

// Role is what a member may do in a group.
type Role string

// The roles, from the most powerful. The owner created the group and alone
// deletes it or changes roles; an admin invites and removes members and
// edits the group; a member sees and works on its tasks.
const (
	Owner  Role = "owner"
	Admin  Role = "admin"
	Member Role = "member"
)

// rank orders the roles for display.
var rank = map[Role]int{Owner: 0, Admin: 1, Member: 2}

// manages reports whether a role runs the group: invites, removes, edits.
func (r Role) manages() bool { return r == Owner || r == Admin }

// Colors are the colors a group may have; the web app has a swatch for each.
var Colors = []string{"slate", "red", "orange", "amber", "green", "teal", "blue", "indigo", "violet", "pink"}

// Membership is one member of a group.
type Membership struct {
	UserID   string    `json:"userId"`
	Role     Role      `json:"role"`
	JoinedAt time.Time `json:"joinedAt"`
}

// Group is a team.
type Group struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Color     string       `json:"color"`
	Members   []Membership `json:"members"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

// GroupRef is how other services show a group.
type GroupRef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Ref returns how the group is shown elsewhere.
func (g Group) Ref() GroupRef { return GroupRef{ID: g.ID, Name: g.Name, Color: g.Color} }

// RoleOf returns the role of a user in the group, "" for a stranger.
func (g Group) RoleOf(uid string) Role {
	for _, m := range g.Members {
		if m.UserID == uid {
			return m.Role
		}
	}
	return ""
}

// memberIDs returns the members' user IDs.
func (g Group) memberIDs() []string {
	ids := make([]string, len(g.Members))
	for i, m := range g.Members {
		ids[i] = m.UserID
	}
	return ids
}

// Groups keeps every group, listed per member.
//
// fr: Groups garde chaque groupe, listé par membre.
var Groups = Service.Store("groups", func(g Group) string { return g.ID },
	kit.Index("member", func(g Group) []string { return g.memberIDs() }))

// The kinds of Event.
const (
	KindInvited = "invited"
	KindJoined  = "joined"
	KindLeft    = "left"
	KindRemoved = "removed"
	KindDeleted = "deleted"
)

// Event is what happened in a group, and whom to tell.
type Event struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	GroupID   string `json:"groupId"`
	GroupName string `json:"groupName"`
	// Color is the group's, for a notification.
	Color string `json:"color"`
	// ActorID did it: invited, joined, left, removed, deleted.
	ActorID string `json:"actorId"`
	// UserID is whom it is about: the invitee, the member who joined, left
	// or was removed.
	UserID       string `json:"userId,omitempty"`
	InvitationID string `json:"invitationId,omitempty"`
	// Recipients are the users who should hear of it, the actor excepted.
	Recipients []string  `json:"recipients"`
	At         time.Time `json:"at"`
}

// Events announces invitations and comings and goings: notify mails the
// invitee, activity writes the feeds, tasks follows its groups.
//
// fr: Events annonce les invitations, les arrivées et les départs : notify
// écrit à l’invité, activity écrit les fils d’activité, tasks suit ses groupes.
var Events = Service.Topic[Event]("events")

// announce publishes what happened in g, telling everyone in tell but the
// actor.
func announce(ctx context.Context, g Group, kind, actor, user, invitation string, tell []string) error {
	recipients := []string{}
	for _, id := range tell {
		if id != actor && id != "" && !slices.Contains(recipients, id) {
			recipients = append(recipients, id)
		}
	}
	return Events.Publish(ctx, Event{
		ID: kit.NewID("event"), Kind: kind, GroupID: g.ID, GroupName: g.Name, Color: g.Color,
		ActorID: actor, UserID: user, InvitationID: invitation, Recipients: recipients, At: wire.Now(ctx),
	})
}

// sortMembers puts the owner first, then the admins, then the members, each
// by the time they joined.
func sortMembers(ms []Membership) {
	slices.SortFunc(ms, func(a, b Membership) int {
		return cmp.Or(cmp.Compare(rank[a.Role], rank[b.Role]), a.JoinedAt.Compare(b.JoinedAt), cmp.Compare(a.UserID, b.UserID))
	})
}
