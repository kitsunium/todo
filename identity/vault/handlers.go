// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"
)

// directory is the query "member"'s body: a member's address and name by
// their id, and Found false when the account is gone. It filters instead of
// reading the key: nobody there is an answer, never a failed read.
func directory(ctx context.Context, in ByID) (out Member, err error) {
	found, err := AccountsStore.Filter(ctx, func(r AccountRecord) bool { return r.ID == in.ID })
	if err != nil || len(found) == 0 {
		return Member{}, err
	}
	return Member{Found: true, ID: found[0].ID, Email: found[0].Email, Name: found[0].Name}, nil
}
