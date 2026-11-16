// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package task

import (
	"slices"
	"time"
)

// overdue is Task.Overdue's body: an open task with a due date behind now.
func (tx *Task) overdue(now time.Time) (r0 bool) {
	return tx.State == LifecycleOpen && !tx.Due.IsZero() && tx.Due.Before(now)
}

// seenBy is Task.SeenBy's body: the owner, the members it is shared with, its
// assignee, and a member of its group. The one rule of who sees a task.
func (tx *Task) seenBy(user string, inGroup bool) (r0 bool) {
	return user != "" && (user == tx.OwnerID || user == tx.AssigneeID || slices.Contains(tx.SharedWith, user) || inGroup && tx.GroupID != "")
}
