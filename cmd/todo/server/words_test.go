package server_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// The API words its refusals in English; the page says them in the
// account's language. Every sentence a handler can refuse with, and every
// rule of a field it can name, has its words in both dictionaries of the
// page: a refusal added to a handler without them fails here.
func TestThePageHasWordsForEveryRefusalOfTheAPI(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	raw, err := os.ReadFile(filepath.Join(root, "web", "site", "page", "lang.js"))
	if err != nil {
		t.Fatal(err)
	}
	words := string(raw)
	handlers, err := filepath.Glob(filepath.Join(root, "*", "*", "handlers.go"))
	if err != nil || len(handlers) < 5 {
		t.Fatalf("the handlers: %v (%v), want those of every API", handlers, err)
	}
	sentence := regexp.MustCompile(`(?:app\.(?:NotFound|Conflict|Forbidden|Unauthenticated)\(|Message: )("(?:[^"\\]|\\.)*")`)
	field := regexp.MustCompile(`invalid\("([A-Za-z]+)", "([a-z_]+)"`)
	keys := map[string]string{}
	for _, file := range handlers {
		if strings.Contains(file, "legacy") {
			continue
		}
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, m := range sentence.FindAllStringSubmatch(string(src), -1) {
			said, err := strconv.Unquote(m[1])
			if err != nil {
				t.Fatalf("%s: %s: %v", file, m[1], err)
			}
			keys["api."+said] = file
		}
		for _, m := range field.FindAllStringSubmatch(string(src), -1) {
			keys["api.field."+m[1]+"."+m[2]] = file
		}
	}
	if len(keys) < 40 {
		t.Fatalf("%d refusals found in the handlers, want them all: the patterns no longer match the code", len(keys))
	}
	for key, file := range keys {
		if n := strings.Count(words, strconv.Quote(key)+":"); n != 2 {
			t.Errorf("%s refuses with %q, which the page's dictionaries hold %d time(s), want once in each language", filepath.Base(filepath.Dir(file)), key, n)
		}
	}
}
