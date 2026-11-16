// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package groupstore

import (
	"context"

	"github.com/kitsunium/todo/groups/roster"
	"github.com/kitsunium/todo/identity/vault"
)

// directory is Directory's own state: it asks the identity, it holds nothing.
type directory struct {
}

// newDirectory is NewDirectory's body.
func newDirectory() *Directory {
	return &Directory{directory{}}
}

// byID is Directory.ByID's body: the identity's query "member".
func (a *Directory) byID(ctx context.Context, id string) (r0 roster.Person, r1 bool, err error) {
	m, err := vault.MemberQuery.Ask(ctx, vault.ByID{ID: id})
	return roster.Person{ID: m.ID, Name: m.Name}, m.Found, err
}
