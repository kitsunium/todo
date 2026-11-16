// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package link

// pairOf is PairOf's body: the two ids, the lesser first, joined by a
// character no id holds.
func pairOf(a string, b string) (r0 string) {
	if b < a {
		a, b = b, a
	}
	return a + "|" + b
}
