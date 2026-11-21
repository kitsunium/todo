// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"slices"
	"testing"
)

// givenRemoveMemberLeavesAGroup builds RemoveMember for "leaves a group": the use case, its input, and
// what its output must be.
func givenRemoveMemberLeavesAGroup(t *testing.T) (RemoveMember, RemoveMemberInput, func(*testing.T, RemoveMemberOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.leave(), RemoveMemberInput{User: "user_cy", ID: g.ID, Member: "user_cy"}, func(t *testing.T, _ RemoveMemberOutput) {
		t.Helper()
		kept := w.groups[g.ID]
		if slices.Contains(kept.MemberIDs(), "user_cy") || len(kept.Members) != 2 || !kept.UpdatedAt.After(epoch) {
			t.Errorf("the group kept: %+v", kept)
		}
		want := "group-left by user_cy on " + g.ID + " to user_bob,user_ada"
		if len(w.notices) != 1 || w.notices[0] != want {
			t.Errorf("the notices: %q, want %q", w.notices, want)
		}
	}
}

// givenRemoveMemberRemovesAMember builds RemoveMember for "removes a member": the use case, its input, and
// what its output must be.
func givenRemoveMemberRemovesAMember(t *testing.T) (RemoveMember, RemoveMemberInput, func(*testing.T, RemoveMemberOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.leave(), RemoveMemberInput{User: "user_bob", ID: g.ID, Member: "user_cy"}, func(t *testing.T, _ RemoveMemberOutput) {
		t.Helper()
		if kept := w.groups[g.ID]; slices.Contains(kept.MemberIDs(), "user_cy") || len(kept.Members) != 2 {
			t.Errorf("the group kept: %+v", kept)
		}
		want := "group-removed by user_bob on " + g.ID + " to user_bob,user_ada,user_cy"
		if len(w.notices) != 1 || w.notices[0] != want {
			t.Errorf("the notices: %q, want %q (the removed one is told with the others, not the actor)", w.notices, want)
		}
	}
}

// givenRemoveMemberRefusesAnUnknownMember builds RemoveMember for "refuses an unknown member": the use case and the input it
// refuses with team.ErrNoMember.
func givenRemoveMemberRefusesAnUnknownMember(t *testing.T) (RemoveMember, RemoveMemberInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.leave(), RemoveMemberInput{User: "user_ada", ID: g.ID, Member: "user_dee"}
}

// givenRemoveMemberRefusesTheOwnerWhoLeaves builds RemoveMember for "refuses the owner who leaves": the use case and the input it
// refuses with team.ErrOwnerStays.
func givenRemoveMemberRefusesTheOwnerWhoLeaves(t *testing.T) (RemoveMember, RemoveMemberInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.leave(), RemoveMemberInput{User: "user_ada", ID: g.ID, Member: "user_ada"}
}

// givenRemoveMemberRefusesAPlainMemberWhoRemoves builds RemoveMember for "refuses a plain member who removes": the use case and the input it
// refuses with team.ErrCannotRemove.
func givenRemoveMemberRefusesAPlainMemberWhoRemoves(t *testing.T) (RemoveMember, RemoveMemberInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.leave(), RemoveMemberInput{User: "user_cy", ID: g.ID, Member: "user_bob"}
}

// givenRemoveMemberRefusesAnAdminWhoRemovesAnAdmin builds RemoveMember for "refuses an admin who removes an admin": the use case and the input it
// refuses with team.ErrCannotRemove.
func givenRemoveMemberRefusesAnAdminWhoRemovesAnAdmin(t *testing.T) (RemoveMember, RemoveMemberInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	other := g.Members[0]
	other.UserID, other.Role = "user_dee", "admin"
	g.Members = append(g.Members, other)
	w.groups[g.ID] = g
	return w.leave(), RemoveMemberInput{User: "user_bob", ID: g.ID, Member: "user_dee"}
}

// givenRemoveMemberRefusesTheRemovalOfTheOwner builds RemoveMember for "refuses the removal of the owner": the use case and the input it
// refuses with team.ErrCannotRemove.
func givenRemoveMemberRefusesTheRemovalOfTheOwner(t *testing.T) (RemoveMember, RemoveMemberInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.leave(), RemoveMemberInput{User: "user_bob", ID: g.ID, Member: "user_ada"}
}

// givenRemoveMemberRefusesTheGroupOfOthers builds RemoveMember for "refuses the group of others": the use case and the input it
// refuses with team.ErrNotFound.
func givenRemoveMemberRefusesTheGroupOfOthers(t *testing.T) (RemoveMember, RemoveMemberInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	return w.leave(), RemoveMemberInput{User: "user_dee", ID: g.ID, Member: "user_cy"}
}
