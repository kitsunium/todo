// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package roster

import (
	"testing"
)

// givenEditGroupEditsAGroup builds EditGroup for "edits a group": the use case, its input, and
// what its output must be.
func givenEditGroupEditsAGroup(t *testing.T) (EditGroup, EditGroupInput, func(*testing.T, EditGroupOutput)) {
	t.Helper()
	w := newWorld()
	g := w.core()
	name := " Platform "
	return w.edit(), EditGroupInput{User: "user_bob", ID: g.ID, Name: &name}, func(t *testing.T, out EditGroupOutput) {
		t.Helper()
		kept := w.groups[g.ID]
		if out.Group.Name != "Platform" || kept.Name != "Platform" || kept.Color != "teal" || out.Group.Color != "teal" {
			t.Errorf("an edit of the name only: %+v, kept %+v", out.Group, kept)
		}
		if !kept.UpdatedAt.After(epoch) || kept.CreatedAt != epoch || len(kept.Members) != 3 {
			t.Errorf("the group kept: %+v", kept)
		}
	}
}

// givenEditGroupRefusesTheGroupOfOthers builds EditGroup for "refuses the group of others": the use case and the input it
// refuses with team.ErrNotFound.
func givenEditGroupRefusesTheGroupOfOthers(t *testing.T) (EditGroup, EditGroupInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	name := "Mine now"
	return w.edit(), EditGroupInput{User: "user_dee", ID: g.ID, Name: &name}
}

// givenEditGroupRefusesAPlainMember builds EditGroup for "refuses a plain member": the use case and the input it
// refuses with team.ErrNotManager.
func givenEditGroupRefusesAPlainMember(t *testing.T) (EditGroup, EditGroupInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	name := "Mine now"
	return w.edit(), EditGroupInput{User: "user_cy", ID: g.ID, Name: &name}
}

// givenEditGroupRefusesABlankColor builds EditGroup for "refuses a blank color": the use case and the input it
// refuses with team.ErrBadColor.
func givenEditGroupRefusesABlankColor(t *testing.T) (EditGroup, EditGroupInput) {
	t.Helper()
	w := newWorld()
	g := w.core()
	color := ""
	return w.edit(), EditGroupInput{User: "user_ada", ID: g.ID, Color: &color}
}
