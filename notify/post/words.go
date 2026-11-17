package post

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	// The zone database, inside the binary: the product's image has none.
	_ "time/tzdata"
)

// The words of the mails: one flat catalogue per language, named after it
// (locales/fr.json, locales/en.json), each message a text whose {name}
// placeholders are filled when it is said. French comes first: it is what a
// reader gets when nothing says otherwise.
//
//go:embed locales/*.json
var catalogueFiles embed.FS

// languages are the languages the product writes in, French first.
var languages = []string{"fr", "en"}

// catalogue is the words of one language: a message by its key.
type catalogue map[string]string

// catalogues are the words of every language. They are loaded when the
// process starts, which refuses to start on a catalogue that is missing, that
// is not a flat object of texts, or that lacks a key another has: a startup
// failure, never a mail with a hole in it.
var catalogues = func() map[string]catalogue {
	c, err := loadCatalogues(catalogueFiles, "locales")
	if err != nil {
		panic("post: the mail catalogues: " + err.Error())
	}
	return c
}()

// loadCatalogues reads the catalogue of each language from dir: exactly one
// file per language, every one with the keys of the first.
func loadCatalogues(fsys fs.FS, dir string) (map[string]catalogue, error) {
	out := make(map[string]catalogue, len(languages))
	for _, lang := range languages {
		raw, err := fs.ReadFile(fsys, path.Join(dir, lang+".json"))
		if err != nil {
			return nil, err
		}
		var c catalogue
		if err := json.Unmarshal(raw, &c); err != nil {
			return nil, fmt.Errorf("%s.json: %w", lang, err)
		}
		out[lang] = c
	}
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	if len(entries) != len(languages) {
		return nil, fmt.Errorf("%d files in %s for the %d languages %v", len(entries), dir, len(languages), languages)
	}
	first := out[languages[0]]
	for _, lang := range languages[1:] {
		if missing, extra := keysOnlyIn(first, out[lang]), keysOnlyIn(out[lang], first); len(missing)+len(extra) > 0 {
			return nil, fmt.Errorf("%s lacks %v and alone has %v: every language says everything", lang, missing, extra)
		}
	}
	return out, nil
}

// keysOnlyIn are the keys of a that b lacks, in order.
func keysOnlyIn(a, b catalogue) []string {
	var out []string
	for k := range a {
		if _, ok := b[k]; !ok {
			out = append(out, k)
		}
	}
	slices.Sort(out)
	return out
}

// words is one mail being written in one language, its times in one zone. It
// keeps the first message the catalogue could not give; the mail is refused
// rather than sent with a key in place of a sentence.
type words struct {
	locale string
	zone   *time.Location
	cat    catalogue
	err    error
}

// wordsIn starts writing a mail in locale, French when it is none the product
// speaks, with its times in zone, UTC when it is empty or unknown.
func wordsIn(locale, zone string) *words {
	if !slices.Contains(languages, locale) {
		locale = languages[0]
	}
	loc := time.UTC
	if zone != "" && zone != "Local" {
		if l, err := time.LoadLocation(zone); err == nil {
			loc = l
		}
	}
	return &words{locale: locale, zone: loc, cat: catalogues[locale]}
}

// say is the message key with its {name} placeholders filled from the
// name, value pairs.
func (w *words) say(key string, pairs ...string) string {
	text, ok := w.cat[key]
	if !ok {
		if w.err == nil {
			w.err = fmt.Errorf("post: no %s message %q", w.locale, key)
		}
		return key
	}
	if len(pairs) == 0 {
		return text
	}
	args := make([]string, 0, len(pairs))
	for i := 0; i+1 < len(pairs); i += 2 {
		args = append(args, "{"+pairs[i]+"}", pairs[i+1])
	}
	return strings.NewReplacer(args...).Replace(text)
}

// greeting opens a mail with the recipient's name: "Bonjour Ada,", "Hi Ada,",
// or no name when there is none.
func (w *words) greeting(name string) string {
	if name == "" {
		return w.say("mail.greeting_anonymous")
	}
	return w.say("mail.greeting", "name", name)
}

// date writes a moment the way the language does, in the reader's zone:
// "jeudi 24 septembre à 19:00 CEST", "Thursday, September 24 at 13:00 EDT".
func (w *words) date(t time.Time) string {
	t = t.In(w.zone)
	day := strconv.Itoa(t.Day())
	if t.Day() == 1 {
		day = w.say("date.first_of_month")
	}
	return w.say("date.long",
		"weekday", w.say("date.weekday."+strconv.Itoa(int(t.Weekday()))),
		"day", day,
		"month", w.say("date.month."+strconv.Itoa(int(t.Month()))),
		"time", t.Format("15:04"),
		"zone", t.Format("MST"))
}
