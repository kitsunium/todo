// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package taskstore

import (
	"context"

	"github.com/kitsunium/todo/groups/groupstore"
	"github.com/kitsunium/todo/tasks/planner"
)

// The kinds of a group's event the tasks follow; the groups' topic tells of
// others, an invitation or a role, that change no task.
const (
	groupDeleted = "group-deleted"
	groupLeft    = "group-left"
	groupRemoved = "group-removed"
)

// followGroups is the subscription "groups"'s body: a group deleted, or a
// member who left it or was removed from it, and the tasks follow. The group
// tells each of its members, so the message comes once for each, and the use
// case changes nothing the second time.
func followGroups(ctx context.Context, msg groupstore.Event) error {
	switch msg.Kind {
	case groupDeleted, groupLeft, groupRemoved:
		_, err := wired.followGroup.Execute(ctx, planner.FollowGroupInput{Group: msg.GroupID, Deleted: msg.Kind == groupDeleted})
		return err
	}
	return nil
}
