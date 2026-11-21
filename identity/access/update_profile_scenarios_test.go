// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"testing"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

// updateProfile is the use case of a world.
func (w *world) updateProfile() UpdateProfile {
	return NewUpdateProfile(clockPort{w}, accountsPort{w})
}

// text is a pointer to a string, as a request that sends the member holds it.
func text(s string) *string { return &s }

// givenUpdateProfileChangesWhatIsGiven builds UpdateProfile for "changes what is given": the use case, its input, and
// what its output must be.
func givenUpdateProfileChangesWhatIsGiven(t *testing.T) (UpdateProfile, UpdateProfileInput, func(*testing.T, UpdateProfileOutput)) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	w.now = epoch.Add(time.Hour)
	in := UpdateProfileInput{UserID: a.ID, Name: text("  Ada   King "), Locale: text("en"), TimeZone: text("Europe/Paris")}
	return w.updateProfile(), in, func(t *testing.T, out UpdateProfileOutput) {
		t.Helper()
		got := w.accounts[a.ID]
		if got.Name != "Ada King" || got.Locale != account.LocaleEN || got.TimeZone != "Europe/Paris" || !got.UpdatedAt.Equal(w.now) || out.Account != got {
			t.Errorf("the account: %+v, answered %+v", got, out.Account)
		}
	}
}

// givenUpdateProfileKeepsWhatIsNotGiven builds UpdateProfile for "keeps what is not given": the use case, its input, and
// what its output must be.
func givenUpdateProfileKeepsWhatIsNotGiven(t *testing.T) (UpdateProfile, UpdateProfileInput, func(*testing.T, UpdateProfileOutput)) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	a.Locale, a.TimeZone = account.LocaleEN, "Asia/Tokyo"
	w.accounts[a.ID] = a
	return w.updateProfile(), UpdateProfileInput{UserID: a.ID, Name: text("Grace")}, func(t *testing.T, out UpdateProfileOutput) {
		t.Helper()
		if out.Account.Name != "Grace" || out.Account.Locale != account.LocaleEN || out.Account.TimeZone != "Asia/Tokyo" {
			t.Errorf("the account: %+v, want the name changed and the rest kept", out.Account)
		}
	}
}

// givenUpdateProfileEmptiesTheTimeZoneToUtc builds UpdateProfile for "empties the time zone to utc": the use case, its input, and
// what its output must be.
func givenUpdateProfileEmptiesTheTimeZoneToUtc(t *testing.T) (UpdateProfile, UpdateProfileInput, func(*testing.T, UpdateProfileOutput)) {
	t.Helper()
	w := newWorld()
	a := w.member("ada@example.com")
	a.TimeZone = "Asia/Tokyo"
	w.accounts[a.ID] = a
	return w.updateProfile(), UpdateProfileInput{UserID: a.ID, TimeZone: text("")}, func(t *testing.T, out UpdateProfileOutput) {
		t.Helper()
		if out.Account.TimeZone != "" || w.accounts[a.ID].TimeZone != "" {
			t.Errorf("the zone: %q, want none", out.Account.TimeZone)
		}
	}
}

// givenUpdateProfileRefusesAnEmptyName builds UpdateProfile for "refuses an empty name": the use case and the input it
// refuses with account.ErrNameRequired.
func givenUpdateProfileRefusesAnEmptyName(t *testing.T) (UpdateProfile, UpdateProfileInput) {
	t.Helper()
	w := newWorld()
	return w.updateProfile(), UpdateProfileInput{UserID: w.member("ada@example.com").ID, Name: text("   ")}
}

// givenUpdateProfileRefusesAnUnknownLanguage builds UpdateProfile for "refuses an unknown language": the use case and the input it
// refuses with account.ErrInvalidLocale.
func givenUpdateProfileRefusesAnUnknownLanguage(t *testing.T) (UpdateProfile, UpdateProfileInput) {
	t.Helper()
	w := newWorld()
	return w.updateProfile(), UpdateProfileInput{UserID: w.member("ada@example.com").ID, Locale: text("")}
}

// givenUpdateProfileRefusesAnUnknownTimeZone builds UpdateProfile for "refuses an unknown time zone": the use case and the input it
// refuses with account.ErrInvalidTimeZone.
func givenUpdateProfileRefusesAnUnknownTimeZone(t *testing.T) (UpdateProfile, UpdateProfileInput) {
	t.Helper()
	w := newWorld()
	return w.updateProfile(), UpdateProfileInput{UserID: w.member("ada@example.com").ID, TimeZone: text("Europe/Nowhere")}
}
