package wire

import (
	"strings"
	"time"

	// The zone database, inside the binary: the product's image has none.
	_ "time/tzdata"
)

// maxTimeZone bounds a zone name: the longest the IANA database holds is
// half of it.
const maxTimeZone = 64

// CheckTimeZone refuses anything but a zone of the IANA database —
// "Europe/Paris", "UTC" — the way kit refuses a malformed field: 400,
// invalid_argument, a violation on path. The empty name is no zone, and
// comes back empty.
func CheckTimeZone(path, name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", nil
	}
	if len(name) <= maxTimeZone && spelledLikeAZone(name) {
		if _, err := time.LoadLocation(name); err == nil {
			return name, nil
		}
	}
	return "", Invalid(path, "time_zone", "must be a time zone of the IANA database, like Europe/Paris")
}

// spelledLikeAZone reports whether name is spelled the way the IANA
// database spells a zone: each part of the path starts with a capital
// letter. A file system that ignores case would otherwise find
// "europe/paris" on one machine and not on the next; "Local" is Go's word
// for the machine's own zone, not the user's.
func spelledLikeAZone(name string) bool {
	if name == "Local" {
		return false
	}
	for part := range strings.SplitSeq(name, "/") {
		if part == "" || part[0] < 'A' || part[0] > 'Z' {
			return false
		}
	}
	return true
}

// Zone is the location name names, UTC when it is empty or unknown.
func Zone(name string) *time.Location {
	if name == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}
