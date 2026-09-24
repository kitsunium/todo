package groups

import (
	"context"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/internal/wire"
)

// What other services may ask about groups, in-process: tasks decides who
// sees a task and who may delete it, activity names the groups in the feeds.
var (
	// RoleAPI tells a user's role in a group.
	RoleAPI = Service.Endpoint("GET /internal/groups/role", RoleIn, kit.Private())

	// BatchAPI turns group IDs into how groups are shown.
	BatchAPI = Service.Endpoint("POST /internal/groups/batch", Batch, kit.Private())

	// MembershipsAPI lists the groups a user belongs to, with their role.
	MembershipsAPI = Service.Endpoint("GET /internal/groups/memberships", MembershipsOf, kit.Private())

	// MembersAPI lists the members of a group.
	MembersAPI = Service.Endpoint("GET /internal/groups/members", MembersOf, kit.Private())
)

// RoleQuery names a group and a user.
type RoleQuery struct {
	Group string `query:"group"`
	User  string `query:"user"`
}

// RoleOutput is the user's role: "" when they are not a member.
type RoleOutput struct {
	Role Role `json:"role"`
}

// RoleIn returns a user's role in a group; a group that does not exist has
// no members.
func RoleIn(ctx context.Context, in RoleQuery) (RoleOutput, error) {
	if in.Group == "" {
		return RoleOutput{}, nil
	}
	g, err := Groups.Get(ctx, in.Group)
	if wire.Is(err, kit.CodeNotFound) {
		return RoleOutput{}, nil
	}
	if err != nil {
		return RoleOutput{}, err
	}
	return RoleOutput{Role: g.RoleOf(in.User)}, nil
}

// IDs are group IDs.
type IDs struct {
	IDs []string `json:"ids"`
}

// BatchOutput is the groups found, in the order asked; an unknown ID is
// left out.
type BatchOutput struct {
	Groups []GroupRef `json:"groups"`
}

// Batch returns how the groups with the given IDs are shown.
func Batch(ctx context.Context, in IDs) (BatchOutput, error) {
	out := BatchOutput{Groups: []GroupRef{}}
	seen := map[string]bool{}
	for _, id := range in.IDs {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		g, err := Groups.Get(ctx, id)
		if wire.Is(err, kit.CodeNotFound) {
			continue
		}
		if err != nil {
			return BatchOutput{}, err
		}
		out.Groups = append(out.Groups, g.Ref())
	}
	return out, nil
}

// UserQuery names a user.
type UserQuery struct {
	User string `query:"user"`
}

// Belonging is one group a user belongs to, and their role in it.
type Belonging struct {
	Group GroupRef `json:"group"`
	Role  Role     `json:"role"`
}

// MembershipsOutput is the groups a user belongs to.
type MembershipsOutput struct {
	Memberships []Belonging `json:"memberships"`
}

// MembershipsOf lists the groups a user belongs to.
func MembershipsOf(ctx context.Context, in UserQuery) (MembershipsOutput, error) {
	out := MembershipsOutput{Memberships: []Belonging{}}
	if in.User == "" {
		return out, nil
	}
	gs, err := Groups.Find(ctx, "member", in.User)
	if err != nil {
		return MembershipsOutput{}, err
	}
	for _, g := range gs {
		out.Memberships = append(out.Memberships, Belonging{Group: g.Ref(), Role: g.RoleOf(in.User)})
	}
	return out, nil
}

// GroupQuery names a group.
type GroupQuery struct {
	Group string `query:"group"`
}

// MembersOutput is a group's members; empty for a group that does not exist.
type MembersOutput struct {
	Members []Membership `json:"members"`
}

// MembersOf lists the members of a group.
func MembersOf(ctx context.Context, in GroupQuery) (MembersOutput, error) {
	out := MembersOutput{Members: []Membership{}}
	if in.Group == "" {
		return out, nil
	}
	g, err := Groups.Get(ctx, in.Group)
	if wire.Is(err, kit.CodeNotFound) {
		return out, nil
	}
	if err != nil {
		return MembersOutput{}, err
	}
	out.Members = g.Members
	return out, nil
}
