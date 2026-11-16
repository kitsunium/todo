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

// manages reports whether a member deletes a task: its owner, or an owner or
// an admin of its group, whose role is given.
func manages(t task.Task, user string, role string) bool {
	return user == t.OwnerID || role == roleOwner || role == roleAdmin
}

// assignable reports whether a task is assigned to a member by the caller:
// to the caller, to its owner, to a member it is shared with, to a member of
// its group when it is on a group's list, to a contact of the caller when it
// is a personal task.
func assignable(ctx context.Context, groups GroupsV1, circle CircleV1, t task.Task, caller string, candidate string) (bool, error) {
	if candidate == caller || candidate == t.OwnerID || slices.Contains(t.SharedWith, candidate) {
		return true, nil
	}
	if t.GroupID != "" {
		role, err := groups.RoleOf(ctx, t.GroupID, candidate)
		return role != "", err
	}
	return circle.AreContacts(ctx, caller, candidate)
}
