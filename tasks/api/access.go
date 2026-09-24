package api

import (
	"cmp"
	"context"
	"slices"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/groups"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/tasks"
)

// Who sees a task, and who may do what with it:
//
//   - its owner, the users it is shared with and the members of its group see
//     it, and may edit, complete, archive and restore it — its assignee is
//     always one of them;
//   - its owner shares it and moves it between groups;
//   - its owner, or an owner or admin of its group, deletes it.
//
// A task a user does not see does not exist for them: 404, never 403.

// access is what one user may do with one task.
type access struct {
	sees    bool
	deletes bool
	shares  bool
}

// accessOf is the access of uid to t, given uid's role in t's group ("" for
// none).
func accessOf(t tasks.Task, uid string, role groups.Role) access {
	owner := t.OwnerID == uid
	return access{
		sees:    owner || slices.Contains(t.SharedWith, uid) || role != "",
		deletes: owner || role == groups.Owner || role == groups.Admin,
		shares:  owner,
	}
}

// roleIn asks groups for a user's role in a group; "" for no group.
func roleIn(ctx context.Context, group, uid string) (groups.Role, error) {
	if group == "" || uid == "" {
		return "", nil
	}
	r, err := groups.RoleAPI.Call(ctx, groups.RoleQuery{Group: group, User: uid})
	return r.Role, err
}

// accessTo is the access of uid to t, asking groups for their role.
func accessTo(ctx context.Context, t tasks.Task, uid string) (access, error) {
	role, err := roleIn(ctx, t.GroupID, uid)
	return accessOf(t, uid, role), err
}

// memberships are a user's groups, by ID: what a list of tasks needs to know
// about them.
type memberships map[string]groups.Belonging

// membershipsOf asks groups which groups a user belongs to.
func membershipsOf(ctx context.Context, uid string) (memberships, error) {
	found, err := groups.MembershipsAPI.Call(ctx, groups.UserQuery{User: uid})
	if err != nil {
		return nil, err
	}
	ms := memberships{}
	for _, m := range found.Memberships {
		ms[m.Group.ID] = m
	}
	return ms, nil
}

// access is uid's access to t, from uid's memberships.
func (ms memberships) access(t tasks.Task, uid string) access {
	return accessOf(t, uid, ms[t.GroupID].Role)
}

// me is the signed-in user.
func me(ctx context.Context) (string, error) {
	uid, ok := kit.UserID(ctx)
	if !ok {
		return "", kit.Unauthenticated("sign in first")
	}
	return string(uid), nil
}

// errNoTask answers for a task that does not exist, or not for the caller.
func errNoTask() error { return kit.NotFound("no such task") }

// load reads a task the caller sees, with what they may do with it.
func load(ctx context.Context, id string) (tasks.Task, access, string, error) {
	uid, err := me(ctx)
	if err != nil {
		return tasks.Task{}, access{}, "", err
	}
	t, err := tasks.Tasks.Get(ctx, id)
	if wire.Is(err, kit.CodeNotFound) {
		return tasks.Task{}, access{}, "", errNoTask()
	}
	if err != nil {
		return tasks.Task{}, access{}, "", err
	}
	acc, err := accessTo(ctx, t, uid)
	if err != nil {
		return tasks.Task{}, access{}, "", err
	}
	if !acc.sees {
		return tasks.Task{}, access{}, "", errNoTask()
	}
	return t, acc, uid, nil
}

// visible returns every task uid sees, found through the store's indexes —
// owned, shared with them, and on the lists of their groups — ordered by ID,
// with uid's memberships.
func visible(ctx context.Context, uid string) ([]tasks.Task, memberships, error) {
	ms, err := membershipsOf(ctx, uid)
	if err != nil {
		return nil, nil, err
	}
	byID := map[string]tasks.Task{}
	add := func(ts []tasks.Task, err error) error {
		for _, t := range ts {
			byID[t.ID] = t
		}
		return err
	}
	if err := add(tasks.Tasks.Find(ctx, "owner", uid)); err != nil {
		return nil, nil, err
	}
	if err := add(tasks.Tasks.Find(ctx, "shared", uid)); err != nil {
		return nil, nil, err
	}
	for id := range ms {
		if err := add(tasks.Tasks.Find(ctx, "group", id)); err != nil {
			return nil, nil, err
		}
	}
	out := make([]tasks.Task, 0, len(byID))
	for _, t := range byID {
		out = append(out, t)
	}
	slices.SortFunc(out, func(a, b tasks.Task) int { return cmp.Compare(a.ID, b.ID) })
	return out, ms, nil
}

// assignable reports whether a user may be the assignee of t: its owner, a
// user it is shared with, or a member of its group.
func assignable(ctx context.Context, t tasks.Task, uid string) (bool, error) {
	if uid == t.OwnerID || slices.Contains(t.SharedWith, uid) {
		return true, nil
	}
	role, err := roleIn(ctx, t.GroupID, uid)
	return role != "", err
}
