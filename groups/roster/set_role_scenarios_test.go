// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenSetRoleMakesAMemberAnAdmin builds SetRole for "makes a member an admin": the use case, its input, and
// what its output must be.
func givenSetRoleMakesAMemberAnAdmin(t *testing.T) (SetRole, SetRoleInput, func(*testing.T, SetRoleOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	before := append([]team.Membership(nil), g.Members...)
	return w.setRole(), SetRoleInput{User: "user_ada", ID: g.ID, Member: "user_cy", Role: "admin"}, func(t *testing.T, out SetRoleOutput) {
		t.Helper()
		if role := w.role(g.ID, "user_cy"); role != team.RoleAdmin {
			t.Errorf("cy is %q in the store", role)
		}
		if w.role(g.ID, "user_bob") != team.RoleAdmin || w.role(g.ID, "user_ada") != team.RoleOwner || out.Group.Role != team.RoleOwner {
			t.Errorf("the others' roles moved: %+v", w.groups[g.ID].Members)
		}
		if before[0].Role != team.RoleMember {
			t.Errorf("the group the caller held was changed in place: %+v", before)
		}
		want := "group-role by user_ada on " + g.ID + " to user_cy"
		if len(w.notices) != 1 || w.notices[0] != want {
			t.Errorf("the notices: %q, want %q", w.notices, want)
		}
	}
}

// givenSetRoleRefusesAnUnknownRole builds SetRole for "refuses an unknown role": the use case and the input it
// refuses with team.ErrBadRole.
func givenSetRoleRefusesAnUnknownRole(t *testing.T) (SetRole, SetRoleInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.setRole(), SetRoleInput{User: "user_ada", ID: g.ID, Member: "user_cy", Role: "owner"}
}

// givenSetRoleRefusesTheGroupOfOthers builds SetRole for "refuses the group of others": the use case and the input it
// refuses with team.ErrNotFound.
func givenSetRoleRefusesTheGroupOfOthers(t *testing.T) (SetRole, SetRoleInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.setRole(), SetRoleInput{User: "user_dee", ID: g.ID, Member: "user_cy", Role: "admin"}
}

// givenSetRoleRefusesAnAdmin builds SetRole for "refuses an admin": the use case and the input it
// refuses with team.ErrNotOwner.
func givenSetRoleRefusesAnAdmin(t *testing.T) (SetRole, SetRoleInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.setRole(), SetRoleInput{User: "user_bob", ID: g.ID, Member: "user_cy", Role: "admin"}
}

// givenSetRoleRefusesAnUnknownMember builds SetRole for "refuses an unknown member": the use case and the input it
// refuses with team.ErrNoMember.
func givenSetRoleRefusesAnUnknownMember(t *testing.T) (SetRole, SetRoleInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.setRole(), SetRoleInput{User: "user_ada", ID: g.ID, Member: "user_dee", Role: "admin"}
}

// givenSetRoleRefusesToChangeTheOwner builds SetRole for "refuses to change the owner": the use case and the input it
// refuses with team.ErrOwnerRole.
func givenSetRoleRefusesToChangeTheOwner(t *testing.T) (SetRole, SetRoleInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.setRole(), SetRoleInput{User: "user_ada", ID: g.ID, Member: "user_ada", Role: "member"}
}
