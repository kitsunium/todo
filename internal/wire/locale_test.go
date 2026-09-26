package wire

import (
	"strings"
	"testing"
	"time"
)

// A browser's languages pick one of the product's, French first: French when
// the header names neither, is empty, or is garbage.
func TestNegotiateLocale(t *testing.T) {
	for header, want := range map[string]Locale{
		"":                        French,
		"en-US,en;q=0.9":          English,
		"en-GB":                   English,
		"fr-CA,fr;q=0.9,en;q=0.8": French,
		"de-DE,de;q=0.9,en;q=0.5": English,
		"de-DE,de;q=0.9":          French,
		"en;q=0, de":              French,
		"*":                       French,
		"%%% not a header ;;;":    French,
	} {
		if got := NegotiateLocale(header); got != want {
			t.Errorf("Accept-Language %q: %q, want %q", header, got, want)
		}
	}
}

// Only the product's languages pass, spelled as the API spells them; an
// unknown or absent one reads French.
func TestLocalesAreChecked(t *testing.T) {
	for _, s := range []string{"fr", "en"} {
		if l, err := CheckLocale("locale", s); err != nil || string(l) != s {
			t.Errorf("%q: %q, %v", s, l, err)
		}
	}
	for _, s := range []string{"", "de", "EN", "fr-FR", " fr"} {
		if _, err := CheckLocale("locale", s); !Is(err, "invalid_argument") {
			t.Errorf("%q was accepted: %v", s, err)
		}
	}
	if English.Resolve() != English || Locale("").Resolve() != French || Locale("de").Resolve() != French {
		t.Error("Resolve does not fall back to French")
	}
}

// A zone of the IANA database passes, spelled as the database spells it;
// anything else is refused, and an unknown one reads UTC.
func TestTimeZonesAreChecked(t *testing.T) {
	for _, s := range []string{"Europe/Paris", "UTC", "America/Argentina/Buenos_Aires", "  Asia/Tokyo "} {
		if z, err := CheckTimeZone("timeZone", s); err != nil || z != strings.TrimSpace(s) {
			t.Errorf("%q: %q, %v", s, z, err)
		}
	}
	if z, err := CheckTimeZone("timeZone", ""); err != nil || z != "" {
		t.Errorf("no zone: %q, %v", z, err)
	}
	for _, s := range []string{"Local", "Mars/Olympus_Mons", "../../etc/passwd", "europe/paris", strings.Repeat("A", 65)} {
		if _, err := CheckTimeZone("timeZone", s); !Is(err, "invalid_argument") {
			t.Errorf("%q was accepted: %v", s, err)
		}
	}
	if Zone("") != time.UTC || Zone("Not/A_Zone") != time.UTC || Zone("Europe/Paris").String() != "Europe/Paris" {
		t.Error("Zone does not fall back to UTC")
	}
}
