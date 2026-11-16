// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"context"

	"github.com/kitsunium/todo/groups/team"
)

// The kinds of notice a group gives.
const (
	kindInvited = "group-invited"
	kindJoined  = "group-joined"
	kindLeft    = "group-left"
	kindRemoved = "group-removed"
	kindRole    = "group-role"
	kindDeleted = "group-deleted"
)

// personOf is a member as the directory names them; one the directory no
// longer has keeps their id.
func personOf(ctx context.Context, people PeopleV1, id string) (Person, error) {
	p, found, err := people.ByID(ctx, id)
	if err != nil {
		return Person{}, err
	}
	if !found {
		return Person{ID: id}, nil
	}
	return p, nil
}

// present is a group as the member who looks sees it: the members ranked,
// named by the directory.
func present(ctx context.Context, people PeopleV1, user string, g team.Group) (GroupView, error) {
	v := GroupView{
		ID: g.ID, Name: g.Name, Color: g.Color, Role: g.RoleOf(user),
		Members: make([]MemberView, 0, len(g.Members)), CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
	for _, m := range team.Ranked(g.Members) {
		who, err := personOf(ctx, people, m.UserID)
		if err != nil {
			return GroupView{}, err
		}
		v.Members = append(v.Members, MemberView{User: who, Role: m.Role, JoinedAt: m.JoinedAt})
	}
	return v, nil
}

// refOf is how a group is shown elsewhere.
func refOf(g team.Group) GroupRef {
	return GroupRef{ID: g.ID, Name: g.Name, Color: g.Color}
}

// roleIn is the role a member has in a group they look for: a group that is
// not there, or that does not hold them, does not exist for them.
func roleIn(g team.Group, found bool, user string) (team.Role, error) {
	if !found {
		return "", team.ErrNotFound
	}
	role := g.RoleOf(user)
	if role == "" {
		return "", team.ErrNotFound
	}
	return role, nil
}
