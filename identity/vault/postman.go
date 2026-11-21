// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"
	"net/url"

	"github.com/kitsunium/todo/identity/account"
	"github.com/kitsunium/todo/notify/post"
)

// postman is Postman's own state: the words and the links are post's.
type postman struct {
}

// newPostman is NewPostman's body.
func newPostman() *Postman {
	return &Postman{postman{}}
}

// alreadyMember is Postman.AlreadyMember's body: someone signed up with an
// address that has an account, and its member is told in the language of
// that account.
func (a *Postman) alreadyMember(ctx context.Context, to account.Account) (err error) {
	return a.send(ctx, to, post.KindExists, "/forgot")
}

// reset is Postman.Reset's body: the link that chooses a new password.
func (a *Postman) reset(ctx context.Context, to account.Account, secret string) (err error) {
	return a.send(ctx, to, post.KindReset, "/reset?token="+url.QueryEscape(secret))
}

// verification is Postman.Verification's body: the link that proves the
// address.
func (a *Postman) verification(ctx context.Context, to account.Account, secret string) (err error) {
	return a.send(ctx, to, post.KindVerify, "/verify?token="+url.QueryEscape(secret))
}

// send hands a member's mail of a kind to the command send of post, in the
// language and the time zone of their account, with the path of its action.
func (a *Postman) send(ctx context.Context, to account.Account, kind, path string) error {
	_, err := post.SendCommand.Dispatch(ctx, post.Letter{
		To: to.Email, Name: to.Name, Locale: to.Locale, TimeZone: to.TimeZone, Kind: kind, Path: path,
	})
	return err
}
