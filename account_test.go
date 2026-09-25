package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/kitsunium/todo/identity"
)

// sessionCookie returns the todo_session cookie a response set, if any.
func sessionCookie(h http.Header) (*http.Cookie, bool) {
	for _, line := range h.Values("Set-Cookie") {
		c, err := http.ParseSetCookie(line)
		if err == nil && c.Name == identity.CookieName {
			return c, true
		}
	}
	return nil, false
}

// A new account: signed up, told to verify, verified by the link of its
// mail, signed in by it, renamed, used from an API client, signed out.
func TestAnAccountFromSignupToSignOut(t *testing.T) {
	h := start(t, false)
	alice := h.client("alice")
	signup := map[string]any{"email": "  Alice@Example.com ", "name": "Alice Martin", "password": testPassword}

	got := call[struct{ Status, Email string }](alice, http.StatusOK, "POST", "/api/auth/signup", signup)
	if got.Status != "verification_sent" || got.Email != "alice@example.com" {
		t.Fatalf("signup answered %+v", got)
	}
	verification := h.mail("alice@example.com", "/verify?token=")
	if verification.Subject != "Confirmez votre adresse e-mail" || !strings.Contains(verification.HTML, "Confirmer mon adresse e-mail") {
		t.Errorf("verification mail %q", verification.Subject)
	}

	// Unverified: signed out, and told to verify only once the password is right.
	alice.fails(http.StatusUnauthorized, "unauthenticated", "GET", "/api/auth/me", nil)
	alice.fails(http.StatusUnauthorized, "invalid_credentials", "POST", "/api/auth/login", map[string]any{"email": "alice@example.com", "password": "not the password"})
	alice.fails(http.StatusForbidden, "email_unverified", "POST", "/api/auth/login", map[string]any{"email": "alice@example.com", "password": testPassword})

	// Signing up again tells a stranger nothing, and mails the owner.
	again := call[struct{ Status, Email string }](h.client("stranger"), http.StatusOK, "POST", "/api/auth/signup",
		map[string]any{"email": "alice@example.com", "name": "Mallory", "password": "another password"})
	if again != got {
		t.Errorf("a second sign-up answered %+v, want %+v", again, got)
	}
	h.mail("alice@example.com", "Vous en avez déjà un")

	// A resend mails a second link; the first stays good until one is used.
	call[struct{ Status string }](alice, http.StatusOK, "POST", "/api/auth/verify/resend", map[string]any{"email": "alice@example.com"})
	call[struct{ Status string }](alice, http.StatusOK, "POST", "/api/auth/verify/resend", map[string]any{"email": "nobody@example.com"})
	h.settle("the second verification mail", 50*time.Millisecond, func() bool { return len(mailsTo("alice@example.com", "/verify?token=")) == 2 })
	links := mailsTo("alice@example.com", "/verify?token=")

	code, raw, header := alice.do("POST", "/api/auth/verify", map[string]any{"token": token(t, links[0])})
	if code != http.StatusOK {
		t.Fatalf("verify: %d %s", code, raw)
	}
	cookie, ok := sessionCookie(header)
	if !ok || !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge != int(identity.SessionFor/time.Second) || len(cookie.Value) != 43 {
		t.Fatalf("session cookie %+v", cookie)
	}
	who := call[me](alice, http.StatusOK, "GET", "/api/auth/me", nil).User
	if who.Name != "Alice Martin" || who.Email != "alice@example.com" || !strings.HasPrefix(who.ID, "user_") || who.Locale != "fr" {
		t.Fatalf("me %+v", who)
	}

	// A link works once, and the second link of a verified account signs no one in.
	alice.fails(http.StatusBadRequest, "invalid_token", "POST", "/api/auth/verify", map[string]any{"token": token(t, links[0])})
	alice.fails(http.StatusBadRequest, "invalid_token", "POST", "/api/auth/verify", map[string]any{"token": token(t, links[1])})
	alice.fails(http.StatusBadRequest, "invalid_token", "POST", "/api/auth/verify", map[string]any{"token": "not-a-token"})

	renamed := call[me](alice, http.StatusOK, "PATCH", "/api/auth/me", map[string]any{"name": " Alice M. "})
	if renamed.User.Name != "Alice M." {
		t.Errorf("renamed to %q", renamed.User.Name)
	}
	alice.violates("name", "PATCH", "/api/auth/me", map[string]any{"name": "   "})
	alice.violates("name", "PATCH", "/api/auth/me", map[string]any{"name": "two\nlines"})

	sessions := call[struct {
		Sessions []struct {
			ID      string
			Current bool
			IP      string
		}
	}](alice, http.StatusOK, "GET", "/api/auth/sessions", nil)
	if len(sessions.Sessions) != 1 || !sessions.Sessions[0].Current || sessions.Sessions[0].IP == "" {
		t.Fatalf("sessions %+v", sessions)
	}

	// An API client presents the same secret as a bearer token.
	api := h.client("api")
	api.bearer = cookie.Value
	if got := call[me](api, http.StatusOK, "GET", "/api/auth/me", nil); got.User.ID != who.ID {
		t.Errorf("bearer is %+v", got)
	}

	// Using the session slides it, at most every five minutes.
	h.clk.Advance(6 * time.Minute)
	if _, _, header := alice.do("GET", "/api/auth/me", nil); header.Get("Set-Cookie") == "" {
		t.Error("a session used after five minutes did not slide its cookie")
	}

	_, _, header = alice.do("POST", "/api/auth/logout", nil)
	if c, ok := sessionCookie(header); !ok || c.MaxAge >= 0 {
		t.Errorf("logout left the cookie: %+v", c)
	}
	alice.fails(http.StatusUnauthorized, "unauthenticated", "GET", "/api/auth/me", nil)
	api.fails(http.StatusUnauthorized, "unauthenticated", "GET", "/api/auth/me", nil)

	// The rules of the sign-up form.
	anon := h.client("anon")
	anon.violates("email", "POST", "/api/auth/signup", map[string]any{"email": "not an address", "name": "X", "password": testPassword})
	anon.violates("password", "POST", "/api/auth/signup", map[string]any{"email": "x@example.com", "name": "X", "password": "short"})
	anon.violates("name", "POST", "/api/auth/signup", map[string]any{"email": "x@example.com", "name": "", "password": testPassword})
	anon.fails(http.StatusBadRequest, "invalid_argument", "POST", "/api/auth/signup", map[string]any{"email": "x@example.com", "name": "X", "password": testPassword, "admin": true})
}

// A session lasts thirty days without use: then it signs no one in, and its
// cookie is cleared.
func TestASessionExpires(t *testing.T) {
	h := start(t, false)
	eve := h.signup("Eve", "eve@example.com")
	eve.expect(http.StatusOK, "GET", "/api/auth/me", nil)
	h.clk.Advance(identity.SessionFor + time.Minute)
	code, _, header := eve.do("GET", "/api/auth/me", nil)
	if c, ok := sessionCookie(header); code != http.StatusUnauthorized || !ok || c.MaxAge >= 0 {
		t.Fatalf("an expired session: %d, cookie %+v", code, c)
	}
}

// Wrong passwords answer the same as unknown addresses; five in a row lock
// the account, whose own timer unlocks it fifteen minutes later.
func TestSignInFailuresLockTheAccount(t *testing.T) {
	h := start(t, false)
	h.signup("Bob", "bob@example.com")
	anon := h.client("anon")
	wrong := map[string]any{"email": "bob@example.com", "password": "not the password"}
	right := map[string]any{"email": "bob@example.com", "password": testPassword}

	unknown := anon.fails(http.StatusUnauthorized, "invalid_credentials", "POST", "/api/auth/login", map[string]any{"email": "nobody@example.com", "password": testPassword})
	for range identity.MaxFailedLogins {
		if e := anon.fails(http.StatusUnauthorized, "invalid_credentials", "POST", "/api/auth/login", wrong); e.Error.Message != unknown.Error.Message {
			t.Fatalf("a wrong password says %q, an unknown address %q", e.Error.Message, unknown.Error.Message)
		}
	}
	anon.fails(http.StatusForbidden, "account_locked", "POST", "/api/auth/login", right)

	status := func() identity.AccountStatus {
		found, err := identity.Accounts.Find(context.Background(), "email", "bob@example.com")
		if err != nil || len(found) != 1 {
			t.Fatalf("bob's account: %v %v", found, err)
		}
		return found[0].Status
	}
	h.settle("the lock guard", 200*time.Millisecond, func() bool { return status() == identity.Locked })
	anon.fails(http.StatusForbidden, "account_locked", "POST", "/api/auth/login", right)
	h.clk.Advance(identity.LockFor - time.Minute)
	h.drain()
	anon.fails(http.StatusForbidden, "account_locked", "POST", "/api/auth/login", right)
	h.settle("the unlock timer", 5*time.Second, func() bool { return status() == identity.Active })

	// Unlocked with a clean slate: one wrong password is one failure again.
	anon.fails(http.StatusUnauthorized, "invalid_credentials", "POST", "/api/auth/login", wrong)
	if got := call[me](anon, http.StatusOK, "POST", "/api/auth/login", right); got.User.Email != "bob@example.com" {
		t.Fatalf("signed in as %+v", got)
	}
}

// A password reset signs out every session and signs in the browser that
// used the link; changing the password signs out every other session. A
// reset also unlocks a locked account.
func TestPasswordsResetAndChange(t *testing.T) {
	h := start(t, false)
	laptop := h.signup("Carol", "carol@example.com")
	phone := h.client("phone")
	phone.expect(http.StatusOK, "POST", "/api/auth/login", map[string]any{"email": "carol@example.com", "password": testPassword})

	anon := h.client("anon")
	sent := call[struct{ Status string }](anon, http.StatusOK, "POST", "/api/auth/password/forgot", map[string]any{"email": "carol@example.com"})
	unknown := call[struct{ Status string }](anon, http.StatusOK, "POST", "/api/auth/password/forgot", map[string]any{"email": "nobody@example.com"})
	if sent != unknown || sent.Status != "sent" {
		t.Fatalf("forgot answered %+v and %+v", sent, unknown)
	}
	reset := h.mail("carol@example.com", "/reset?token=")
	secret := token(t, reset)
	anon.violates("password", "POST", "/api/auth/password/reset", map[string]any{"token": secret, "password": "short"})
	newPassword := "a brand new passphrase"
	if got := call[me](anon, http.StatusOK, "POST", "/api/auth/password/reset", map[string]any{"token": secret, "password": newPassword}); got.User.Email != "carol@example.com" {
		t.Fatalf("reset signed in %+v", got)
	}
	anon.fails(http.StatusBadRequest, "invalid_token", "POST", "/api/auth/password/reset", map[string]any{"token": secret, "password": newPassword})
	laptop.fails(http.StatusUnauthorized, "unauthenticated", "GET", "/api/auth/me", nil)
	phone.fails(http.StatusUnauthorized, "unauthenticated", "GET", "/api/auth/me", nil)
	anon.expect(http.StatusOK, "GET", "/api/auth/me", nil)
	h.client("x").fails(http.StatusUnauthorized, "invalid_credentials", "POST", "/api/auth/login", map[string]any{"email": "carol@example.com", "password": testPassword})

	// Change: the current password first, then every other session is out.
	laptop.expect(http.StatusOK, "POST", "/api/auth/login", map[string]any{"email": "carol@example.com", "password": newPassword})
	anon.violates("current", "POST", "/api/auth/password", map[string]any{"current": "wrong wrong wrong", "password": testPassword})
	anon.expect(http.StatusNoContent, "POST", "/api/auth/password", map[string]any{"current": newPassword, "password": testPassword})
	laptop.fails(http.StatusUnauthorized, "unauthenticated", "GET", "/api/auth/me", nil)
	anon.expect(http.StatusOK, "GET", "/api/auth/me", nil)

	// A reset unlocks.
	for range identity.MaxFailedLogins {
		h.client("x").fails(http.StatusUnauthorized, "invalid_credentials", "POST", "/api/auth/login", map[string]any{"email": "carol@example.com", "password": "not the password"})
	}
	phone.fails(http.StatusForbidden, "account_locked", "POST", "/api/auth/login", map[string]any{"email": "carol@example.com", "password": testPassword})
	h.drain()
	phone.expect(http.StatusOK, "POST", "/api/auth/password/forgot", map[string]any{"email": "carol@example.com"})
	h.settle("a second reset mail", 50*time.Millisecond, func() bool { return len(mailsTo("carol@example.com", "/reset?token=")) == 2 })
	links := mailsTo("carol@example.com", "/reset?token=")
	phone.expect(http.StatusOK, "POST", "/api/auth/password/reset", map[string]any{"token": token(t, links[1]), "password": newPassword})
	h.client("x").expect(http.StatusOK, "POST", "/api/auth/login", map[string]any{"email": "carol@example.com", "password": newPassword})
}

// A session is revoked from the list of sessions; revoking another user's
// is revoking nothing.
func TestSessionsAreRevoked(t *testing.T) {
	h := start(t, false)
	laptop := h.signup("Dan", "dan@example.com")
	phone := h.client("phone")
	phone.expect(http.StatusOK, "POST", "/api/auth/login", map[string]any{"email": "dan@example.com", "password": testPassword})
	other := h.signup("Erin", "erin@example.com")

	type session struct {
		ID      string
		Current bool
	}
	list := call[struct{ Sessions []session }](laptop, http.StatusOK, "GET", "/api/auth/sessions", nil).Sessions
	if len(list) != 2 {
		t.Fatalf("sessions %+v", list)
	}
	var phoneSession string
	for _, s := range list {
		if !s.Current {
			phoneSession = s.ID
		}
	}
	other.fails(http.StatusNotFound, "not_found", "DELETE", "/api/auth/sessions/"+phoneSession, nil)
	laptop.expect(http.StatusNoContent, "DELETE", "/api/auth/sessions/"+phoneSession, nil)
	phone.fails(http.StatusUnauthorized, "unauthenticated", "GET", "/api/auth/me", nil)
	laptop.expect(http.StatusOK, "GET", "/api/auth/me", nil)
}

// Each account reads the product in its own language, French first: the
// sign-up form names it, or the browser's Accept-Language decides, and
// French when neither does. The account changes it later, and the next
// mails follow; a mail to someone without an account is in the inviter's.
func TestEachAccountReadsItsLanguage(t *testing.T) {
	h := start(t, false)

	// A browser that prefers English opens an English account.
	eve := h.client("eve")
	eve.languages = "en-US,en;q=0.9"
	eve.expect(http.StatusOK, "POST", "/api/auth/signup", map[string]any{"email": "eve@example.com", "name": "Eve", "password": testPassword})
	welcome := h.mail("eve@example.com", "/verify?token=")
	if welcome.Subject != "Confirm your email address" || !strings.Contains(welcome.HTML, `<html lang="en">`) || !strings.Contains(welcome.Text, "Hi Eve,") {
		t.Errorf("an English browser was welcomed with %q:\n%s", welcome.Subject, welcome.Text)
	}
	if got := call[me](eve, http.StatusOK, "POST", "/api/auth/verify", map[string]any{"token": token(t, welcome)}); got.User.Locale != "en" {
		t.Errorf("eve reads %q", got.User.Locale)
	}

	// No header, a language the product does not speak: French.
	for i, languages := range []string{"", "de-DE,de;q=0.9", "*"} {
		c := h.client("fr")
		c.languages = languages
		email := fmt.Sprintf("fr%d@example.com", i)
		c.expect(http.StatusOK, "POST", "/api/auth/signup", map[string]any{"email": email, "name": "Zoé", "password": testPassword})
		m := h.mail(email, "/verify?token=")
		if m.Subject != "Confirmez votre adresse e-mail" || !strings.Contains(m.HTML, `<html lang="fr">`) {
			t.Errorf("Accept-Language %q: %q", languages, m.Subject)
		}
		if got := call[me](c, http.StatusOK, "POST", "/api/auth/verify", map[string]any{"token": token(t, m)}); got.User.Locale != "fr" {
			t.Errorf("Accept-Language %q reads %q", languages, got.User.Locale)
		}
	}

	// The form's language wins over the browser's; a language the product
	// does not speak is refused.
	form := h.client("form")
	form.languages = "en"
	form.expect(http.StatusOK, "POST", "/api/auth/signup", map[string]any{"email": "form@example.com", "name": "Form", "password": testPassword, "locale": "fr"})
	h.mail("form@example.com", "Confirmez votre adresse e-mail")
	form.violates("locale", "POST", "/api/auth/signup", map[string]any{"email": "x@example.com", "name": "X", "password": testPassword, "locale": "de"})

	// A French user and an English one work together, each in their language.
	alice := h.signup("Alice", "alice@example.com")
	added := call[struct{ Request struct{ ID string } }](alice, http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "eve@example.com"})
	h.mail("eve@example.com", "Alice wants to add you as a contact")
	eve.expect(http.StatusOK, "POST", "/api/contacts/"+added.Request.ID+"/accept", nil)
	h.mail("alice@example.com", "Eve a accepté votre demande de contact")
	eve.expect(http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "dan@example.com"})
	h.mail("dan@example.com", "Eve invited you to Todo")
	alice.expect(http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "yann@example.com"})
	h.mail("yann@example.com", "Alice vous invite sur Todo")

	// Switching languages: the account says so, and the next mails follow.
	if got := call[me](alice, http.StatusOK, "PATCH", "/api/auth/me", map[string]any{"locale": "en"}); got.User.Locale != "en" || got.User.Name != "Alice" {
		t.Fatalf("alice switched to %+v", got.User)
	}
	h.client("anon").expect(http.StatusOK, "POST", "/api/auth/password/forgot", map[string]any{"email": "alice@example.com"})
	h.mail("alice@example.com", "Reset your Todo password")
	alice.expect(http.StatusOK, "POST", "/api/contacts", map[string]any{"email": "zed@example.com"})
	h.mail("zed@example.com", "Alice invited you to Todo")
	both := call[me](alice, http.StatusOK, "PATCH", "/api/auth/me", map[string]any{"name": "Alice M.", "locale": "fr"})
	if both.User.Locale != "fr" || both.User.Name != "Alice M." {
		t.Errorf("name and language at once: %+v", both.User)
	}
	if same := call[me](alice, http.StatusOK, "PATCH", "/api/auth/me", map[string]any{}); same.User != both.User {
		t.Errorf("an empty change changed %+v into %+v", both.User, same.User)
	}
	for _, bad := range []string{"de", "EN", "", "fr-FR"} {
		alice.violates("locale", "PATCH", "/api/auth/me", map[string]any{"locale": bad})
	}
	if got := call[me](alice, http.StatusOK, "GET", "/api/auth/me", nil); got.User.Locale != "fr" {
		t.Errorf("a refused change changed the language to %q", got.User.Locale)
	}

	// An account opened before the product spoke two languages reads French.
	eveID := eve.id()
	if _, err := identity.Accounts.Update(context.Background(), eveID, func(a *identity.Account) error { a.Locale = ""; return nil }); err != nil {
		t.Fatal(err)
	}
	if got := call[me](eve, http.StatusOK, "GET", "/api/auth/me", nil); got.User.Locale != "fr" {
		t.Errorf("an account without a language reads %q", got.User.Locale)
	}
	h.client("anon").expect(http.StatusOK, "POST", "/api/auth/password/forgot", map[string]any{"email": "eve@example.com"})
	h.mail("eve@example.com", "Réinitialisez votre mot de passe Todo")
}
