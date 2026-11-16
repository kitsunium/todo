// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package team

import (
	"hash/fnv"
	"slices"
)

// palette is the colors a group may have, the web app has a swatch for each.
func palette() []string {
	return []string{"slate", "red", "orange", "amber", "green", "teal", "blue", "indigo", "violet", "pink"}
}

// colorFor is ColorFor's body: a color of the palette is kept, none is picked
// from the name by a hash, so a name has one color whoever creates it.
func colorFor(color string, name string) (r0 string, err error) {
	colors := palette()
	if color == "" {
		h := fnv.New32a()
		_, _ = h.Write([]byte(name))
		return colors[h.Sum32()%uint32(len(colors))], nil
	}
	if !slices.Contains(colors, color) {
		return "", ErrBadColor
	}
	return color, nil
}
