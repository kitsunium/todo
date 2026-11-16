// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package link

// requests is Requests's own state: write it freely.
type requests struct{}

// newRequests is NewRequests's body.
func newRequests() *Requests {
	return &Requests{}
}
