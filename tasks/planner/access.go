// What the use cases share about who may do what with a task: the one place
// the access rule is applied. Who sees a task is the domain's (Task.SeenBy);
// what this file adds is the role the caller has in the task's group, asked
// of the groups.

package planner

import (
	"context"
	"slices"

	"github.com/kitsunium/todo/tasks/task"
)

// The roles of a group that run its list: they delete its tasks.
const (
	roleOwner = "owner"
	roleAdmin = "admin"
)

// roleIn is the role a member has in the group a task is on: empty for a task
// on no group, and for a member who is in none.
func roleIn(ctx context.Context, groups GroupsV1, t task.Task, user string) (string, error) {
	if t.GroupID == "" || user == "" {
		return "", nil
	}
	return groups.RoleOf(ctx, t.GroupID, user)
}

// load reads a task for a member who sees it, with the role they have in its
// group; a task they do not see is not found, as if there were none.
func load(ctx context.Context, store TaskStoreV1, groups GroupsV1, id string, user string) (task.Task, string, error) {
	t, err := store.Get(ctx, id)
	if err != nil {
		return task.Task{}, "", err
	}
	role, err := roleIn(ctx, groups, t, user)
	if err != nil {
		return task.Task{}, "", err
	}
	if !t.SeenBy(user, role != "") {
		return task.Task{}, "", task.ErrNotFound
	}
	return t, role, nil
}

// change applies fn to a task a member sees, as the store holds it at that
// instant and alone among its writers: two changes made at once both stay.
// The member's role in the task's group is read first; fn is given it, and
// the task as it is now. A task that left that group meanwhile is judged
// with no role, and one the member no longer sees is not found.
func change(ctx context.Context, store TaskStoreV1, groups GroupsV1, id string, user string, fn func(t *task.Task, role string) error) (task.Task, error) {
	seen, role, err := load(ctx, store, groups, id, user)
	if err != nil {
		return task.Task{}, err
	}
	return store.Change(ctx, id, func(t *task.Task) error {
		now := role
		if t.GroupID != seen.GroupID {
			now = ""
		}
		if !t.SeenBy(user, now != "") {
			return task.ErrNotFound
		}
		return fn(t, now)
	})
}

// manages reports whether a member deletes a task: its owner, or an owner or
// an admin of its group, whose role is given.
func manages(t task.Task, user string, role string) bool {
	return user == t.OwnerID || role == roleOwner || role == roleAdmin
}

// assignable reports whether a task may be assigned to a member: its owner,
// a member it is shared with, or a member of its group, which the caller
// asked of the groups. Being a contact of whoever asks is not enough:
// assigning a task lets its assignee see it, and only its owner decides who
// else does, by sharing it.
func assignable(t task.Task, candidate string, inGroup bool) bool {
	return candidate == t.OwnerID || slices.Contains(t.SharedWith, candidate) || inGroup && t.GroupID != ""
}
