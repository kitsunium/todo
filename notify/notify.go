// Package notify is every mail the product sends: one outbox, one design,
// and the mails themselves, each in its recipient's language — French, or
// English (locales/). identity sends its transactional mails through
// SendAPI, a synchronous call, so a one-time link never travels on a topic.
// The rules that decide who hears about what — the subscriptions and the
// reminders loop — live in notify/dispatch, on this same service: they listen
// to services that depend on identity, which depends on this package.
package notify

import (
	"context"
	"net/url"
	"os"
	"strings"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/sdk/pkg/v1/mail"
	"github.com/kitsunium/todo/internal/wire"
)

// Service owns every mail the product sends.
var Service = kit.NewService("notify", "Every mail the product sends: the outbox, one design, the transactional mails and the notifications.\n\nfr: Tous les mails qu’envoie le produit : l’outbox, un seul design, les mails transactionnels et les notifications.")

// Mail is the product's outbound mail. Every message waits in its durable
// outbox until the transport accepts it: SMTP when KIT_SMTP_URL is set, and
// otherwise the capture transport, whose messages the Studio's mailbox shows.
//
// fr: Mail est le courrier sortant du produit. Chaque message attend dans son
// outbox durable jusqu’à ce que le transport l’accepte : SMTP quand
// KIT_SMTP_URL est défini, sinon le transport de capture, dont la boîte mail du
// Studio montre les messages.
var Mail = Service.Mailer("mail", kit.From("Todo", "hello@todo.localhost"))

// SendAPI sends one of identity's transactional mails — the verification,
// the password reset, the "you already have an account" — on its behalf.
//
// fr: SendAPI envoie l’un des mails transactionnels du service identity, en son
// nom — la vérification, la réinitialisation du mot de passe, le « vous avez
// déjà un compte ».
var SendAPI = Service.Endpoint("POST /internal/notify/send", Send, kit.Private())

// The transactional templates SendAPI renders.
const (
	TemplateVerifyEmail   = "verify-email"
	TemplatePasswordReset = "reset-password"
	TemplateAccountExists = "account-exists"
)

// SendInput is one transactional mail to send.
type SendInput struct {
	// Template is verify-email, reset-password or account-exists.
	Template string `json:"template" validate:"required,oneof=verify-email|reset-password|account-exists"`
	// To is the recipient's address.
	To string `json:"to" validate:"required,maxlen=254"`
	// Name is the recipient's name, for the greeting.
	Name string `json:"name" validate:"maxlen=80"`
	// Locale is the language the recipient reads: fr or en.
	Locale wire.Locale `json:"locale" validate:"required,oneof=fr|en"`
	// Data is what the template needs: the one-time token of its link. It is
	// a secret, redacted wherever kit shows a payload.
	Data map[string]string `json:"data,omitempty" kit:"secret"`
}

// SendOutput identifies the queued mail.
type SendOutput struct {
	// ID is the mail's ID in the outbox.
	ID string `json:"id"`
}

// Send renders a transactional mail in the product's design and the
// recipient's language, and puts it in the outbox. It returns once the mail is safely queued; the mailer's own
// loop delivers it, retrying a failure.
//
// fr: Send met en forme un mail transactionnel dans le design du produit et la
// langue du destinataire, puis le dépose dans l’outbox. Il rend la main dès que
// le mail est à l’abri dans la file ; la boucle propre au mailer le livre, et
// réessaie en cas d’échec.
func Send(ctx context.Context, in SendInput) (SendOutput, error) {
	var m Message
	switch in.Template {
	case TemplateVerifyEmail:
		m = VerifyEmail(in.Locale, in.Name, in.Data["token"])
	case TemplatePasswordReset:
		m = ResetPassword(in.Locale, in.Name, in.Data["token"])
	case TemplateAccountExists:
		m = AccountExists(in.Locale, in.Name)
	default:
		return SendOutput{}, kit.Invalid("unknown template")
	}
	id, err := Deliver(ctx, Recipient{Name: in.Name, Email: in.To}, m)
	return SendOutput{ID: id}, err
}

// Recipient is who a mail goes to.
type Recipient struct {
	Name  string
	Email string
}

// Deliver renders m in the product's design — an HTML part and its plain
// text alternative — and puts it in the outbox. It is the one path every
// mail takes, so every mail looks the same and the diagram shows each sender
// reaching the mailer.
func Deliver(ctx context.Context, to Recipient, m Message) (string, error) {
	html, text, err := render(m)
	if err != nil {
		return "", err
	}
	return Mail.Send(ctx, mail.Message{
		To:      []mail.Address{{Name: to.Name, Addr: to.Email}},
		Subject: m.Subject,
		Text:    text,
		HTML:    html,
	})
}

// BaseURL is where users reach the product, for the links of its mails:
// TODO_BASE_URL, or http://localhost:4000.
func BaseURL() string {
	if u := strings.TrimRight(strings.TrimSpace(os.Getenv("TODO_BASE_URL")), "/"); u != "" {
		return u
	}
	return "http://localhost:4000"
}

// link returns the absolute URL of a page of the web app, with its query
// parameters given as name, value pairs.
func link(path string, query ...string) string {
	u := BaseURL() + path
	if len(query) >= 2 {
		q := url.Values{}
		for i := 0; i+1 < len(query); i += 2 {
			q.Set(query[i], query[i+1])
		}
		u += "?" + q.Encode()
	}
	return u
}
