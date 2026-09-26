package main

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// A mail writes its times in its reader's zone: the one their browser gave
// at sign-up, then at each sign-in — a user who travelled reads the zone
// they signed in from. A zone the IANA database does not know is refused.
func TestMailsSpeakTheReadersTimeZone(t *testing.T) {
	h := start(t, false)
	signupIn := func(name, email, zone string) *client {
		t.Helper()
		c := h.client(name)
		c.languages = "en"
		c.expect(http.StatusOK, "POST", "/api/auth/signup", map[string]any{"email": email, "name": name, "password": testPassword, "timeZone": zone})
		c.expect(http.StatusOK, "POST", "/api/auth/verify", map[string]any{"token": token(t, h.mail(email, "/verify?token="))})
		return c
	}
	alice := signupIn("Alice", "alice@example.com", "Europe/Paris")
	bob := signupIn("Bob", "bob@example.com", "America/New_York")
	befriend(alice, bob, "bob@example.com")
	bobID := bob.id()

	// Due at 02:30 UTC the day after the epoch: the evening before in New
	// York, the early morning in Paris.
	due := time.Date(epoch.Year(), epoch.Month(), epoch.Day()+1, 2, 30, 0, 0, time.UTC)
	share := func(title string) {
		t.Helper()
		plan := call[struct{ ID string }](alice, http.StatusOK, "POST", "/api/tasks", map[string]any{"title": title, "due": due.Format(time.RFC3339)})
		alice.expect(http.StatusOK, "POST", "/api/tasks/"+plan.ID+"/share", map[string]any{"userId": bobID})
	}
	share("Plan")
	inNewYork := due.In(mustZone(t, "America/New_York")).Format("15:04 MST")
	if m := h.mail("bob@example.com", "Plan"); !strings.Contains(strings.Join(strings.Fields(m.Text), " "), "at "+inNewYork) {
		t.Errorf("bob's mail does not say %q:\n%s", inNewYork, m.Text)
	}

	// Bob signs in from Tokyo: the next mail is in Tokyo's time.
	bob.expect(http.StatusOK, "POST", "/api/auth/login", map[string]any{"email": "bob@example.com", "password": testPassword, "timeZone": "Asia/Tokyo"})
	share("Retro")
	inTokyo := due.In(mustZone(t, "Asia/Tokyo")).Format("15:04 MST")
	if m := h.mail("bob@example.com", "Retro"); !strings.Contains(strings.Join(strings.Fields(m.Text), " "), "at "+inTokyo) {
		t.Errorf("after a sign-in from Tokyo, bob's mail does not say %q:\n%s", inTokyo, m.Text)
	}

	anon := h.client("anon")
	for _, zone := range []string{"Mars/Olympus_Mons", "Local", "../../etc/passwd", strings.Repeat("A", 65)} {
		anon.violates("timeZone", "POST", "/api/auth/signup", map[string]any{"email": "x@example.com", "name": "X", "password": testPassword, "timeZone": zone})
	}
	anon.violates("timeZone", "POST", "/api/auth/login", map[string]any{"email": "bob@example.com", "password": testPassword, "timeZone": "Nowhere/Land"})
}

func mustZone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Fatal(err)
	}
	return loc
}
