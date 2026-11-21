// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package team

import (
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func someUser(r *rand.Rand) string { return "user_" + strconv.Itoa(r.IntN(30)) }

// someMembers draws members of distinct ids, any role (the owner too).
func someMembers(r *rand.Rand) []Membership {
	roles := []Role{RoleOwner, RoleAdmin, RoleMember}
	var out []Membership
	seen := map[string]bool{}
	for range r.IntN(12) {
		id := someUser(r)
		if seen[id] {
			continue
		}
		seen[id] = true
		out = append(out, Membership{UserID: id, Role: roles[r.IntN(len(roles))], JoinedAt: time.Unix(int64(r.IntN(1000)), 0)})
	}
	return out
}

// propertyGroupMemberIDsOneIDAMember holds groups/Group.MemberIDs: one id a
// member, in the group's order.
func propertyGroupMemberIDsOneIDAMember(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	for range 2000 {
		g := Group{Members: someMembers(r)}
		ids := g.MemberIDs()
		if len(ids) != len(g.Members) {
			t.Fatalf("%d ids for %d members", len(ids), len(g.Members))
		}
		for i, m := range g.Members {
			if ids[i] != m.UserID {
				t.Fatalf("id %d is %s, want %s", i, ids[i], m.UserID)
			}
		}
	}
}

// propertyGroupRoleOfEachMemberHasItsRole holds groups/Group.RoleOf: a member
// has the role the group keeps for them, a stranger has none.
func propertyGroupRoleOfEachMemberHasItsRole(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	for range 2000 {
		g := Group{Members: someMembers(r)}
		for _, m := range g.Members {
			if got := g.RoleOf(m.UserID); got != m.Role {
				t.Fatalf("%s is %q, want %q", m.UserID, got, m.Role)
			}
		}
		if stranger := someUser(r); !slices.Contains(g.MemberIDs(), stranger) && g.RoleOf(stranger) != "" {
			t.Fatalf("the stranger %s has the role %q", stranger, g.RoleOf(stranger))
		}
	}
}

// propertyCleanNameTrimmedOrRefused holds groups/function/CleanName: a name
// is answered trimmed, or refused when it is blank, too long or holds a
// control character.
func propertyCleanNameTrimmedOrRefused(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	pieces := []string{"a", "Équipe", " ", "  ", "x", "\n", "\t", "é", "名", strings.Repeat("z", 30)}
	for range 3000 {
		var b strings.Builder
		for range r.IntN(6) {
			b.WriteString(pieces[r.IntN(len(pieces))])
		}
		in := b.String()
		got, err := CleanName(in)
		if err != nil {
			continue
		}
		if got != strings.TrimSpace(in) || got == "" || len([]rune(got)) > MaxName || strings.ContainsAny(got, "\n\t") {
			t.Fatalf("%q is answered %q", in, got)
		}
	}
	for _, in := range []string{"", "   ", "a\nb", strings.Repeat("x", MaxName+1)} {
		if _, err := CleanName(in); err == nil {
			t.Fatalf("%q is accepted", in)
		}
	}
	if got, err := CleanName(" " + strings.Repeat("x", MaxName) + " "); err != nil || len(got) != MaxName {
		t.Fatalf("a name of MaxName characters: %q (%v)", got, err)
	}
}

// propertyColorForKnownOrRefused holds groups/function/ColorFor: a color of
// the palette is kept, an unknown one refused, and none is picked from the
// name, the same one every time.
func propertyColorForKnownOrRefused(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	colors := palette()
	for range 2000 {
		name := "g" + strconv.Itoa(r.IntN(1000))
		known := colors[r.IntN(len(colors))]
		if got, err := ColorFor(known, name); err != nil || got != known {
			t.Fatalf("%s is answered %q (%v)", known, got, err)
		}
		picked, err := ColorFor("", name)
		if err != nil || !slices.Contains(colors, picked) {
			t.Fatalf("the color picked for %s: %q (%v)", name, picked, err)
		}
		if again, _ := ColorFor("", name); again != picked {
			t.Fatalf("%s is picked %q then %q", name, picked, again)
		}
		if _, err := ColorFor("not-"+known, name); err == nil {
			t.Fatalf("an unknown color is accepted")
		}
	}
}

// propertyManagesOwnerAndAdminOnly holds groups/function/Manages: the owner
// and the admins run the group, a member and a stranger do not.
func propertyManagesOwnerAndAdminOnly(t *testing.T, seed uint64) {
	t.Helper()
	for _, c := range []struct {
		role Role
		want bool
	}{{RoleOwner, true}, {RoleAdmin, true}, {RoleMember, false}, {"", false}, {"root", false}} {
		if got := Manages(c.role); got != c.want {
			t.Fatalf("Manages(%q) is %v, want %v", c.role, got, c.want)
		}
	}
}

// propertyParseRoleAssignableOnly holds groups/function/ParseRole: admin and
// member are read, anything else, the owner too, is refused.
func propertyParseRoleAssignableOnly(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	for range 2000 {
		var b strings.Builder
		for range r.IntN(8) {
			b.WriteByte(byte('a' + r.IntN(26)))
		}
		s := b.String()
		got, err := ParseRole(s)
		assigned := s == "admin" || s == "member"
		if assigned != (err == nil) || (err == nil && string(got) != s) {
			t.Fatalf("ParseRole(%q) is %q (%v)", s, got, err)
		}
	}
	if _, err := ParseRole(string(RoleOwner)); err == nil {
		t.Fatal("the owner is given")
	}
}

// propertyRankedOwnerFirst holds groups/function/Ranked: the same members
// come back, the owner before the admins before the members, each rank by when
// they joined, and the input is left as it was.
func propertyRankedOwnerFirst(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	rank := map[Role]int{RoleOwner: 0, RoleAdmin: 1, RoleMember: 2}
	for range 2000 {
		in := someMembers(r)
		before := slices.Clone(in)
		out := Ranked(in)
		if !slices.Equal(in, before) {
			t.Fatal("Ranked changed its input")
		}
		if len(out) != len(in) {
			t.Fatalf("%d members come back of %d", len(out), len(in))
		}
		for i := 1; i < len(out); i++ {
			a, b := out[i-1], out[i]
			if rank[a.Role] > rank[b.Role] || (a.Role == b.Role && a.JoinedAt.After(b.JoinedAt)) {
				t.Fatalf("%+v comes before %+v", a, b)
			}
		}
	}
}
