// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package link

// other is Link.Other's body.
func (l *Link) other(user string) (r0 string, r1 bool) {
	switch user {
	case l.RequesterID:
		return l.AddresseeID, true
	case l.AddresseeID:
		return l.RequesterID, true
	}
	return "", false
}
