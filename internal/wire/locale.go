package wire

import (
	"slices"

	"github.com/kitsunium/sdk/pkg/v1/i18n"
)

// Locale is a language the product speaks, spelled the way the API and the
// stores spell it: "fr" or "en".
type Locale string

// The languages of the product.
const (
	French  Locale = "fr"
	English Locale = "en"
)

// Locales are the languages the product speaks, French first: French is
// what a person reads when nothing says otherwise.
var Locales = []Locale{French, English}

// Resolve returns the language a person with this locale reads: the locale
// itself when the product speaks it, French otherwise — an account opened
// before the product spoke two languages has none.
func (l Locale) Resolve() Locale {
	if slices.Contains(Locales, l) {
		return l
	}
	return Locales[0]
}

// CheckLocale refuses anything but a language the product speaks, the way
// kit refuses a oneof tag: 400, invalid_argument, a violation on path.
func CheckLocale(path, s string) (Locale, error) {
	if l := Locale(s); slices.Contains(Locales, l) {
		return l, nil
	}
	return "", Invalid(path, "one_of", "must be one of: fr, en")
}

// negotiator resolves a browser's Accept-Language to one of the product's
// languages with the SDK's RFC 4647 lookup, French when the header names
// none of them. It is built once: a language the SDK has no plural rules
// for stops the process here, at startup.
var negotiator = func() *i18n.Negotiator {
	tags := make([]i18n.Tag, len(Locales))
	for i, l := range Locales {
		tag, err := i18n.ParseTag(string(l))
		if err != nil {
			panic(err)
		}
		tags[i] = tag
	}
	n, err := i18n.NewNegotiator(tags, tags[0])
	if err != nil {
		panic(err)
	}
	return n
}()

// NegotiateLocale returns the product's language an Accept-Language header
// prefers: "en-US,en;q=0.9" is English, "fr-CA" French, and "de" — or no
// header at all — French. A malformed header is never an error.
func NegotiateLocale(acceptLanguage string) Locale {
	return Locale(negotiator.Negotiate(acceptLanguage).String())
}
