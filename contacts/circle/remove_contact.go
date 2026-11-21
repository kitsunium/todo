// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"context"

	"github.com/kitsunium/todo/contacts/link"
)

// execute is RemoveContact's body: either member of a link deletes it;
// anyone else finds none.
func (h *RemoveContactHandler) execute(ctx context.Context, in *RemoveContactInput) (RemoveContactOutput, error) {
	l, found, err := h.links.Get(ctx, in.ID)
	if err != nil {
		return RemoveContactOutput{}, err
	}
	if _, mine := l.Other(in.User); !found || !mine {
		return RemoveContactOutput{}, link.ErrNotFound
	}
	return RemoveContactOutput{}, h.links.Delete(ctx, l.ID)
}
