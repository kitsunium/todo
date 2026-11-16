// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package task

import (
	"strings"
)

// parsePriority is ParsePriority's body: the three priorities, whatever
// their case and the spaces around them, and normal for an empty one.
func parsePriority(s string) (r0 Priority, err error) {
	switch p := Priority(strings.ToLower(strings.TrimSpace(s))); p {
	case "":
		return PriorityNormal, nil
	case PriorityLow, PriorityNormal, PriorityHigh:
		return p, nil
	}
	return "", ErrUnknownPriority
}
