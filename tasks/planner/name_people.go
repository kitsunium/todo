// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"context"
)

// execute is NamePeople's body: the owner, the assignee and the members
// shared with, task after task, are looked up in the directory once each
// whoever else names them, and a member it does not know is left out. Only a
// name is kept: nothing else of the member crosses the port.
func (h *NamePeopleHandler) execute(ctx context.Context, in *NamePeopleInput) (NamePeopleOutput, error) {
	out := NamePeopleOutput{People: []Person{}}
	asked := map[string]bool{}
	for _, t := range in.Tasks {
		ids := append([]string{t.OwnerID, t.AssigneeID}, t.SharedWith...)
		for _, id := range ids {
			if id == "" || asked[id] {
				continue
			}
			asked[id] = true
			p, known, err := h.people.ByID(ctx, id)
			if err != nil {
				return NamePeopleOutput{}, err
			}
			if known {
				out.People = append(out.People, Person{ID: id, Name: p.Name})
			}
		}
	}
	return out, nil
}
