package notify

import (
	"bytes"
	"embed"
	htmltemplate "html/template"
	"net/url"
	"regexp"
	"strings"
	texttemplate "text/template"
	"unicode/utf8"
)

// The design: one HTML layout — tables and inline styles, what every mail
// client renders, 560 pixels wide, light and dark — and its plain text
// twin. Every mail is a Message poured into both.
//
//go:embed templates/mail.html templates/mail.txt
var layouts embed.FS

var (
	htmlLayout = htmltemplate.Must(htmltemplate.ParseFS(layouts, "templates/mail.html"))
	textLayout = texttemplate.Must(texttemplate.New("mail.txt").
			Funcs(texttemplate.FuncMap{"wrap": wrap, "rule": rule}).
			ParseFS(layouts, "templates/mail.txt"))
)

// page is what the layouts render: the message, and what the design adds.
type page struct {
	Message
	// CardColor is the card's edge, checked to be a color: a template may
	// only put trusted text in a style attribute.
	CardColor htmltemplate.CSS
	// Home is the product's address; Host its host name, for the footer.
	Home string
	Host string
}

// hexColor is the only shape a card color may have.
var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// render pours m into the HTML layout and the text layout. User text — a
// name, a task title — is escaped by html/template in the HTML part.
func render(m Message) (html, text string, err error) {
	home := BaseURL()
	p := page{Message: m, Home: home, Host: home}
	if u, err := url.Parse(home); err == nil && u.Host != "" {
		p.Host = u.Host
	}
	if m.Card != nil {
		color := m.Card.Color
		if !hexColor.MatchString(color) {
			color = accent
		}
		p.CardColor = htmltemplate.CSS(color)
	}
	var hb, tb bytes.Buffer
	if err := htmlLayout.Execute(&hb, p); err != nil {
		return "", "", err
	}
	if err := textLayout.Execute(&tb, p); err != nil {
		return "", "", err
	}
	return hb.String(), strings.TrimLeft(tb.String(), "\n"), nil
}

// textWidth is where the plain text part wraps its paragraphs.
const textWidth = 72

// wrap folds a paragraph at textWidth columns, between words. A word longer
// than the width — a link — stays whole.
func wrap(s string) string {
	var b strings.Builder
	col := 0
	for _, word := range strings.Fields(s) {
		n := utf8.RuneCountInString(word)
		switch {
		case col == 0:
		case col+1+n > textWidth:
			b.WriteByte('\n')
			col = 0
		default:
			b.WriteByte(' ')
			col++
		}
		b.WriteString(word)
		col += n
	}
	return b.String()
}

// rule underlines a headline in the plain text part.
func rule(s string) string {
	return strings.Repeat("=", min(utf8.RuneCountInString(s), textWidth))
}
