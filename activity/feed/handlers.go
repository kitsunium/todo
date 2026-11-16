// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package feed

import (
	"context"
)

// pageSize is how many entries a read answers when it asks for no number.
const pageSize = 50

// recent is GET /api/activity's body: the newest entries, Limit of them at
// most, fifty when the request names no number.
func recent(ctx context.Context, in PageRequest) (out Page, err error) {
	all, err := newestFirst(ctx)
	if err != nil {
		return Page{}, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = pageSize
	}
	return Page{Entries: all[:min(limit, len(all))]}, nil
}
