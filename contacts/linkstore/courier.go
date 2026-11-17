// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package linkstore

import (
	"context"

	"github.com/kitsunium/todo/contacts/circle"
	"github.com/kitsunium/todo/notify/post"
)

// courier is Courier's own state: the mail is the product's, it holds nothing.
type courier struct {
}

// newCourier is NewCourier's body.
func newCourier() *Courier {
	return &Courier{courier{}}
}

// invite is Courier.Invite's body: the address nobody has is told who asks to
// be its contact and that an account made with this address answers the
// request, in the language of the member who asks, since the address has none.
// The mail links to the web app's root, where the account is made.
func (a *Courier) invite(ctx context.Context, to string, by circle.Person) (err error) {
	_, err = post.SendCommand.Dispatch(ctx, post.Letter{
		To: to, Locale: by.Locale, Kind: post.KindInvite, Path: "/", Params: map[string]string{"actor": by.Name},
	})
	return err
}
