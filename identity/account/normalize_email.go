// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package account

import (
	"strings"
	"unicode"
)

// maxEmail is the longest address there is (RFC 5321).
const maxEmail = 254

// normalizeEmail is NormalizeEmail's body: the address trimmed and in lower
// case. It is refused when it is not one local part, one @ and one domain,
// or holds a space or a control character: the product mails it, and a
// line break in an address would be a header.
func normalizeEmail(s string) (r0 string, err error) {
	s = strings.ToLower(strings.TrimSpace(s))
	local, domain, ok := strings.Cut(s, "@")
	switch {
	case !ok, local == "", domain == "", len(s) > maxEmail:
		return "", ErrInvalidEmail
	case strings.Contains(domain, "@"):
		return "", ErrInvalidEmail
	case strings.IndexFunc(s, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) >= 0:
		return "", ErrInvalidEmail
	}
	return s, nil
}
