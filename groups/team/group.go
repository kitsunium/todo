// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package team

// memberIDs is Group.MemberIDs's body: one id a membership, in their order.
func (g *Group) memberIDs() (r0 []string) {
	ids := make([]string, len(g.Members))
	for i, m := range g.Members {
		ids[i] = m.UserID
	}
	return ids
}

// roleOf is Group.RoleOf's body: the role of the first membership of the
// member, empty when there is none.
func (g *Group) roleOf(user string) (r0 Role) {
	for _, m := range g.Members {
		if m.UserID == user {
			return m.Role
		}
	}
	return ""
}
