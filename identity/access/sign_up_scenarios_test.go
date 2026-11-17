// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"errors"
	"strings"
	"testing"

	"github.com/kitsunium/todo/identity/account"
)

// byEmail is the one account of an address in the world, for a test to read.
func (w *world) byEmail(t *testing.T, email string) account.Account {
	t.Helper()
	a, ok, _ := accountsPort{w}.ByEmail(t.Context(), email)
	if !ok {
		t.Fatalf("no account of %s", email)
	}
	return a
}

// givenSignUpCreatesAnAccountToVerify builds SignUp for "creates an account to verify": the use case, its input, and
// what its output must be.
func givenSignUpCreatesAnAccountToVerify(t *testing.T) (SignUp, SignUpInput, func(*testing.T, SignUpOutput)) {
	t.Helper()
	w := newWorld()
	in := SignUpInput{Email: "  Ada@Example.com ", Name: " Ada   Lovelace ", Password: passphrase}
	return w.signUp(), in, func(t *testing.T, out SignUpOutput) {
		t.Helper()
		if out.Email != "ada@example.com" || out.Verified {
			t.Errorf("the answer: %+v", out)
		}
		a := w.byEmail(t, "ada@example.com")
		if a.Name != "Ada Lovelace" || a.Verified || !a.CreatedAt.Equal(epoch) || a.PasswordHash == "" || a.PasswordHash == passphrase {
			t.Errorf("the account created: %+v", a)
		}
		if len(w.links) != 1 || len(w.mails) != 1 {
			t.Fatalf("%d link(s) and %d mail(s), want one of each", len(w.links), len(w.mails))
		}
		for _, l := range w.links {
			if l.UserID != a.ID || l.Purpose != account.PurposeVerify || !l.ExpiresAt.Equal(epoch.Add(account.VerifyFor)) {
				t.Errorf("the link: %+v", l)
			}
		}
	}
}

// givenSignUpCreatesAVerifiedAccountWhereAddressesAreTrusted builds SignUp for "creates a verified account where addresses are trusted": the use case, its input, and
// what its output must be.
func givenSignUpCreatesAVerifiedAccountWhereAddressesAreTrusted(t *testing.T) (SignUp, SignUpInput, func(*testing.T, SignUpOutput)) {
	t.Helper()
	w := newWorld()
	in := SignUpInput{Email: "ada@example.com", Name: "Ada", Password: passphrase, Trusted: true}
	return w.signUp(), in, func(t *testing.T, out SignUpOutput) {
		t.Helper()
		if len(w.heard) != 1 || w.heard[0] != "verified ada@example.com" {
			t.Errorf("announced %q, want the address proven, once", w.heard)
		}
		if !out.Verified || !w.byEmail(t, "ada@example.com").Verified || len(w.links) != 0 || len(w.mails) != 0 {
			t.Errorf("answer %+v, %d link(s), %d mail(s): want a verified account and nothing mailed", out, len(w.links), len(w.mails))
		}
	}
}

// givenSignUpAnswersATakenAddressAlike builds SignUp for "answers a taken address alike": the use case, its input, and
// what its output must be.
func givenSignUpAnswersATakenAddressAlike(t *testing.T) (SignUp, SignUpInput, func(*testing.T, SignUpOutput)) {
	t.Helper()
	w := newWorld()
	first := w.member("ada@example.com")
	in := SignUpInput{Email: "ADA@example.com", Name: "Another Ada", Password: "another passphrase"}
	return w.signUp(), in, func(t *testing.T, out SignUpOutput) {
		t.Helper()
		if out.Email != "ada@example.com" || out.Verified {
			t.Errorf("the answer: %+v, want what a new address gets", out)
		}
		if len(w.accounts) != 1 || w.accounts[first.ID] != first {
			t.Errorf("the accounts: %+v, want the first one untouched", w.accounts)
		}
		if len(w.mails) != 1 || w.mails[0] != "already ada@example.com" {
			t.Errorf("the mails: %q, want its member told", w.mails)
		}
	}
}

// givenSignUpRefusesAnInvalidAddress builds SignUp for "refuses an invalid address": the use case and the input it
// refuses with account.ErrInvalidEmail.
func givenSignUpRefusesAnInvalidAddress(t *testing.T) (SignUp, SignUpInput) {
	t.Helper()
	return newWorld().signUp(), SignUpInput{Email: "ada at example.com", Name: "Ada", Password: passphrase}
}

// givenSignUpRefusesAShortPassword builds SignUp for "refuses a short password": the use case and the input it
// refuses with account.ErrWeakPassword.
func givenSignUpRefusesAShortPassword(t *testing.T) (SignUp, SignUpInput) {
	t.Helper()
	return newWorld().signUp(), SignUpInput{Email: "ada@example.com", Name: "Ada", Password: "too short"}
}

// givenSignUpKeepsTheLanguageAndTheTimeZone builds SignUp for "keeps the language and the time zone": the use case, its input, and
// what its output must be.
func givenSignUpKeepsTheLanguageAndTheTimeZone(t *testing.T) (SignUp, SignUpInput, func(*testing.T, SignUpOutput)) {
	t.Helper()
	w := newWorld()
	in := SignUpInput{Email: "ada@example.com", Name: "Ada", Password: passphrase, Locale: "en", TimeZone: " Europe/Paris "}
	return w.signUp(), in, func(t *testing.T, out SignUpOutput) {
		t.Helper()
		a := w.byEmail(t, "ada@example.com")
		if a.Locale != account.LocaleEN || a.TimeZone != "Europe/Paris" {
			t.Errorf("the account keeps %q and %q, want en and Europe/Paris", a.Locale, a.TimeZone)
		}
		// Nothing said is French and UTC.
		w2 := newWorld()
		if _, err := w2.signUp().Execute(t.Context(), SignUpInput{Email: "bob@example.com", Name: "Bob", Password: passphrase}); err != nil {
			t.Fatal(err)
		}
		if b := w2.byEmail(t, "bob@example.com"); b.Locale != account.LocaleFR || b.TimeZone != "" {
			t.Errorf("the account with nothing said: %q and %q, want fr and none", b.Locale, b.TimeZone)
		}
	}
}

// givenSignUpRefusesAnUnknownLanguage builds SignUp for "refuses an unknown language": the use case and the input it
// refuses with account.ErrInvalidLocale.
func givenSignUpRefusesAnUnknownLanguage(t *testing.T) (SignUp, SignUpInput) {
	t.Helper()
	return newWorld().signUp(), SignUpInput{Email: "ada@example.com", Name: "Ada", Password: passphrase, Locale: "de"}
}

// givenSignUpRefusesAnUnknownTimeZone builds SignUp for "refuses an unknown time zone": the use case and the input it
// refuses with account.ErrInvalidTimeZone.
func givenSignUpRefusesAnUnknownTimeZone(t *testing.T) (SignUp, SignUpInput) {
	t.Helper()
	return newWorld().signUp(), SignUpInput{Email: "ada@example.com", Name: "Ada", Password: passphrase, TimeZone: "Mars/Olympus"}
}

// Zones are spelled as the database spells them, and nothing else is one.
func TestCheckZone(t *testing.T) {
	for _, ok := range []string{"", "UTC", "Europe/Paris", "America/Argentina/Buenos_Aires", "  Asia/Tokyo "} {
		if _, err := checkZone(ok); err != nil {
			t.Errorf("checkZone(%q): %v, want it taken", ok, err)
		}
	}
	for _, bad := range []string{"Local", "europe/paris", "Europe//Paris", "Europe/Nowhere", "/etc/passwd", "../Europe/Paris", strings.Repeat("A", 65)} {
		if _, err := checkZone(bad); !errors.Is(err, account.ErrInvalidTimeZone) {
			t.Errorf("checkZone(%q): %v, want ErrInvalidTimeZone", bad, err)
		}
	}
}
