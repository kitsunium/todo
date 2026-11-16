// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"

	"github.com/kitsunium/todo/identity/account"
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

// lookup is the query "member-by-email"'s body: whose account an address
// is, read as the product keeps addresses, and Found false when it is
// nobody's or no address at all.
func lookup(ctx context.Context, in ByEmail) (out Member, err error) {
	email, err := account.NormalizeEmail(in.Email)
	if err != nil {
		return Member{}, nil
	}
	found, err := AccountsStore.Find(ctx, "by-email", email)
	if err != nil || len(found) == 0 {
		return Member{}, err
	}
	return Member{Found: true, ID: found[0].ID, Email: found[0].Email, Name: found[0].Name}, nil
}
