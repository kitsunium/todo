// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package account

import (
	"errors"
	"math/rand/v2"
	"strings"
	"testing"
	"time"
	"unicode"
)

// propertyAccountLockedOnlyBeforeItsEnd holds identity/Account.Locked: an
// account is locked exactly while now is before a LockedUntil it has.
func propertyAccountLockedOnlyBeforeItsEnd(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for range 2000 {
		now := base.Add(time.Duration(r.Int64N(int64(100 * time.Hour))))
		a := Account{FailedLogins: r.IntN(10)}
		if r.IntN(3) > 0 {
			a.LockedUntil = base.Add(time.Duration(r.Int64N(int64(100 * time.Hour))))
		}
		got := a.Locked(now)
		if a.LockedUntil.IsZero() && got {
			t.Fatalf("an account with no lock is locked at %s", now)
		}
		if want := !a.LockedUntil.IsZero() && now.Before(a.LockedUntil); got != want {
			t.Fatalf("locked until %s, at %s: %v, want %v", a.LockedUntil, now, got, want)
		}
	}
}

// propertyNormalizeEmailLowerOrRefused holds identity/function/NormalizeEmail:
// a text is answered as an address in lower case, with no space, one @
// between two parts, and stable under a second pass, or refused with
// ErrInvalidEmail and nothing.
func propertyNormalizeEmailLowerOrRefused(t *testing.T, seed uint64) {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed))
	pieces := []string{"ada", "Ada", "@", "@", "example.com", "EXAMPLE.org", " ", "\n", ".", "é", "", "a@b", "\t"}
	for range 3000 {
		var b strings.Builder
		for range 1 + r.IntN(5) {
			b.WriteString(pieces[r.IntN(len(pieces))])
		}
		s := b.String()
		got, err := NormalizeEmail(s)
		if err != nil {
			if !errors.Is(err, ErrInvalidEmail) || got != "" {
				t.Fatalf("NormalizeEmail(%q) = %q, %v: want ErrInvalidEmail and nothing", s, got, err)
			}
			continue
		}
		local, domain, _ := strings.Cut(got, "@")
		if local == "" || domain == "" || strings.Count(got, "@") != 1 || got != strings.ToLower(got) ||
			strings.IndexFunc(got, unicode.IsSpace) >= 0 {
			t.Fatalf("NormalizeEmail(%q) = %q, no address the product keeps", s, got)
		}
		if again, err := NormalizeEmail(got); err != nil || again != got {
			t.Fatalf("NormalizeEmail(%q) = %q, %v: an address is its own normal form", got, again, err)
		}
	}
}
