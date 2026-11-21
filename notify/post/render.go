package post

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"net/url"
	"strings"
	texttemplate "text/template"
	"unicode"
	"unicode/utf8"
)

// The design: one HTML layout, tables and inline styles, what every mail
// client renders, 560 pixels wide, light only, and its plain text twin. Every
// mail is a message poured into both, in the message's language: the few
// words the layouts add come from the same catalogue.
//
//go:embed templates/mail.html templates/mail.txt
var layouts embed.FS

var (
	htmlLayout = htmltemplate.Must(htmltemplate.ParseFS(layouts, "templates/mail.html"))
	textLayout = texttemplate.Must(texttemplate.New("mail.txt").
			Funcs(texttemplate.FuncMap{"wrap": wrap, "rule": rule}).
			ParseFS(layouts, "templates/mail.txt"))
)

// message is the content of one mail before the design is applied: one
// headline, at most one action, and a footer that says why it was sent.
type message struct {
	Subject   string
	Preheader string
	Headline  string
	// Greeting opens the body; empty for none.
	Greeting string
	// Body is the paragraphs before the action.
	Body   []string
	Action *action
	// Notes are the small print under the action.
	Notes  []string
	Reason string
}

// action is the one button of a mail.
type action struct {
	Label string
	URL   string
}

// page is what the layouts render: the message, and what the design adds.
type page struct {
	message
	Lang string
	// Home is the product's address, Host its host name, for the footer.
	Home string
	Host string
	// PasteLink introduces the button's link written out, Tagline signs the
	// footer, Colon is the language's colon for the plain text part.
	PasteLink string
	Tagline   string
	Colon     string
}

// render pours a message into the HTML layout and the text layout. What
// comes from a person (a name, a task's title) is escaped by html/template in
// the HTML part, and is only text in the other.
func render(m message, w *words, home string) (html, text string, err error) {
	p := page{message: m, Lang: w.locale, Home: home,
		PasteLink: w.say("mail.paste_link"), Tagline: w.say("mail.tagline"), Colon: w.say("mail.colon")}
	if w.err != nil {
		return "", "", w.err
	}
	p.Host = home
	if u, err := url.Parse(home); err == nil && u.Host != "" {
		p.Host = u.Host
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

// link is the address of a page of the web app: the base address and the
// path, which starts with a slash and names no other host.
func link(base, path string) (string, error) {
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, " \t\r\n\\") {
		return "", fmt.Errorf("post: the action path %q is not a path of the web app", path)
	}
	u, err := url.Parse(strings.TrimRight(base, "/") + path)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("post: the base address %q is not an address", base)
	}
	return u.String(), nil
}

// textWidth is where the plain text part wraps its paragraphs.
const textWidth = 72

// wrap folds a paragraph at textWidth columns, between words. A word longer
// than the width, a link, stays whole, and so does what a no-break space
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

// breakable reports whether a line may break at r: at a space, but not at a
// no-break space.
func breakable(r rune) bool {
	return unicode.IsSpace(r) && r != ' ' && r != ' '
}

// rule underlines a headline in the plain text part.
func rule(s string) string {
	return strings.Repeat("=", min(utf8.RuneCountInString(s), textWidth))
}
