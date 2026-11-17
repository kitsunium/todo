package post

import (
	"maps"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"
)

const home = "https://todo.example.com"

// every mail the product writes, as a letter of a language.
func letters(locale string) []Letter {
	base := Letter{To: "ada@example.com", Name: "Ada Lovelace", Locale: locale, TimeZone: "Europe/Paris"}
	with := func(kind, path string, params map[string]string) Letter {
		l := base
		l.Kind, l.Path, l.Params = kind, path, params
		return l
	}
	return []Letter{
		with(KindVerify, "/verify?token=abc_123-x", nil),
		with(KindReset, "/reset?token=abc_123-x", nil),
		with(KindExists, "/forgot", nil),
		with(KindInvite, "/", map[string]string{"actor": "Grace Hopper"}),
		with(KindDueSoon, "/", map[string]string{"title": "Ship the release", "due": "2026-11-16T14:30:00Z"}),
	}
}

// The two languages say the same things: the same keys.
func TestTheCataloguesHaveTheSameKeys(t *testing.T) {
	fr, en := catalogues["fr"], catalogues["en"]
	if len(fr) == 0 {
		t.Fatal("no French catalogue")
	}
	if a, b := slices.Sorted(maps.Keys(fr)), slices.Sorted(maps.Keys(en)); !slices.Equal(a, b) {
		t.Errorf("French has %v, English has %v", keysOnlyIn(fr, en), keysOnlyIn(en, fr))
	}
}

// A catalogue that lacks a key another has, or a language, is refused when
// the process starts.
func TestACatalogueThatLacksAKeyIsRefused(t *testing.T) {
	files := func(fr, en string) fstest.MapFS {
		return fstest.MapFS{"locales/fr.json": {Data: []byte(fr)}, "locales/en.json": {Data: []byte(en)}}
	}
	if _, err := loadCatalogues(files(`{"a":"un","b":"deux"}`, `{"a":"one","b":"two"}`), "locales"); err != nil {
		t.Errorf("two whole catalogues: %v", err)
	}
	if _, err := loadCatalogues(files(`{"a":"un","b":"deux"}`, `{"a":"one"}`), "locales"); err == nil {
		t.Error("a catalogue that lacks a key was taken")
	}
	if _, err := loadCatalogues(files(`{"a":"un"}`, `{"a":"one","c":"three"}`), "locales"); err == nil {
		t.Error("a catalogue with a key of its own was taken")
	}
	if _, err := loadCatalogues(files(`{"a":"un"}`, `{"a":["one"]}`), "locales"); err == nil {
		t.Error("a catalogue that is not flat texts was taken")
	}
	if _, err := loadCatalogues(fstest.MapFS{"locales/fr.json": {Data: []byte(`{}`)}}, "locales"); err == nil {
		t.Error("a missing language was taken")
	}
}

// Every mail, in each language, is a whole page with a plain text twin: no
// placeholder or key left in it, no tag in the text, the link built from the
// base address.
func TestEveryMailIsWrittenInBothLanguages(t *testing.T) {
	tag := regexp.MustCompile(`</?[a-z][^>]*>`)
	for _, locale := range []string{"fr", "en"} {
		for _, l := range letters(locale) {
			m, html, text, err := write(l, home)
			if err != nil {
				t.Fatalf("%s %s: %v", locale, l.Kind, err)
			}
			for _, part := range []string{m.Subject, text, html} {
				if regexp.MustCompile(`\{[a-z]+\}|\b(mail|verify|reset|exists|invite|due_soon|date)\.[a-z_0-9]+`).MatchString(strings.ReplaceAll(part, home, "")) {
					t.Errorf("%s %s: a placeholder or a key is left in %q", locale, l.Kind, part)
				}
			}
			if strings.ContainsAny(m.Subject, "\r\n") || m.Subject == "" {
				t.Errorf("%s %s: subject %q is not one line", locale, l.Kind, m.Subject)
			}
			if !strings.HasPrefix(html, "<!doctype html>") || !strings.Contains(html, `lang="`+locale+`"`) || !regexp.MustCompile(`<a\s`).MatchString(html) {
				t.Errorf("%s %s: not a page in %s: %.80q", locale, l.Kind, locale, html)
			}
			if tag.MatchString(text) || strings.Contains(text, "&amp;") {
				t.Errorf("%s %s: the plain text has markup: %q", locale, l.Kind, text)
			}
			if want := home + l.Path; !strings.Contains(text, want) || !strings.Contains(html, want) {
				t.Errorf("%s %s: the link %s is not in both parts", locale, l.Kind, want)
			}
			if !strings.Contains(text, "Ada Lovelace,") {
				t.Errorf("%s %s: no greeting by name in %q", locale, l.Kind, text)
			}
		}
	}
}

// Each language speaks its own words, French first.
func TestEachMailSpeaksItsLanguage(t *testing.T) {
	subjects := map[string][2]string{
		KindVerify:  {"Confirmez votre adresse", "Confirm your email address"},
		KindReset:   {"Choisissez un nouveau mot de passe", "Reset your Todo password"},
		KindExists:  {"Vous avez déjà un compte", "You already have a Todo account"},
		KindInvite:  {"Grace Hopper vous invite à créer un compte", "Grace Hopper invited you to Todo"},
		KindDueSoon: {"Bientôt à échéance : Ship the release", "Due soon: Ship the release"},
	}
	for i, locale := range []string{"fr", "en"} {
		for _, l := range letters(locale) {
			if m, _, _, err := write(l, home); err != nil || m.Subject != subjects[l.Kind][i] {
				t.Errorf("%s %s: subject %q (%v), want %q", locale, l.Kind, m.Subject, err, subjects[l.Kind][i])
			}
		}
	}
	// A language nobody speaks, or none, is French.
	l := letters("")[0]
	if m, _, _, err := write(l, home); err != nil || m.Subject != "Confirmez votre adresse" {
		t.Errorf("no language: %q (%v), want French", m.Subject, err)
	}
	l.Locale = "de"
	if m, _, _, err := write(l, home); err != nil || m.Subject != "Confirmez votre adresse" {
		t.Errorf("an unknown language: %q (%v), want French", m.Subject, err)
	}
}

// What a person put in a name or a title is only text: escaped in the HTML,
// one line in the subject, never a tag of the plain text's own.
func TestAHostileNameAndTitleAreOnlyText(t *testing.T) {
	hostile := `<script>alert(1)</script> "x" & <img src=x onerror=y>`
	for _, locale := range []string{"fr", "en"} {
		l := letters(locale)[4]
		l.Name = hostile + "\r\nBcc: evil@example.com"
		l.Params = map[string]string{"title": hostile + "\r\nBcc: evil@example.com", "due": "2026-11-16T14:30:00Z"}
		m, html, text, err := write(l, home)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(html, "<script>") || strings.Contains(html, "<img src=x") || !strings.Contains(html, "&lt;script&gt;") {
			t.Errorf("%s: the HTML lets a hostile text through: %s", locale, html)
		}
		if strings.ContainsAny(m.Subject, "\r\n") || utf8Len(m.Subject) > 90 {
			t.Errorf("%s: subject %q is not a short line", locale, m.Subject)
		}
		if !strings.Contains(text, "<script>alert(1)</script>") {
			t.Errorf("%s: the plain text lost the person's own words: %q", locale, text)
		}
	}
	// The same through an invitation, whose actor is a name too.
	l := letters("en")[3]
	l.Params = map[string]string{"actor": hostile}
	if _, html, _, err := write(l, home); err != nil || strings.Contains(html, "<script>") {
		t.Errorf("an invitation by a hostile name: %v, %s", err, html)
	}
}

// Links are the base address and a path of the web app, nothing else.
func TestALinkIsBuiltFromTheBaseAddress(t *testing.T) {
	for _, base := range []string{"https://todo.example.com", "https://todo.example.com/"} {
		if got, err := link(base, "/verify?token=a%2Fb"); err != nil || got != "https://todo.example.com/verify?token=a%2Fb" {
			t.Errorf("link(%q): %q (%v)", base, got, err)
		}
	}
	for _, path := range []string{"verify", "//evil.example.com/x", "https://evil.example.com", "/a b", `/a\b`} {
		if got, err := link(home, path); err == nil {
			t.Errorf("link(%q) = %q, want it refused", path, got)
		}
	}
	if _, err := link("not an address", "/x"); err == nil {
		t.Error("a base that is no address was taken")
	}
	l := letters("fr")[0]
	l.Path = "//evil.example.com"
	if _, _, _, err := write(l, home); err == nil {
		t.Error("a letter whose path names another host was written")
	}
}

// A reminder writes the due date in the recipient's zone, in their language.
func TestAReminderWritesTheDueDateInTheRecipientsZone(t *testing.T) {
	cases := []struct{ locale, zone, want string }{
		{"fr", "Europe/Paris", "lundi 16 novembre à 15:30 CET"},
		{"fr", "", "lundi 16 novembre à 14:30 UTC"},
		{"en", "America/New_York", "Monday, November 16 at 09:30 EST"},
		{"en", "Local", "Monday, November 16 at 14:30 UTC"},
		{"en", "Mars/Olympus", "Monday, November 16 at 14:30 UTC"},
		{"fr", "Asia/Tokyo", "lundi 16 novembre à 23:30 JST"},
	}
	for _, c := range cases {
		l := letters(c.locale)[4]
		l.TimeZone = c.zone
		if _, _, text, err := write(l, home); err != nil || !strings.Contains(strings.Join(strings.Fields(text), " "), strings.Join(strings.Fields(c.want), " ")) {
			t.Errorf("%s in %q: %q (%v), want %q", c.locale, c.zone, text, err, c.want)
		}
	}
	first := time.Date(2026, 12, 1, 9, 0, 0, 0, time.UTC)
	if got := wordsIn("fr", "").date(first); !strings.Contains(got, "1er") {
		t.Errorf("the first of the month: %q, want 1er", got)
	}
}

// A kind nobody wrote, a due date that is none, and no link where the mail
// needs none: the first two are refused, the third is written without a button.
func TestALetterThatFitsNoMailIsRefused(t *testing.T) {
	l := letters("fr")[0]
	l.Kind = "newsletter"
	if _, _, _, err := write(l, home); err == nil {
		t.Error("an unknown kind was written")
	}
	d := letters("fr")[4]
	d.Params = map[string]string{"title": "x", "due": "tomorrow"}
	if _, _, _, err := write(d, home); err == nil {
		t.Error("a due date that is none was written")
	}
	if _, _, text, err := write(Letter{To: "a@example.com", Locale: "en", Kind: KindExists}, home); err != nil || strings.Count(text, "http") != 1 { // the footer's address alone
		t.Errorf("a mail with no path: %q (%v), want no action link", text, err)
	}
}

// utf8Len is the number of characters of s.
func utf8Len(s string) int { return len([]rune(s)) }
