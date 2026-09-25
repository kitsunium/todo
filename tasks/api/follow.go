package api

import (
	"context"
	"slices"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/groups"
	"github.com/kitsunium/todo/internal/wire"
	"github.com/kitsunium/todo/tasks"
)

// The task list follows its groups: a deleted group's tasks go back to their
// owners' lists, and a member who leaves stops doing the group's tasks.
//
// fr: La liste de tâches suit ses groupes : les tâches d’un groupe supprimé
// retournent sur les listes de leurs propriétaires, et un membre qui part cesse
// de faire les tâches du groupe.
var _ = tasks.Service.Subscribe("group-changes", groups.Events, FollowGroups)

// FollowGroups keeps tasks consistent with the groups they are on. It is
// idempotent, as a subscription must be: applying it twice changes nothing
// the first time did not.
//
// fr: FollowGroups garde les tâches cohérentes avec les groupes où elles
// figurent. Comme tout abonnement, il est idempotent : l’appliquer deux fois ne
// change rien que la première fois n’ait déjà changé.
func FollowGroups(ctx context.Context, e groups.Event) error {
	switch e.Kind {
	case groups.KindDeleted, groups.KindLeft, groups.KindRemoved:
	default:
		return nil
	}
	// A deleted group's tasks all change; a departure changes only the tasks
	// of the group the member who left was doing.
	listed, err := tasks.Tasks.Find(ctx, "assignee", e.UserID)
	if e.Kind == groups.KindDeleted {
		listed, err = tasks.Tasks.Find(ctx, "group", e.GroupID)
	}
	if err != nil {
		return err
	}
	now := wire.Now(ctx)
	for _, t := range listed {
		if t.GroupID != e.GroupID {
			continue
		}
		_, err := tasks.Tasks.Update(ctx, t.ID, func(cur *tasks.Task) error {
			if cur.GroupID != e.GroupID {
				return nil
			}
			if e.Kind == groups.KindDeleted {
				cur.GroupID = ""
			}
			// An assignee who no longer sees the task stops doing it.
			if cur.AssigneeID != "" && cur.AssigneeID != cur.OwnerID && !slices.Contains(cur.SharedWith, cur.AssigneeID) &&
				(e.Kind == groups.KindDeleted || cur.AssigneeID == e.UserID) {
				cur.AssigneeID = ""
			}
			cur.UpdatedAt = now
			return nil
		})
		if err != nil && !wire.Is(err, kit.CodeNotFound) {
			return err
		}
	}
	return nil
}
