// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package linkstore

import (
	"context"
	"strings"

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

// invite is Courier.Invite's body: the address nobody has is told, in
// French and as plain text, who asks to be its contact and that an account
// made with this address answers the request. The mail carries no link: the
// address of the web app is the member's to know.
func (a *Courier) invite(ctx context.Context, to string, by circle.Person) (err error) {
	who := strings.Join(strings.Fields(by.Name), " ")
	if who == "" {
		who = "Un membre"
	}
	body := "Bonjour,\n\n" + who + " souhaite vous compter parmi ses contacts sur la liste de tâches de l’équipe, mais cette adresse n’a pas encore de compte.\n\n" +
		"Créez votre compte avec cette adresse pour répondre à sa demande.\n\n" +
		"Si vous ne connaissez pas cette personne, ignorez ce message.\n"
	_, err = post.SendCommand.Dispatch(ctx, post.Letter{To: to, Subject: who + " vous invite à créer un compte", Body: body})
	return err
}
