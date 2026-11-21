// The ports the use cases' tests run on: groups, invitations and members in
// memory, a clock that stands still, notices that are remembered.

package roster

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/kitsunium/todo/groups/team"
)

// epoch is the instant every test starts at.
var epoch = time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)

// world is every port of the domain, in memory, with five members: ada, bob,
// cy, dee and eve.
type world struct {
	groups      map[string]team.Group
	invitations map[string]team.Invitation
	people      map[string]Person
	contacts    map[string]bool
	notices     []string
	n           int
}

func newWorld() *world {
	w := &world{
		groups: map[string]team.Group{}, invitations: map[string]team.Invitation{},
		people: map[string]Person{}, contacts: map[string]bool{},
	}
	for _, name := range []string{"ada", "bob", "cy", "dee", "eve"} {
		w.people["user_"+name] = Person{ID: "user_" + name, Name: name}
	}
	return w
}

// core is a group named Core where ada is the owner, bob an admin and cy a
// member; dee and eve are strangers to it.
func (w *world) core() team.Group {
	w.n++
	g := team.Group{
		ID: "group_" + strconv.Itoa(w.n), Name: "Core", Color: "teal", CreatedAt: epoch, UpdatedAt: epoch,
		Members: []team.Membership{
			{UserID: "user_cy", Role: team.RoleMember, JoinedAt: epoch.Add(2 * time.Minute)},
			{UserID: "user_bob", Role: team.RoleAdmin, JoinedAt: epoch.Add(time.Minute)},
			{UserID: "user_ada", Role: team.RoleOwner, JoinedAt: epoch},
		},
	}
	w.groups[g.ID] = g
	return g
}

// invited adds an invitation into a group, in a state.
func (w *world) invited(g team.Group, inviter, invitee string, state team.InvitingState) team.Invitation {
	w.n++
	inv := team.Invitation{
		ID: "invitation_" + strconv.Itoa(w.n), GroupID: g.ID, InviterID: inviter, InviteeID: invitee, State: state,
		CreatedAt: epoch.Add(time.Duration(w.n) * time.Second),
	}
	w.invitations[inv.ID] = inv
	return inv
}

// befriend makes two members contacts.
func (w *world) befriend(one, other string) { w.contacts[one+"|"+other] = true }

// role is what the store holds of a member's role in a group.
func (w *world) role(group, user string) team.Role {
	g := w.groups[group]
	return g.RoleOf(user)
}

func (w *world) NewID(context.Context) (string, error) {
	w.n++
	return "id_" + strconv.Itoa(w.n), nil
}

func (w *world) Now(context.Context) (time.Time, error) { return epoch.Add(time.Hour), nil }

// groupsPort is the world as Groups@v1: the two stores share method names,
// so each is a view of the world of its own.
type groupsPort struct{ w *world }

func (p groupsPort) NewID(ctx context.Context) (string, error) {
	id, err := p.w.NewID(ctx)
	return "group_" + strings.TrimPrefix(id, "id_"), err
}

func (p groupsPort) Save(_ context.Context, g team.Group) error {
	p.w.groups[g.ID] = g
	return nil
}

func (p groupsPort) Get(_ context.Context, id string) (team.Group, bool, error) {
	g, ok := p.w.groups[id]
	return g, ok, nil
}

func (p groupsPort) Of(_ context.Context, user string) ([]team.Group, error) {
	var out []team.Group
	for _, g := range p.w.groups {
		if g.RoleOf(user) != "" {
			out = append(out, g)
		}
	}
	return out, nil
}

func (p groupsPort) Change(_ context.Context, id string, change func(*team.Group) error) (team.Group, error) {
	g, ok := p.w.groups[id]
	if !ok {
		return team.Group{}, team.ErrNotFound
	}
	if err := change(&g); err != nil {
		return team.Group{}, err
	}
	p.w.groups[id] = g
	return g, nil
}

func (p groupsPort) Delete(_ context.Context, id string) error {
	delete(p.w.groups, id)
	return nil
}

// invitationsPort is the world as Invitations@v1.
type invitationsPort struct{ w *world }

func (p invitationsPort) NewID(ctx context.Context) (string, error) {
	id, err := p.w.NewID(ctx)
	return "invitation_" + strings.TrimPrefix(id, "id_"), err
}

func (p invitationsPort) Save(_ context.Context, inv team.Invitation) error {
	p.w.invitations[inv.ID] = inv
	return nil
}

func (p invitationsPort) Get(_ context.Context, id string) (team.Invitation, bool, error) {
	inv, ok := p.w.invitations[id]
	return inv, ok, nil
}

func (p invitationsPort) Of(_ context.Context, user string) ([]team.Invitation, error) {
	var out []team.Invitation
	for _, inv := range p.w.invitations {
		if inv.InviteeID == user {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (p invitationsPort) OfGroup(_ context.Context, group string) ([]team.Invitation, error) {
	var out []team.Invitation
	for _, inv := range p.w.invitations {
		if inv.GroupID == group {
			out = append(out, inv)
		}
	}
	return out, nil
}

func (p invitationsPort) Change(_ context.Context, id string, change func(*team.Invitation) error) (team.Invitation, error) {
	inv, ok := p.w.invitations[id]
	if !ok {
		return team.Invitation{}, team.ErrNoInvitation
	}
	if err := change(&inv); err != nil {
		return team.Invitation{}, err
	}
	p.w.invitations[id] = inv
	return inv, nil
}

func (p invitationsPort) Delete(_ context.Context, id string) error {
	delete(p.w.invitations, id)
	return nil
}

func (w *world) ByID(_ context.Context, id string) (Person, bool, error) {
	p, ok := w.people[id]
	return p, ok, nil
}

func (w *world) AreContacts(_ context.Context, one, other string) (bool, error) {
	return w.contacts[one+"|"+other] || w.contacts[other+"|"+one], nil
}

// Tell remembers a notice as "kind by actor on group to who, who".
func (w *world) Tell(_ context.Context, kind string, g team.Group, actor Person, to []string) error {
	w.notices = append(w.notices, kind+" by "+actor.ID+" on "+g.ID+" to "+strings.Join(to, ","))
	return nil
}

func (w *world) groupsPort() groupsPort           { return groupsPort{w} }
func (w *world) invitationsPort() invitationsPort { return invitationsPort{w} }

func (w *world) create() CreateGroup { return NewCreateGroup(w, w.groupsPort(), w) }
func (w *world) edit() EditGroup     { return NewEditGroup(w, w.groupsPort(), w) }
func (w *world) remove() DeleteGroup {
	return NewDeleteGroup(w.groupsPort(), w.invitationsPort(), w, w)
}
func (w *world) get() GetGroup { return NewGetGroup(w.groupsPort(), w) }
func (w *world) list() ListGroups {
	return NewListGroups(w.groupsPort(), w)
}
func (w *world) invite() Invite {
	return NewInvite(w.groupsPort(), w, w.invitationsPort(), w, w, w)
}
func (w *world) answer() AnswerInvitation {
	return NewAnswerInvitation(w, w.invitationsPort(), w.groupsPort(), w, w)
}
func (w *world) invitationsList() ListInvitations {
	return NewListInvitations(w.invitationsPort(), w.groupsPort(), w)
}
func (w *world) groupInvitations() ListGroupInvitations {
	return NewListGroupInvitations(w.groupsPort(), w.invitationsPort(), w)
}
func (w *world) withdraw() WithdrawInvitation {
	return NewWithdrawInvitation(w.groupsPort(), w.invitationsPort())
}
func (w *world) leave() RemoveMember { return NewRemoveMember(w, w.groupsPort(), w, w) }
func (w *world) setRole() SetRole    { return NewSetRole(w, w.groupsPort(), w, w) }
func (w *world) roleOf() RoleOf      { return NewRoleOf(w.groupsPort()) }
func (w *world) memberships() Memberships {
	return NewMemberships(w.groupsPort())
}
