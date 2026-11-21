package authapi

import "testing"

// The language of a request is the heaviest of fr and en its Accept-Language
// header names, French when it names neither, and a malformed member of the
// header is skipped, never an error.
func TestNegotiateTheLanguageOfARequest(t *testing.T) {
	for header, want := range map[string]string{
		"":                           "fr",
		"fr-FR,fr;q=0.9":             "fr",
		"en-US,en;q=0.9":             "en",
		"en":                         "en",
		"EN-gb":                      "en",
		"de,en;q=0.5":                "en",
		"de,es;q=0.8":                "fr",
		"fr;q=0.3,en;q=0.7":          "en",
		"en;q=0.8, fr;q=0.8":         "en", // equal weights: the header's order
		"fr;q=0,en;q=0.1":            "en", // q=0 is "not acceptable"
		"en;q=oops,fr;q=0.2":         "fr",
		"*":                          "fr",
		";;;,,,en;q=1.5,fr-CA;q=0.1": "fr",
		"english":                    "fr",
	} {
		if got := negotiate(header); got != want {
			t.Errorf("negotiate(%q) = %q, want %q", header, got, want)
		}
	}
}
