// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package team

// parseRole is ParseRole's body: admin or member, which are the roles a
// member is given; the owner's is the creator's and never given.
func parseRole(s string) (r0 Role, err error) {
	switch Role(s) {
	case RoleAdmin:
		return RoleAdmin, nil
	case RoleMember:
		return RoleMember, nil
	}
	return "", ErrBadRole
}
