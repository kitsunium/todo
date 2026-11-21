// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package planner

import (
	"testing"
	"time"
)

// givenFollowGroupSendsTheTasksOfADeletedGroupBackToTheirOwners builds FollowGroup for "sends the tasks of a deleted group back to their owners": the use case, its input, and
// what its output must be.
func givenFollowGroupSendsTheTasksOfADeletedGroupBackToTheirOwners(t *testing.T) (FollowGroup, FollowGroupInput, func(*testing.T, FollowGroupOutput)) {
	t.Helper()
	mine, theirs, elsewhere := grouped("task_mine"), grouped("task_theirs"), grouped("task_elsewhere")
	mine.AssigneeID = ada // its owner keeps doing it
	theirs.AssigneeID = cy
	elsewhere.GroupID = "group_other"
	elsewhere.AssigneeID = cy
	store := newMemTasks(mine, theirs, elsewhere)
	later := epoch.Add(time.Hour)
	uc := NewFollowGroup(store, memGroups{}, fixedClock(later))
	return uc, FollowGroupInput{Group: core, Deleted: true}, func(t *testing.T, _ FollowGroupOutput) {
		t.Helper()
		if got := store.byID["task_mine"]; got.GroupID != "" || got.AssigneeID != ada || !got.UpdatedAt.Equal(later) {
			t.Errorf("the task of its owner: %+v, want it on no group, still assigned to Ada", got)
		}
		if got := store.byID["task_theirs"]; got.GroupID != "" || got.AssigneeID != "" {
			t.Errorf("the task assigned to a member of the group: %+v, want it on no group and unassigned", got)
		}
		if got := store.byID["task_elsewhere"]; got.GroupID != "group_other" || got.AssigneeID != cy || !got.UpdatedAt.Equal(epoch) {
			t.Errorf("the task of another group: %+v, want it untouched", got)
		}
	}
}

// givenFollowGroupUnassignsWhoLeftTheGroup builds FollowGroup for "unassigns who left the group": the use case, its input, and
// what its output must be.
func givenFollowGroupUnassignsWhoLeftTheGroup(t *testing.T) (FollowGroup, FollowGroupInput, func(*testing.T, FollowGroupOutput)) {
	t.Helper()
	left, stays, kept := grouped("task_left"), grouped("task_stays"), grouped("task_kept")
	left.AssigneeID = dee // no longer in the group
	stays.AssigneeID = bob
	kept.AssigneeID, kept.SharedWith = dee, []string{dee} // still sees it, shared with
	store := newMemTasks(left, stays, kept)
	later := epoch.Add(time.Hour)
	uc := NewFollowGroup(store, crew(), fixedClock(later))
	return uc, FollowGroupInput{Group: core}, func(t *testing.T, _ FollowGroupOutput) {
		t.Helper()
		if got := store.byID["task_left"]; got.AssigneeID != "" || got.GroupID != core || !got.UpdatedAt.Equal(later) {
			t.Errorf("the task of who left: %+v, want it still on the group, unassigned", got)
		}
		if got := store.byID["task_stays"]; got.AssigneeID != bob || !got.UpdatedAt.Equal(epoch) {
			t.Errorf("the task of a member who stays: %+v, want it untouched", got)
		}
		if got := store.byID["task_kept"]; got.AssigneeID != dee || !got.UpdatedAt.Equal(epoch) {
			t.Errorf("the task of who it is shared with: %+v, want it untouched", got)
		}
		again := NewFollowGroup(store, crew(), fixedClock(later.Add(time.Hour)))
		if _, err := again.Execute(t.Context(), FollowGroupInput{Group: core}); err != nil || !store.byID["task_left"].UpdatedAt.Equal(later) {
			t.Errorf("heard twice: %v, %+v, want nothing changed the second time", err, store.byID["task_left"])
		}
	}
}
