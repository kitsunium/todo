package notify

import (
	"bytes"
	"embed"
	htmltemplate "html/template"
	"net/url"
	"regexp"
	"strings"
	texttemplate "text/template"
	"unicode"
	"unicode/utf8"
)

// The design: one HTML layout — tables and inline styles, what every mail
// client renders, 560 pixels wide, light only — and its plain text twin.
// Every mail is a Message poured into both, in the Message's language: the
// few words the layouts add come from the same catalogue.
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
	// PasteLink introduces the button's link written out; Tagline signs the
	// footer; Colon is the language's colon — "\u00a0:" in French — for
	// the plain text part's "label: value" lines.
	PasteLink string
	Tagline   string
	Colon     string
}

// hexColor is the only shape a card color may have.
var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

// render pours m into the HTML layout and the text layout. User text — a
// name, a task title — is escaped by html/template in the HTML part. A
// message the catalogue could not write in full is refused.
func render(m Message) (html, text string, err error) {
	if m.err != nil {
		return "", "", m.err
	}
	w := wordsIn(m.Lang)
	home := BaseURL()
	p := page{Message: m, Home: home, Host: home,
		PasteLink: w.say("mail.paste_link"), Tagline: w.say("mail.tagline"), Colon: w.say("mail.colon")}
	if w.err != nil {
		return "", "", w.err
	}
	p.Lang = w.locale
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
// than the width — a link — stays whole, and so does what a no-break space
// holds together: French puts one before « : ? ! » and inside « ».
func wrap(s string) string {
	var b strings.Builder
	col := 0
	for _, word := range strings.FieldsFunc(s, breakable) {
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

// breakable reports whether a line may break at r: at a space, but not at
// a no-break space.
func breakable(r rune) bool {
	return unicode.IsSpace(r) && r != '\u00a0' && r != '\u202f'
}

// rule underlines a headline in the plain text part.
func rule(s string) string {
	return strings.Repeat("=", min(utf8.RuneCountInString(s), textWidth))
}
