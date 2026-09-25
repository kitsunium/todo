package notify

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

	"github.com/kitsunium/sdk/pkg/v1/i18n"
	"github.com/kitsunium/todo/internal/wire"
)

// The mails' words: one catalogue per language, named after it —
// locales/fr.json, locales/en.json — in the shape the SDK's i18n.LoadFS
// reads: one flat object whose members are a message, "Bonjour {name},", or
// the CLDR plural forms of a counted one. They are decoded with
// encoding/json and compiled by i18n.NewStore, LoadFS's own last step:
// LoadFS decodes through the SDK's codec registry, whose one switch links
// all of its formats — BSON, CBOR, YAML… — into the binary.
//
//go:embed locales/*.json
var catalogueFiles embed.FS

// printers render the mails' words, one per language the product speaks.
// They are built when the process starts, which refuses to start on a
// catalogue the SDK rejects — a counted message missing a plural form its
// language needs, a malformed pattern — or on a key one language has and
// another lacks: a startup failure, never a mail with a hole in it.
var printers = func() map[wire.Locale]*i18n.Printer {
	p, err := loadPrinters(catalogueFiles, "locales")
	if err != nil {
		panic("notify: the mail catalogues: " + err.Error())
	}
	return p
}()

// loadPrinters compiles the catalogues of dir into one printer per language
// the product speaks.
func loadPrinters(fsys fs.FS, dir string) (map[wire.Locale]*i18n.Printer, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	catalogues := map[i18n.Tag]i18n.Catalogue{}
	for _, e := range entries {
		if e.IsDir() || path.Ext(e.Name()) != ".json" {
			continue
		}
		tag, err := i18n.ParseTag(strings.TrimSuffix(e.Name(), ".json"))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if _, twice := catalogues[tag]; twice {
			return nil, fmt.Errorf("%s: a second catalogue for %s", e.Name(), tag)
		}
		raw, err := fs.ReadFile(fsys, path.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		if catalogues[tag], err = decodeCatalogue(raw); err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
	}
	tags := make([]i18n.Tag, len(wire.Locales))
	for i, l := range wire.Locales {
		if tags[i], err = i18n.ParseTag(string(l)); err != nil {
			return nil, err
		}
	}
	if len(catalogues) != len(tags) {
		return nil, fmt.Errorf("%d catalogues for the %d languages %v", len(catalogues), len(tags), wire.Locales)
	}
	store, err := i18n.NewStore(tags[0], catalogues)
	if err != nil {
		return nil, err
	}
	out := make(map[wire.Locale]*i18n.Printer, len(tags))
	for i, tag := range tags {
		if store.Keys(tag) == nil {
			return nil, fmt.Errorf("no catalogue for %s", tag)
		}
		if missing, extra := store.Missing(tag), onlyIn(store.Keys(tag), store.Keys(tags[0])); len(missing)+len(extra) > 0 {
			return nil, fmt.Errorf("%s lacks %v and alone has %v: every language says everything", tag, missing, extra)
		}
		if out[wire.Locales[i]], err = i18n.NewPrinter(store, tag); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// decodeCatalogue reads one catalogue file: a message is a string, a counted
// message an object of plural forms; anything else is refused.
func decodeCatalogue(raw []byte) (i18n.Catalogue, error) {
	var members map[string]any
	if err := json.Unmarshal(raw, &members); err != nil {
		return nil, err
	}
	out := make(i18n.Catalogue, len(members))
	for key, v := range members {
		switch v := v.(type) {
		case string:
			out[i18n.Key(key)] = i18n.Plain(v)
		case map[string]any:
			forms := make(map[string]string, len(v))
			for form, text := range v {
				s, ok := text.(string)
				if !ok {
					return nil, fmt.Errorf("the %q form of %q is not a text", form, key)
				}
				forms[form] = s
			}
			out[i18n.Key(key)] = i18n.PluralForms(forms)
		default:
			return nil, fmt.Errorf("%q is neither a text nor its plural forms", key)
		}
	}
	return out, nil
}

// onlyIn returns the keys of have that want lacks.
func onlyIn(have, want []i18n.Key) []i18n.Key {
	var out []i18n.Key
	for _, k := range have {
		if !slices.Contains(want, k) {
			out = append(out, k)
		}
	}
	return out
}

// words is one mail being written in one language. It keeps the first
// message the catalogue could not render; the mail carries that error to
// Deliver, which does not send a mail with a key in place of a sentence.
type words struct {
	locale wire.Locale
	p      *i18n.Printer
	err    error
}

// wordsIn starts writing a mail in l — in French when l is not a language
// the product speaks.
func wordsIn(l wire.Locale) *words {
	l = l.Resolve()
	return &words{locale: l, p: printers[l]}
}

// say renders the message key, its arguments given as name, value pairs.
func (w *words) say(key string, args ...string) string {
	s, err := w.p.Render(i18n.Key(key), argsOf(args))
	w.keep(key, err)
	return s
}

// count renders the counted message key for n, which its pattern names {n}.
func (w *words) count(key string, n int, args ...string) string {
	s, err := w.p.RenderCount(i18n.Key(key), i18n.Int(int64(n)), argsOf(append(args, "n", strconv.Itoa(n))))
	w.keep(key, err)
	return s
}

// keep records the first failure.
func (w *words) keep(key string, err error) {
	if err != nil && w.err == nil {
		w.err = fmt.Errorf("notify: the %s message %q: %w", w.locale, key, err)
	}
}

// done stamps m with its language and whatever failed while writing it.
func (w *words) done(m Message) Message {
	m.Lang, m.err = w.locale, w.err
	return m
}

// argsOf turns name, value pairs into the SDK's arguments.
func argsOf(pairs []string) i18n.Args {
	args := make(i18n.Args, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		args[pairs[i]] = pairs[i+1]
	}
	return args
}

// greeting opens a mail with the recipient's first name: "Bonjour Alice,",
// "Hi Alice,", or no name when there is none.
func (w *words) greeting(name string) string {
	first, _, _ := strings.Cut(strings.TrimSpace(name), " ")
	if first == "" {
		return w.say("mail.greeting_anonymous")
	}
	return w.say("mail.greeting", "name", first)
}

// someone is a person's name, or "Quelqu’un" when the directory no longer
// knows them.
func (w *words) someone(name string) string {
	if strings.TrimSpace(name) == "" {
		return w.say("mail.someone")
	}
	return name
}

// date writes a moment the way the language does — "jeudi 24 septembre à
// 17:00 UTC", "Thursday, September 24 at 17:00 UTC" — in UTC: the product
// does not know its users' time zones.
func (w *words) date(t time.Time) string {
	t = t.UTC()
	day := strconv.Itoa(t.Day())
	if t.Day() == 1 {
		day = w.say("date.first_of_month")
	}
	return w.say("date.long",
		"weekday", w.say("date.weekday."+strconv.Itoa(int(t.Weekday()))),
		"day", day,
		"month", w.say("date.month."+strconv.Itoa(int(t.Month()))),
		"time", t.Format("15:04"))
}
