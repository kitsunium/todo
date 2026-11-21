// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenGetGroupShowsAGroupWithItsMembersRanked builds GetGroup for "shows a group with its members ranked": the use case, its input, and
// what its output must be.
func givenGetGroupShowsAGroupWithItsMembersRanked(t *testing.T) (GetGroup, GetGroupInput, func(*testing.T, GetGroupOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.get(), GetGroupInput{User: "user_cy", ID: g.ID}, func(t *testing.T, out GetGroupOutput) {
		t.Helper()
		v := out.Group
		if v.ID != g.ID || v.Name != "Core" || v.Role != team.RoleMember {
			t.Errorf("the group as cy sees it: %+v", v)
		}
		var got []string
		for _, m := range v.Members {
			got = append(got, m.User.Name+":"+string(m.Role))
		}
		if len(got) != 3 || got[0] != "ada:owner" || got[1] != "bob:admin" || got[2] != "cy:member" {
			t.Errorf("the members: %v, want the owner, the admin, the member", got)
		}
	}
}

// givenGetGroupRefusesTheGroupOfOthers builds GetGroup for "refuses the group of others": the use case and the input it
// refuses with team.ErrNotFound.
func givenGetGroupRefusesTheGroupOfOthers(t *testing.T) (GetGroup, GetGroupInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.get(), GetGroupInput{User: "user_dee", ID: g.ID}
}
