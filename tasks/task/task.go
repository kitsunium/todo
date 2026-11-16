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

// seenBy is Task.SeenBy's body: the owner, and the members it is shared with.
func (tx *Task) seenBy(user string) (r0 bool) {
	return user != "" && (user == tx.OwnerID || slices.Contains(tx.SharedWith, user))
}
