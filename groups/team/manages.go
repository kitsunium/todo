// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package team

// manages is Manages's body: the owner and the admins run the group.
func manages(role Role) (r0 bool) {
	return role == RoleOwner || role == RoleAdmin
}
