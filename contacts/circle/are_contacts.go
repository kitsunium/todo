// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"context"

	"github.com/kitsunium/todo/contacts/link"
)

// execute is AreContacts's body: two members are contacts when the link of
// their pair is accepted.
func (h *AreContactsHandler) execute(ctx context.Context, in *AreContactsInput) (AreContactsOutput, error) {
	l, found, err := h.links.ByPair(ctx, link.PairOf(in.A, in.B))
	return AreContactsOutput{Yes: found && l.State == link.RequestsAccepted}, err
}
