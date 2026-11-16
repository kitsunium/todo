// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package linkstore

import (
	"context"

	"github.com/kitsunium/todo/contacts/circle"
	"github.com/kitsunium/todo/identity/vault"
)

// directory is Directory's own state: it asks the identity, it holds nothing.
type directory struct {
}

// newDirectory is NewDirectory's body.
func newDirectory() *Directory {
	return &Directory{directory{}}
}

// byEmail is Directory.ByEmail's body: the identity's query "member-by-email".
func (a *Directory) byEmail(ctx context.Context, email string) (r0 circle.Person, r1 bool, err error) {
	m, err := vault.MemberByEmailQuery.Ask(ctx, vault.ByEmail{Email: email})
	return circle.Person{ID: m.ID, Email: m.Email, Name: m.Name}, m.Found, err
}

// byID is Directory.ByID's body: the identity's query "member".
func (a *Directory) byID(ctx context.Context, id string) (r0 circle.Person, r1 bool, err error) {
	m, err := vault.MemberQuery.Ask(ctx, vault.ByID{ID: id})
	return circle.Person{ID: m.ID, Email: m.Email, Name: m.Name}, m.Found, err
}
