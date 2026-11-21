// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package team

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// cleanName is CleanName's body: the name without the space around it, which
// must hold one to MaxName characters and no line break or other control
// character, since a name travels into notifications.
func cleanName(name string) (r0 string, err error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > MaxName || strings.IndexFunc(name, unicode.IsControl) >= 0 {
		return "", ErrBadName
	}
	return name, nil
}
