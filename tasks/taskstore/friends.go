// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package taskstore

import (
	"context"

	"github.com/kitsunium/todo/contacts/contactsapi"
)

// friends is Friends's own state: write it freely.
type friends struct {
}

// newFriends is NewFriends's body.
func newFriends() *Friends {
	return &Friends{friends{}}
}

// areContacts is Friends.AreContacts's body.
func (a *Friends) areContacts(ctx context.Context, one string, other string) (r0 bool, err error) {
	v, err := contactsapi.AreContactsQuery.Ask(ctx, contactsapi.Pair{One: one, Other: other})
	return v.Yes, err
}
