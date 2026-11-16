// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"slices"
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenCreateGroupCreatesAGroup builds CreateGroup for "creates a group": the use case, its input, and
// what its output must be.
func givenCreateGroupCreatesAGroup(t *testing.T) (CreateGroup, CreateGroupInput, func(*testing.T, CreateGroupOutput)) {
	t.Helper()
	w := newWorld()
	return w.create(), CreateGroupInput{User: "user_ada", Name: "  Core  ", Color: "blue"}, func(t *testing.T, out CreateGroupOutput) {
		t.Helper()
		v := out.Group
		if v.Name != "Core" || v.Color != "blue" || v.Role != team.RoleOwner || len(v.Members) != 1 || v.Members[0].User.Name != "ada" || v.Members[0].Role != team.RoleOwner {
			t.Errorf("the group: %+v", v)
		}
		kept, ok := w.groups[v.ID]
		if !ok || kept.Name != "Core" || !slices.Equal(kept.MemberIDs(), []string{"user_ada"}) || !kept.CreatedAt.After(epoch) || kept.CreatedAt != kept.UpdatedAt {
			t.Errorf("the store keeps %+v", kept)
		}
		if len(w.notices) != 0 {
			t.Errorf("nobody is told of a group nobody is in: %q", w.notices)
		}
	}
}

// givenCreateGroupPicksAColorFromTheName builds CreateGroup for "picks a color from the name": the use case, its input, and
// what its output must be.
func givenCreateGroupPicksAColorFromTheName(t *testing.T) (CreateGroup, CreateGroupInput, func(*testing.T, CreateGroupOutput)) {
	t.Helper()
	w := newWorld()
	return w.create(), CreateGroupInput{User: "user_ada", Name: "Core"}, func(t *testing.T, out CreateGroupOutput) {
		t.Helper()
		want, _ := team.ColorFor("", "Core")
		if out.Group.Color == "" || out.Group.Color != want {
			t.Errorf("the color: %q, want %q", out.Group.Color, want)
		}
	}
}

// givenCreateGroupRefusesABlankName builds CreateGroup for "refuses a blank name": the use case and the input it
// refuses with team.ErrBadName.
func givenCreateGroupRefusesABlankName(t *testing.T) (CreateGroup, CreateGroupInput) {
	t.Helper()
	return newWorld().create(), CreateGroupInput{User: "user_ada", Name: "   "}
}

// givenCreateGroupRefusesAnUnknownColor builds CreateGroup for "refuses an unknown color": the use case and the input it
// refuses with team.ErrBadColor.
func givenCreateGroupRefusesAnUnknownColor(t *testing.T) (CreateGroup, CreateGroupInput) {
	t.Helper()
	return newWorld().create(), CreateGroupInput{User: "user_ada", Name: "Core", Color: "chartreuse"}
}
