// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package feed

import (
	"context"

	"github.com/kitsunium/sdk/framework/app"
)

// pageSize is how many entries a read answers when it asks for no number.
const pageSize = 50

// recent is GET /api/activity's body: the caller's newest entries, Limit of them at
// most, fifty when the request names no number.
func recent(ctx context.Context, in PageRequest) (out Page, err error) {
	uid, ok := app.UserID(ctx)
	if !ok {
		return Page{}, app.Unauthenticated("sign in first")
	}
	all, err := newestFirst(ctx)
	if err != nil {
		return Page{}, err
	}
	limit := in.Limit
	if limit <= 0 {
		limit = pageSize
	}
	out.Entries = []Entry{}
	for _, e := range all {
		if e.UserID == string(uid) && len(out.Entries) < limit {
			out.Entries = append(out.Entries, e)
		}
	}
	return out, nil
}
