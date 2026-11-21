// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package groupstore

import (
	"context"
	"errors"
	"slices"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/groups/team"
)

// groups is Groups's own state: the store is the package's, it holds nothing.
type groups struct {
}

// newGroups is NewGroups's body.
func newGroups() *Groups {
	return &Groups{groups{}}
}

// deleteBody is Groups.Delete's body: a group already gone is deleted.
func (self *Groups) deleteBody(ctx context.Context, id string) (err error) {
	err = GroupsStore.Delete(ctx, id)
	var refused *app.Error
	if errors.As(err, &refused) && refused.Code == app.WireNotFound {
		return nil
	}
	return err
}

// get is Groups.Get's body. It filters instead of reading the key: a group
// that is not there is an answer, never a failed read.
func (self *Groups) get(ctx context.Context, id string) (r0 team.Group, r1 bool, err error) {
	found, err := GroupsStore.Filter(ctx, func(r GroupRecord) bool { return r.ID == id })
	if err != nil || len(found) == 0 {
		return team.Group{}, false, err
	}
	return groupOf(found[0]), true, nil
}

// newID is Groups.NewID's body: a time-ordered id.
func (self *Groups) newID(ctx context.Context) (r0 string, err error) {
	return app.NewID("group"), nil
}

// of is Groups.Of's body: the groups that hold the member.
func (self *Groups) of(ctx context.Context, user string) (r0 []team.Group, err error) {
	found, err := GroupsStore.Filter(ctx, func(r GroupRecord) bool {
		return slices.ContainsFunc(r.Members, func(m team.Membership) bool { return m.UserID == user })
	})
	if err != nil {
		return nil, err
	}
	r0 = make([]team.Group, len(found))
	for i, r := range found {
		r0[i] = groupOf(r)
	}
	return r0, nil
}

// save is Groups.Save's body: the group replaces what its id held.
func (self *Groups) save(ctx context.Context, g team.Group) (err error) {
	return GroupsStore.Put(ctx, recordOfGroup(g))
}

// groupOf is the group a record keeps.
func groupOf(r GroupRecord) team.Group {
	return team.Group{
		ID: r.ID, Name: r.Name, Color: r.Color, Members: slices.Clone(r.Members),
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

// recordOfGroup is a group as the store keeps it.
func recordOfGroup(g team.Group) GroupRecord {
	return GroupRecord{
		ID: g.ID, Name: g.Name, Color: g.Color, Members: slices.Clone(g.Members),
		CreatedAt: g.CreatedAt, UpdatedAt: g.UpdatedAt,
	}
}

// change is Groups.Change's body: the store applies the change to the record it
// holds at that instant, alone among its writers, so two changes made at
// once both stay. A group that is not there is ErrNotFound. The change's own error keeps nothing.
func (a *Groups) change(ctx context.Context, id string, change func(*team.Group) error) (r0 team.Group, err error) {
	rec, err := GroupsStore.Update(ctx, id, func(r *GroupRecord) error {
		v := groupOf(*r)
		if err := change(&v); err != nil {
			return err
		}
		v.ID = r.ID // a change never renames a record
		*r = recordOfGroup(v)
		return nil
	})
	var refused *app.Error
	switch {
	case errors.As(err, &refused) && refused.Code == app.WireNotFound:
		return team.Group{}, team.ErrNotFound
	case err != nil:
		return team.Group{}, err
	}
	return groupOf(rec), nil
}
