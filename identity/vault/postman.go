// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"
	"net/url"
	"strings"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/identity/account"
	"github.com/kitsunium/todo/notify/post"
)

// postman is Postman's own state: where members reach the product.
type postman struct {
	baseUrlSetting *app.Setting[string]
}

// newPostman is NewPostman's body.
func newPostman(baseUrlSetting *app.Setting[string]) *Postman {
	return &Postman{postman{baseUrlSetting: baseUrlSetting}}
}

// alreadyMember is Postman.AlreadyMember's body: someone signed up with an
// address that has an account, and its member is told how to get back in.
func (a *Postman) alreadyMember(ctx context.Context, to account.Account) (err error) {
	return a.write(ctx, to, "Vous avez déjà un compte",
		"Quelqu’un, peut-être vous, a voulu créer un compte avec cette adresse, qui en a déjà un.",
		"Si vous avez oublié votre mot de passe, choisissez-en un nouveau ici :",
		a.page("/forgot", ""),
		"Si ce n’était pas vous, ignorez ce message : rien n’a changé.")
}

// reset is Postman.Reset's body: the link that chooses a new password.
func (a *Postman) reset(ctx context.Context, to account.Account, secret string) (err error) {
	return a.write(ctx, to, "Choisissez un nouveau mot de passe",
		"Pour choisir un nouveau mot de passe, ouvrez ce lien, valable une heure :",
		a.page("/reset", secret),
		"Si vous n’avez rien demandé, ignorez ce message : votre mot de passe ne change pas.")
}

// verification is Postman.Verification's body: the link that proves the
// address.
func (a *Postman) verification(ctx context.Context, to account.Account, secret string) (err error) {
	return a.write(ctx, to, "Confirmez votre adresse",
		"Bienvenue ! Pour confirmer votre adresse et vous connecter, ouvrez ce lien, valable 24 heures :",
		a.page("/verify", secret),
		"Si vous n’avez pas créé de compte, ignorez ce message.")
}

// page is the address of a page of the web app, with a link's secret when
// it carries one.
func (a *Postman) page(path, secret string) string {
	u := strings.TrimRight(a.baseUrlSetting.Get(), "/") + path
	if secret != "" {
		u += "?token=" + url.QueryEscape(secret)
	}
	return u
}

// write sends a member a mail of their account through the product's
// mailer: a greeting, then its paragraphs.
func (a *Postman) write(ctx context.Context, to account.Account, subject string, paragraphs ...string) error {
	body := "Bonjour " + to.Name + ",\n\n" + strings.Join(paragraphs, "\n\n") + "\n"
	_, err := post.SendCommand.Dispatch(ctx, post.Letter{To: to.Email, Name: to.Name, Subject: subject, Body: body})
	return err
}
