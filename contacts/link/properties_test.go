// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package link

import (
	"math/rand/v2"
	"strconv"
	"testing"
)

func someID(r *rand.Rand) string { return "user_" + strconv.Itoa(r.IntN(50)) }

// propertyLinkOtherTheOtherEnd holds contacts/Link.Other: each end of a link
// has the other one, and a member at neither end has nobody.
func propertyLinkOtherTheOtherEnd(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	for range 2000 {
		a, b, c := someID(r), someID(r), someID(r)
		if a == b {
			continue
		}
		l := Link{RequesterID: a, AddresseeID: b}
		if got, ok := l.Other(a); !ok || got != b {
			t.Fatalf("the other end from %s: %q, %v, want %s", a, got, ok, b)
		}
		if got, ok := l.Other(b); !ok || got != a {
			t.Fatalf("the other end from %s: %q, %v, want %s", b, got, ok, a)
		}
		if got, ok := l.Other(c); c != a && c != b && (ok || got != "") {
			t.Fatalf("%s is at no end of %s and %s, and is answered %q", c, a, b, got)
		}
	}
}

// propertyPairOfSymmetric holds contacts/function/PairOf: a pair is the same
// whatever the order, and two different pairs never share a name.
func propertyPairOfSymmetric(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	for range 2000 {
		a, b, c, d := someID(r), someID(r), someID(r), someID(r)
		if PairOf(a, b) != PairOf(b, a) {
			t.Fatalf("PairOf(%s, %s) is not PairOf(%s, %s)", a, b, b, a)
		}
		same := (a == c && b == d) || (a == d && b == c)
		if !same && PairOf(a, b) == PairOf(c, d) {
			t.Fatalf("the pairs %s,%s and %s,%s share the name %s", a, b, c, d, PairOf(a, b))
		}
	}
}
