// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package team

import (
	"cmp"
	"slices"
)

// ranked is Ranked's body: a copy, ordered by rank, then by when the member
// joined, then by id so the order is the same whatever the input's.
func ranked(members []Membership) (r0 []Membership) {
	out := slices.Clone(members)
	rank := map[Role]int{RoleOwner: 0, RoleAdmin: 1, RoleMember: 2}
	slices.SortFunc(out, func(a, b Membership) int {
		return cmp.Or(cmp.Compare(rank[a.Role], rank[b.Role]), a.JoinedAt.Compare(b.JoinedAt), cmp.Compare(a.UserID, b.UserID))
	})
	return out
}
