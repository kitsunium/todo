package notify

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
	"time"
	"unicode/utf8"

	"github.com/kitsunium/sdk/pkg/v1/errs"
	"github.com/kitsunium/todo/internal/wire"
)

// every returns one of each mail the product sends, written in l.
func every(l wire.Locale, actor, to, title, group string) map[string]Message {
	due := time.Date(2026, 9, 25, 17, 0, 0, 0, time.UTC)
	task := Task{ID: "task_1", Title: title, Priority: 1, Due: &due}
	return map[string]Message{
		"verify":           VerifyEmail(l, to, "SECRET"),
		"reset":            ResetPassword(l, to, "SECRET"),
		"exists":           AccountExists(l, to),
		"contact request":  ContactRequest(l, to, actor),
		"contact accepted": ContactAccepted(l, to, actor),
		"invite":           InviteToJoin(l, actor, "dan@example.com"),
		"group":            GroupInvitation(l, to, actor, "group_1", group, "violet"),
		"shared":           TaskShared(l, to, actor, task),
		"assigned":         TaskAssigned(l, to, actor, task),
		"due":              DueSoon(l, to, task, 15*time.Minute),
	}
}

// Every mail renders in the one light design, in each language, as HTML and
// as plain text, with its one button pointing into the web app at
// TODO_BASE_URL — and whatever a user typed stays text.
func TestEveryMailRendersInTheDesign(t *testing.T) {
	t.Setenv("TODO_BASE_URL", "https://todo.example.com/")
	evil := `<script>alert("x")</script> & co`
	for _, l := range wire.Locales {
		mails := every(l, evil, "Alice Martin", evil, evil)
		for name, m := range mails {
			name = string(l) + " " + name
			html, text, err := render(m)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			if m.Lang != l || m.Subject == "" || m.Headline == "" || m.Greeting == "" || len(m.Body) == 0 || m.Action.Label == "" || m.Reason == "" {
				t.Errorf("%s: incomplete %+v", name, m)
			}
			if !strings.HasPrefix(m.Action.URL, "https://todo.example.com/") {
				t.Errorf("%s: the button links to %q", name, m.Action.URL)
			}
			if strings.ContainsAny(m.Subject, "\r\n") || utf8.RuneCountInString(m.Subject) > 120 {
				t.Errorf("%s: subject %q", name, m.Subject)
			}
			for _, want := range []string{"<!doctype html>", `<html lang="` + string(l) + `">`, "#f26b1d", "max-width: 560px", `content="light only"`, m.Action.Label} {
				if !strings.Contains(html, want) {
					t.Errorf("%s: the HTML part lacks %q", name, want)
				}
			}
			for _, dark := range []string{"prefers-color-scheme", "light dark", "#0e0e10", "#161618"} {
				if strings.Contains(html, dark) {
					t.Errorf("%s: the HTML part still has a dark mode: %q", name, dark)
				}
			}
			if strings.Contains(html, "<script>") {
				t.Errorf("%s: user text reached the HTML unescaped", name)
			}
			if !strings.Contains(text, m.Action.URL) || !strings.Contains(text, m.Headline) {
				t.Errorf("%s: the text part lacks the link or the headline:\n%s", name, text)
			}
			for _, line := range strings.Split(text, "\n") {
				if utf8.RuneCountInString(line) > textWidth && !strings.Contains(line, "://") {
					t.Errorf("%s: a text line is %d wide: %q", name, utf8.RuneCountInString(line), line)
				}
				// A no-break space holds French punctuation to its word.
				if strings.HasPrefix(line, ":") || strings.HasPrefix(line, "?") || strings.HasPrefix(line, "»") {
					t.Errorf("%s: a line starts with its punctuation: %q", name, line)
				}
			}
		}
		if u := mails["verify"].Action.URL; u != "https://todo.example.com/verify?token=SECRET" {
			t.Errorf("verify links to %q", u)
		}
		if u := mails["invite"].Action.URL; u != "https://todo.example.com/signup?email=dan%40example.com" {
			t.Errorf("invite links to %q", u)
		}
		if u := mails["group"].Action.URL; u != "https://todo.example.com/app/groups/group_1" {
			t.Errorf("group links to %q", u)
		}
	}

	// Dates, plurals and punctuation, the way each language writes them.
	en, fr := every(wire.English, "Alice", "Bob", evil, "Design"), every(wire.French, "Alice", "Bob", evil, "Design")
	for _, c := range []struct{ got, want string }{
		{en["due"].Subject, "Due in 15 minutes: " + clip(evil, 60)},
		{en["due"].Card.Detail, "Due Friday, September 25 at 17:00 UTC · Urgent"},
		{fr["due"].Subject, "Échéance dans 15\u00a0minutes\u00a0: " + clip(evil, 60)},
		{fr["due"].Card.Detail, "Échéance\u00a0: vendredi 25\u00a0septembre à 17:00\u00a0UTC · Urgente"},
		{fr["shared"].Subject, "Alice a partagé «\u00a0" + clip(evil, 60) + "\u00a0» avec vous"},
		{fr["group"].Subject, "Alice vous invite à rejoindre Design"},
		{fr["verify"].Greeting, "Bonjour Bob,"},
		{fr["invite"].Greeting, "Bonjour,"},
		{en["invite"].Greeting, "Hi,"},
		{DueSoon(wire.French, "Bob", Task{Title: "T"}, time.Minute).Headline, "Échéance dans 1\u00a0minute"},
		{DueSoon(wire.English, "Bob", Task{Title: "T"}, time.Minute).Headline, "Due in 1 minute"},
		{DueSoon(wire.French, "Bob", Task{Title: "T"}, 10*time.Second).Headline, "Échéance imminente"},
		{wordsIn(wire.French).date(time.Date(2026, 9, 24, 17, 0, 0, 0, time.UTC)), "jeudi 24\u00a0septembre à 17:00\u00a0UTC"},
		{wordsIn(wire.French).date(time.Date(2026, 10, 1, 8, 30, 0, 0, time.UTC)), "jeudi 1er\u00a0octobre à 08:30\u00a0UTC"},
		{wordsIn(wire.English).date(time.Date(2026, 10, 1, 8, 30, 0, 0, time.UTC)), "Thursday, October 1 at 08:30 UTC"},
		{ContactRequest(wire.French, "Bob", "").Subject, "Quelqu’un souhaite vous ajouter à ses contacts"},
		{VerifyEmail("de", "Bob", "x").Subject, "Confirmez votre adresse e-mail"}, // French first
	} {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}

// A subject keeps its sentence whole: a long name and a long title are
// clipped, so that every subject stays on one line of at most 120
// characters, in every language.
func TestSubjectsStayShortInEveryLanguage(t *testing.T) {
	name := "Marie-Christine Delacroix-Beaumont de la Tour d’Auvergne et Saint-Ambroise"
	title := strings.Repeat("Préparer la réunion du conseil d’administration ", 5)
	for _, l := range wire.Locales {
		for mail, m := range every(l, name, name, title, title) {
			if n := utf8.RuneCountInString(m.Subject); n > 120 || strings.ContainsAny(m.Subject, "\r\n") {
				t.Errorf("%s %s: a subject of %d characters: %q", l, mail, n, m.Subject)
			}
		}
	}
}

// Nothing of the English catalogue reaches a French mail: every word only
// the English catalogue uses is absent from the French mails — their
// subjects, their plain text and what their HTML shows — names and links
// aside.
func TestFrenchMailsHaveNoEnglishLeft(t *testing.T) {
	english := catalogueWords(t, "en")
	for w := range catalogueWords(t, "fr") {
		delete(english, w)
	}
	if len(english) < 50 {
		t.Fatalf("only %d words are English-only: the test lost its reference", len(english))
	}
	link := regexp.MustCompile(`\S+://\S+|\S+@\S+`)
	tag := regexp.MustCompile(`(?s)<head>.*</head>|<[^>]+>`)
	for name, m := range every(wire.French, "Alice", "Bob", "Budget 2027", "Design") {
		htmlPart, text, err := render(m)
		if err != nil {
			t.Fatal(err)
		}
		shown := html.UnescapeString(tag.ReplaceAllString(htmlPart, " "))
		for part, s := range map[string]string{"subject": m.Subject, "preheader": m.Preheader, "text": text, "html": shown} {
			for _, w := range letters.FindAllString(link.ReplaceAllString(s, " "), -1) {
				if english[strings.ToLower(w)] {
					t.Errorf("French %s mail, %s: an English word %q in\n%s", name, part, w, s)
				}
			}
		}
	}
}

// letters finds the words of a text.
var letters = regexp.MustCompile(`\p{L}+`)

// catalogueWords returns the lower-cased words of a catalogue's messages,
// their placeholders aside.
func catalogueWords(t *testing.T, lang string) map[string]bool {
	t.Helper()
	raw, err := catalogueFiles.ReadFile("locales/" + lang + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var members map[string]any
	if err := json.Unmarshal(raw, &members); err != nil {
		t.Fatal(err)
	}
	placeholder := regexp.MustCompile(`\{[A-Za-z_][A-Za-z0-9_]*\}`)
	out := map[string]bool{}
	add := func(s string) {
		for _, w := range letters.FindAllString(placeholder.ReplaceAllString(s, " "), -1) {
			out[strings.ToLower(w)] = true
		}
	}
	for _, v := range members {
		switch v := v.(type) {
		case string:
			add(v)
		case map[string]any:
			for _, form := range v {
				add(form.(string))
			}
		}
	}
	return out
}

// The catalogues are checked when the process starts: a counted message
// missing a plural form its language needs, a key one language lacks, or a
// language the product does not speak stops it there.
func TestIncompleteCataloguesDoNotLoad(t *testing.T) {
	fr, _ := catalogueFiles.ReadFile("locales/fr.json")
	en, _ := catalogueFiles.ReadFile("locales/en.json")
	if _, err := loadPrinters(fstest.MapFS{"l/fr.json": {Data: fr}, "l/en.json": {Data: en}}, "l"); err != nil {
		t.Fatalf("the shipped catalogues: %v", err)
	}
	without := func(raw []byte, key string) []byte {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			t.Fatal(err)
		}
		delete(m, key)
		out, _ := json.Marshal(m)
		return out
	}
	noMany := strings.Replace(string(fr), `"many": "Échéance dans {n}\u00a0de minutes",`, "", 1)
	if noMany == string(fr) {
		t.Fatal("the French catalogue lost its many form")
	}
	for name, files := range map[string]fstest.MapFS{
		"a plural form French needs": {"l/fr.json": {Data: []byte(noMany)}, "l/en.json": {Data: en}},
		"a key English lacks":        {"l/fr.json": {Data: fr}, "l/en.json": {Data: without(en, "verify.body")}},
		"a key French lacks":         {"l/fr.json": {Data: without(fr, "verify.body")}, "l/en.json": {Data: en}},
		"a language not spoken":      {"l/fr.json": {Data: fr}, "l/en.json": {Data: en}, "l/de.json": {Data: en}},
		"a language missing":         {"l/fr.json": {Data: fr}},
		"a malformed pattern":        {"l/fr.json": {Data: fr}, "l/en.json": {Data: []byte(strings.Replace(string(en), "Hi {name},", "Hi {name,", 1))}},
	} {
		if _, err := loadPrinters(files, "l"); err == nil {
			t.Errorf("%s: the catalogues loaded", name)
		} else if name == "a plural form French needs" && !errs.HasReason(err, "TRANSLATION_INCOMPLETE") {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestClipKeepsSubjectsShortAndOnOneLine(t *testing.T) {
	if got := clip("Plan\r\nthe offsite", 60); got != "Plan  the offsite" {
		t.Errorf("clip kept a line break: %q", got)
	}
	long := strings.Repeat("word ", 30)
	if got := clip(long, 60); utf8.RuneCountInString(got) > 61 || !strings.HasSuffix(got, "…") {
		t.Errorf("clip(%d runes) = %q", utf8.RuneCountInString(long), got)
	}
}

// The plain text part breaks lines between words, never at a no-break
// space.
func TestWrapKeepsNoBreakSpaces(t *testing.T) {
	s := strings.Repeat("a", 64) + " e-mail\u00a0: fin" // "e-mail" fits on the first line, "e-mail\u00a0:" does not
	for _, line := range strings.Split(wrap(s), "\n") {
		if strings.HasPrefix(line, ":") || utf8.RuneCountInString(line) > textWidth {
			t.Errorf("wrapped at the no-break space: %q", wrap(s))
		}
	}
	if !strings.Contains(wrap(s), "e-mail\u00a0:") {
		t.Errorf("the no-break space is gone: %q", wrap(s))
	}
}
