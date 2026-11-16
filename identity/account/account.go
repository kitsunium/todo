// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package account

import (
	"time"
)

// locked is Account.Locked's body: the lock holds until LockedUntil, and an
// account that has none is never locked.
func (a *Account) locked(now time.Time) (r0 bool) {
	return !a.LockedUntil.IsZero() && now.Before(a.LockedUntil)
}
