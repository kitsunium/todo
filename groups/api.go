package groups

import (
	"cmp"
	"context"
	"hash/fnv"
	"slices"
	"strings"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/contacts"
	"github.com/kitsunium/todo/identity"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/tasks"
)

// The groups API. A group a user is not a member of does not exist for them:
// 404, never 403. A member who may not do something is told so: 403.
var (
	_ = Service.Endpoint("GET /api/groups", List, kit.Auth())
	_ = Service.Endpoint("POST /api/groups", Create, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("GET /api/groups/{id}", Get, kit.Auth())
	_ = Service.Endpoint("PATCH /api/groups/{id}", Update, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("DELETE /api/groups/{id}", Delete, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("POST /api/groups/{id}/invitations", Invite, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("GET /api/invitations", ListInvitations, kit.Auth())
	_ = Service.Endpoint("POST /api/invitations/{id}/accept", AcceptInvitation, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("POST /api/invitations/{id}/decline", DeclineInvitation, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("DELETE /api/groups/{id}/members/{userId}", RemoveMember, kit.Auth(), kit.RateLimitPerClient(10, 20))
	_ = Service.Endpoint("PATCH /api/groups/{id}/members/{userId}", SetRole, kit.Auth(), kit.RateLimitPerClient(10, 20))
)

// MemberView is a member, as the group's page shows them.
type MemberView struct {
	User     identity.UserRef `json:"user"`
	Role     Role             `json:"role"`
	JoinedAt time.Time        `json:"joinedAt"`
}

// View is a group as one of its members sees it.
type View struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	// Role is the caller's.
	Role      Role         `json:"role"`
	Members   []MemberView `json:"members"`
	CreatedAt time.Time    `json:"createdAt"`
	// OpenTasks counts the tasks left to do on the group's list.
	OpenTasks int `json:"openTasks"`
}

// ListOutput is the caller's groups.
type ListOutput struct {
	Groups []View `json:"groups"`
}

// me is the signed-in user.
func me(ctx context.Context) (string, error) {
	uid, ok := kit.UserID(ctx)
	if !ok {
		return "", kit.Unauthenticated("sign in first")
	}
	return string(uid), nil
}

// errNoGroup answers for a group that does not exist, or not for the
// caller.
func errNoGroup() error { return kit.NotFound("no such group") }

// membership returns a group the user is a member of, and their role.
func membership(ctx context.Context, id, uid string) (Group, Role, error) {
	g, err := Groups.Get(ctx, id)
	if wire.Is(err, kit.CodeNotFound) {
		return Group{}, "", errNoGroup()
	}
	if err != nil {
		return Group{}, "", err
	}
	role := g.RoleOf(uid)
	if role == "" {
		return Group{}, "", errNoGroup()
	}
	return g, role, nil
}

// present turns groups into what uid sees of them: the members' names from
// identity, the open tasks from tasks.
func present(ctx context.Context, uid string, gs ...Group) ([]View, error) {
	var users, ids []string
	for _, g := range gs {
		ids = append(ids, g.ID)
		users = append(users, g.memberIDs()...)
	}
	names, err := identity.Directory(ctx, users...)
	if err != nil {
		return nil, err
	}
	open := tasks.OpenByGroupOutput{Counts: map[string]int{}}
	if len(ids) > 0 {
		if open, err = tasks.OpenByGroupAPI.Call(ctx, tasks.GroupIDs{IDs: ids}); err != nil {
			return nil, err
		}
	}
	out := make([]View, 0, len(gs))
	for _, g := range gs {
		members := slices.Clone(g.Members)
		sortMembers(members)
		v := View{ID: g.ID, Name: g.Name, Color: g.Color, Role: g.RoleOf(uid), CreatedAt: g.CreatedAt,
			Members: make([]MemberView, 0, len(members)), OpenTasks: open.Counts[g.ID]}
		for _, m := range members {
			v.Members = append(v.Members, MemberView{User: names[m.UserID], Role: m.Role, JoinedAt: m.JoinedAt})
		}
		out = append(out, v)
	}
	return out, nil
}

// presentOne is present for one group.
func presentOne(ctx context.Context, uid string, g Group) (View, error) {
	views, err := present(ctx, uid, g)
	if err != nil {
		return View{}, err
	}
	return views[0], nil
}

// List returns the caller's groups, by name.
func List(ctx context.Context, _ kit.Empty) (ListOutput, error) {
	uid, err := me(ctx)
	if err != nil {
		return ListOutput{}, err
	}
	gs, err := Groups.Find(ctx, "member", uid)
	if err != nil {
		return ListOutput{}, err
	}
	slices.SortFunc(gs, func(a, b Group) int {
		return cmp.Or(cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name)), cmp.Compare(a.ID, b.ID))
	})
	views, err := present(ctx, uid, gs...)
	if err != nil {
		return ListOutput{}, err
	}
	return ListOutput{Groups: views}, nil
}

// CreateInput is a new group.
type CreateInput struct {
	Name string `json:"name" validate:"required,maxlen=60"`
	// Color is one of Colors; empty picks one from the name.
	Color string `json:"color" validate:"maxlen=16"`
}

// color checks a group color, or picks one for a name when there is none.
func color(c, name string) (string, error) {
	if c == "" {
		h := fnv.New32a()
		_, _ = h.Write([]byte(name))
		return Colors[h.Sum32()%uint32(len(Colors))], nil
	}
	if !slices.Contains(Colors, c) {
		return "", wire.Invalid("color", "oneof", "must be one of "+strings.Join(Colors, ", "))
	}
	return c, nil
}

// Create starts a group; its creator is its owner.
func Create(ctx context.Context, in CreateInput) (View, error) {
	uid, err := me(ctx)
	if err != nil {
		return View{}, err
	}
	name, err := wire.Line("name", in.Name, 60)
	if err != nil {
		return View{}, err
	}
	c, err := color(in.Color, name)
	if err != nil {
		return View{}, err
	}
	now := wire.Now(ctx)
	g := Group{ID: kit.NewID("group"), Name: name, Color: c, CreatedAt: now, UpdatedAt: now,
		Members: []Membership{{UserID: uid, Role: Owner, JoinedAt: now}}}
	if err := Groups.Insert(ctx, g); err != nil {
		return View{}, err
	}
	return presentOne(ctx, uid, g)
}

// GroupID addresses a group.
type GroupID struct {
	ID string `path:"id"`
}

// Get returns one of the caller's groups.
func Get(ctx context.Context, in GroupID) (View, error) {
	uid, err := me(ctx)
	if err != nil {
		return View{}, err
	}
	g, _, err := membership(ctx, in.ID, uid)
	if err != nil {
		return View{}, err
	}
	return presentOne(ctx, uid, g)
}

// UpdateInput renames or recolors a group: an absent member is unchanged.
type UpdateInput struct {
	ID    string  `path:"id"`
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
}

// Update renames or recolors a group. Its owner and admins may.
func Update(ctx context.Context, in UpdateInput) (View, error) {
	uid, err := me(ctx)
	if err != nil {
		return View{}, err
	}
	g, role, err := membership(ctx, in.ID, uid)
	if err != nil {
		return View{}, err
	}
	if !role.manages() {
		return View{}, kit.Forbidden("only the group's owner and admins can change it")
	}
	name, c := g.Name, g.Color
	if in.Name != nil {
		if name, err = wire.Line("name", *in.Name, 60); err != nil {
			return View{}, err
		}
	}
	if in.Color != nil {
		if *in.Color == "" {
			return View{}, wire.Invalid("color", "oneof", "must be one of "+strings.Join(Colors, ", "))
		}
		if c, err = color(*in.Color, name); err != nil {
			return View{}, err
		}
	}
	now := wire.Now(ctx)
	g, err = Groups.Update(ctx, g.ID, func(g *Group) error {
		g.Name, g.Color, g.UpdatedAt = name, c, now
		return nil
	})
	if err != nil {
		return View{}, err
	}
	return presentOne(ctx, uid, g)
}

// Delete deletes a group: its pending invitations are revoked, and its tasks
// go back to their owners' lists. Only its owner may.
func Delete(ctx context.Context, in GroupID) (kit.Empty, error) {
	uid, err := me(ctx)
	if err != nil {
		return kit.Empty{}, err
	}
	g, role, err := membership(ctx, in.ID, uid)
	if err != nil {
		return kit.Empty{}, err
	}
	if role != Owner {
		return kit.Empty{}, kit.Forbidden("only the group's owner can delete it")
	}
	invitations, err := Invitations.Find(ctx, "group", g.ID)
	if err != nil {
		return kit.Empty{}, err
	}
	for _, inv := range invitations {
		if inv.Status != Pending {
			continue
		}
		if _, err := InvitationLifecycle.Fire(ctx, inv.ID, "revoke"); err != nil && !wire.Is(err, kit.CodeConflict) {
			return kit.Empty{}, err
		}
	}
	if err := Groups.Delete(ctx, g.ID); err != nil {
		return kit.Empty{}, err
	}
	return kit.Empty{}, announce(ctx, g, KindDeleted, uid, "", "", g.memberIDs())
}

// InviteInput names the contact to invite.
type InviteInput struct {
	ID     string `path:"id"`
	UserID string `json:"userId" validate:"required,maxlen=64"`
}

// InvitationView is an invitation into a group.
type InvitationView struct {
	ID      string           `json:"id"`
	Group   GroupRef         `json:"group"`
	Inviter identity.UserRef `json:"inviter"`
	// Invitee is who is invited, for the inviter's answer.
	Invitee   *identity.UserRef `json:"invitee,omitempty"`
	CreatedAt time.Time         `json:"createdAt"`
	Status    InvitationStatus  `json:"status"`
}

// Invite asks one of the caller's contacts to join a group. Its owner and
// admins may invite.
func Invite(ctx context.Context, in InviteInput) (InvitationView, error) {
	uid, err := me(ctx)
	if err != nil {
		return InvitationView{}, err
	}
	g, role, err := membership(ctx, in.ID, uid)
	if err != nil {
		return InvitationView{}, err
	}
	switch {
	case !role.manages():
		return InvitationView{}, kit.Forbidden("only the group's owner and admins can invite")
	case in.UserID == uid || g.RoleOf(in.UserID) != "":
		return InvitationView{}, kit.Conflict("this person is already a member")
	}
	check, err := contacts.CheckAPI.Call(ctx, contacts.CheckInput{A: uid, B: in.UserID})
	if err != nil {
		return InvitationView{}, err
	}
	if !check.Contacts {
		return InvitationView{}, kit.Forbidden("you can only invite your contacts")
	}
	pending, err := Invitations.Find(ctx, "invitee", in.UserID)
	if err != nil {
		return InvitationView{}, err
	}
	for _, inv := range pending {
		if inv.GroupID == g.ID && inv.Status == Pending {
			return InvitationView{}, kit.Conflict("this person is already invited")
		}
	}
	inv, err := InvitationLifecycle.Start(ctx, Invitation{
		ID: kit.NewID("invitation"), GroupID: g.ID, InviterID: uid, InviteeID: in.UserID, CreatedAt: wire.Now(ctx),
	})
	if err != nil {
		return InvitationView{}, err
	}
	names, err := identity.Directory(ctx, uid, in.UserID)
	if err != nil {
		return InvitationView{}, err
	}
	invitee := names[in.UserID]
	return InvitationView{ID: inv.ID, Group: g.Ref(), Inviter: names[uid], Invitee: &invitee, CreatedAt: inv.CreatedAt, Status: inv.Status}, nil
}

// InvitationsOutput is the invitations waiting for the caller.
type InvitationsOutput struct {
	Invitations []InvitationView `json:"invitations"`
}

// ListInvitations returns the invitations waiting for the caller's answer,
// newest first.
func ListInvitations(ctx context.Context, _ kit.Empty) (InvitationsOutput, error) {
	uid, err := me(ctx)
	if err != nil {
		return InvitationsOutput{}, err
	}
	received, err := Invitations.Find(ctx, "invitee", uid)
	if err != nil {
		return InvitationsOutput{}, err
	}
	out := InvitationsOutput{Invitations: []InvitationView{}}
	var inviters []string
	groups := map[string]Group{}
	for _, inv := range received {
		if inv.Status != Pending {
			continue
		}
		g, err := Groups.Get(ctx, inv.GroupID)
		if wire.Is(err, kit.CodeNotFound) {
			continue
		}
		if err != nil {
			return InvitationsOutput{}, err
		}
		groups[inv.ID] = g
		inviters = append(inviters, inv.InviterID)
	}
	names, err := identity.Directory(ctx, inviters...)
	if err != nil {
		return InvitationsOutput{}, err
	}
	for _, inv := range received {
		if g, ok := groups[inv.ID]; ok {
			out.Invitations = append(out.Invitations, InvitationView{
				ID: inv.ID, Group: g.Ref(), Inviter: names[inv.InviterID], CreatedAt: inv.CreatedAt, Status: inv.Status,
			})
		}
	}
	slices.SortFunc(out.Invitations, func(a, b InvitationView) int { return b.CreatedAt.Compare(a.CreatedAt) })
	return out, nil
}

// InvitationID addresses an invitation.
type InvitationID struct {
	ID string `path:"id"`
}

// received returns an invitation waiting for the caller's answer.
func received(ctx context.Context, id, uid string) (Invitation, error) {
	inv, err := Invitations.Get(ctx, id)
	if wire.Is(err, kit.CodeNotFound) || (err == nil && inv.InviteeID != uid) {
		return Invitation{}, kit.NotFound("no such invitation")
	}
	if err != nil {
		return Invitation{}, err
	}
	if inv.Status != Pending {
		return Invitation{}, kit.Conflict("this invitation is not waiting for an answer any more")
	}
	return inv, nil
}

// AcceptInvitation joins the group the caller is invited to.
func AcceptInvitation(ctx context.Context, in InvitationID) (View, error) {
	uid, err := me(ctx)
	if err != nil {
		return View{}, err
	}
	inv, err := received(ctx, in.ID, uid)
	if err != nil {
		return View{}, err
	}
	if _, err := InvitationLifecycle.Fire(ctx, inv.ID, "accept"); err != nil {
		return View{}, err
	}
	g, _, err := membership(ctx, inv.GroupID, uid)
	if err != nil {
		return View{}, err
	}
	return presentOne(ctx, uid, g)
}

// DeclineInvitation turns an invitation down.
func DeclineInvitation(ctx context.Context, in InvitationID) (kit.Empty, error) {
	uid, err := me(ctx)
	if err != nil {
		return kit.Empty{}, err
	}
	inv, err := received(ctx, in.ID, uid)
	if err != nil {
		return kit.Empty{}, err
	}
	_, err = InvitationLifecycle.Fire(ctx, inv.ID, "decline")
	return kit.Empty{}, err
}

// MemberInput addresses a member of a group.
type MemberInput struct {
	ID     string `path:"id"`
	UserID string `path:"userId"`
}

// RemoveMember takes a member out of a group. Anyone may leave, but the
// owner, who deletes the group instead; the owner removes anyone, an admin
// removes members.
func RemoveMember(ctx context.Context, in MemberInput) (kit.Empty, error) {
	uid, err := me(ctx)
	if err != nil {
		return kit.Empty{}, err
	}
	g, role, err := membership(ctx, in.ID, uid)
	if err != nil {
		return kit.Empty{}, err
	}
	target := g.RoleOf(in.UserID)
	switch {
	case target == "":
		return kit.Empty{}, kit.NotFound("no such member")
	case in.UserID == uid && role == Owner:
		return kit.Empty{}, kit.Conflict("the owner cannot leave the group: delete it instead")
	case in.UserID != uid && (target == Owner || role == Member || (role == Admin && target == Admin)):
		return kit.Empty{}, kit.Forbidden("you cannot remove this member")
	}
	now := wire.Now(ctx)
	g, err = Groups.Update(ctx, g.ID, func(g *Group) error {
		g.Members = slices.DeleteFunc(g.Members, func(m Membership) bool { return m.UserID == in.UserID })
		g.UpdatedAt = now
		return nil
	})
	if err != nil {
		return kit.Empty{}, err
	}
	if in.UserID == uid {
		return kit.Empty{}, announce(ctx, g, KindLeft, uid, uid, "", g.memberIDs())
	}
	return kit.Empty{}, announce(ctx, g, KindRemoved, uid, in.UserID, "", append(g.memberIDs(), in.UserID))
}

// RoleInput is a member's new role.
type RoleInput struct {
	ID     string `path:"id"`
	UserID string `path:"userId"`
	Role   Role   `json:"role" validate:"required,oneof=admin|member"`
}

// SetRole makes a member an admin, or an admin a member. Only the owner may.
func SetRole(ctx context.Context, in RoleInput) (View, error) {
	uid, err := me(ctx)
	if err != nil {
		return View{}, err
	}
	g, role, err := membership(ctx, in.ID, uid)
	if err != nil {
		return View{}, err
	}
	switch target := g.RoleOf(in.UserID); {
	case role != Owner:
		return View{}, kit.Forbidden("only the group's owner can change roles")
	case target == "":
		return View{}, kit.NotFound("no such member")
	case target == Owner:
		return View{}, kit.Forbidden("the owner's role cannot change")
	}
	now := wire.Now(ctx)
	g, err = Groups.Update(ctx, g.ID, func(g *Group) error {
		for i := range g.Members {
			if g.Members[i].UserID == in.UserID {
				g.Members[i].Role = in.Role
			}
		}
		g.UpdatedAt = now
		return nil
	})
	if err != nil {
		return View{}, err
	}
	return presentOne(ctx, uid, g)
}
