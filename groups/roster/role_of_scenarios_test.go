// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"

	"github.com/kitsunium/todo/groups/team"
)

// givenRoleOfTellsTheRoleOfAMemberAndNothingOfAStranger builds RoleOf for "tells the role of a member and nothing of a stranger": the use case, its input, and
// what its output must be.
func givenRoleOfTellsTheRoleOfAMemberAndNothingOfAStranger(t *testing.T) (RoleOf, RoleOfInput, func(*testing.T, RoleOfOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	uc := w.roleOf()
	return uc, RoleOfInput{Group: g.ID, User: "user_bob"}, func(t *testing.T, out RoleOfOutput) {
		t.Helper()
		if out.Role != team.RoleAdmin {
			t.Errorf("bob is %q", out.Role)
		}
		for _, in := range []RoleOfInput{{Group: g.ID, User: "user_dee"}, {Group: "group_nobody", User: "user_ada"}, {Group: "", User: "user_ada"}} {
			if got, err := uc.Execute(t.Context(), in); err != nil || got.Role != "" {
				t.Errorf("%+v is answered %q (%v), want nothing", in, got.Role, err)
			}
		}
	}
}
