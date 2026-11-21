// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package task

// lifecycle is Lifecycle's own state: write it freely.
type lifecycle struct{}

// newLifecycle is NewLifecycle's body.
func newLifecycle() *Lifecycle {
	return &Lifecycle{}
}
