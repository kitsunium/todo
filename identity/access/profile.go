package access

import (
	"strings"
	"time"

	// The zone database, inside the binary: the product's image has none.
	_ "time/tzdata"

	"github.com/kitsunium/todo/identity/account"
)

// maxTimeZone bounds a zone name: the longest the IANA database holds is
// half of it.
const maxTimeZone = 64

// checkLocale is the language a member asked for: fr or en, nothing else.
func checkLocale(s string) (string, error) {
	switch s {
	case account.LocaleFR, account.LocaleEN:
		return s, nil
	}
	return "", account.ErrInvalidLocale
}

// checkZone is the zone of the IANA database a member named, or empty when
// they named none. Each part of the name starts with a capital letter, as the
// database spells it: a file system that ignores case would find
// "europe/paris" on one machine and not on the next, and "Local" is Go's word
// for the machine's own zone, not the member's.
func checkZone(s string) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", nil
	}
	if len(s) > maxTimeZone || s == "Local" {
		return "", account.ErrInvalidTimeZone
	}
	for part := range strings.SplitSeq(s, "/") {
		if part == "" || part[0] < 'A' || part[0] > 'Z' {
			return "", account.ErrInvalidTimeZone
		}
	}
	if _, err := time.LoadLocation(s); err != nil {
		return "", account.ErrInvalidTimeZone
	}
	return s, nil
}
