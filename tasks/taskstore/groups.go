// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package taskstore

import (
	"context"

	"github.com/kitsunium/todo/groups/groupsapi"
)

// groups is Groups's own state: it asks the groups, it holds nothing.
type groups struct {
}

// newGroups is NewGroups's body.
func newGroups() *Groups {
	return &Groups{groups{}}
}

// memberships is Groups.Memberships's body: the groups' query "memberships",
// as the id of each group and the member's role in it.
func (a *Groups) memberships(ctx context.Context, user string) (r0 map[string]string, err error) {
	found, err := groupsapi.MembershipsQuery.Ask(ctx, groupsapi.WhoseGroups{User: user})
	if err != nil {
		return nil, err
	}
	r0 = make(map[string]string, len(found.Memberships))
	for _, m := range found.Memberships {
		r0[m.Group.ID] = string(m.Role)
	}
	return r0, nil
}

// roleOf is Groups.RoleOf's body: the groups' query "role".
func (a *Groups) roleOf(ctx context.Context, group string, user string) (r0 string, err error) {
	v, err := groupsapi.RoleQuery.Ask(ctx, groupsapi.RoleAsked{Group: group, User: user})
	return string(v.Role), err
}
