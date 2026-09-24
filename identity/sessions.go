package identity

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/internal/wire"
)

// CookieName is the session cookie. Credentials reads it by this name.
const CookieName = "todo_session"

// SessionFor is how long a session lives without being used; every use
// slides it, at most every slideEvery.
const (
	SessionFor = 30 * 24 * time.Hour
	slideEvery = 5 * time.Minute
)

// Session is one signed-in browser or API client. The product keeps the
// SHA-256 of its secret; the secret itself lives only in the client's cookie
// or bearer header.
type Session struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	TokenHash  string    `json:"tokenHash" kit:"secret"`
	CreatedAt  time.Time `json:"createdAt"`
	LastSeenAt time.Time `json:"lastSeenAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	UserAgent  string    `json:"userAgent"`
	IP         string    `json:"ip"`
}

// Sessions keeps every session, found by the hash of its secret and listed
// per user.
var Sessions = Service.Store("sessions", func(s Session) string { return s.ID },
	kit.Unique("token", func(s Session) string { return s.TokenHash }),
	kit.Index("user", func(s Session) []string { return []string{s.UserID} }))

// Credentials are what a request proves its user with: the session cookie a
// browser sends, or the same secret as a bearer token from an API client.
type Credentials struct {
	// Cookie is the todo_session cookie (CookieName).
	Cookie string `cookie:"todo_session"`
	// Authorization is "Bearer <secret>".
	Authorization string `header:"Authorization"`
}

// Principal is what the session handler says about the caller: which of the
// user's sessions the request runs in.
type Principal struct {
	SessionID string `json:"sessionId"`
}

// SessionAuth is the app's authentication: every endpoint declared with
// kit.Auth() runs it first.
var SessionAuth = Service.AuthHandler("session", Authenticate)

// Authenticate turns a session secret into its user. An unknown or expired
// session is refused, and its cookie cleared; a live one slides its expiry
// forward, at most every five minutes, so an active user stays signed in.
func Authenticate(ctx context.Context, c Credentials) (kit.UID, Principal, error) {
	secret, fromCookie := c.Cookie, c.Cookie != ""
	if !fromCookie {
		scheme, token, ok := strings.Cut(c.Authorization, " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") {
			return "", Principal{}, nil
		}
		secret = strings.TrimSpace(token)
	}
	refuse := func() (kit.UID, Principal, error) {
		if fromCookie {
			kit.ClearCookie(ctx, CookieName)
		}
		return "", Principal{}, kit.Unauthenticated("your session has expired: sign in again")
	}
	if !wellFormed(secret) {
		return refuse()
	}
	digest := digestOf(secret)
	found, err := Sessions.Find(ctx, "token", digest)
	if err != nil {
		return "", Principal{}, err
	}
	now := wire.Now(ctx)
	if len(found) == 0 || !sameDigest(found[0].TokenHash, digest) || !now.Before(found[0].ExpiresAt) {
		return refuse()
	}
	s := found[0]
	if now.Sub(s.LastSeenAt) >= slideEvery {
		s, err = Sessions.Update(ctx, s.ID, func(s *Session) error {
			s.LastSeenAt, s.ExpiresAt = now, now.Add(SessionFor)
			return nil
		})
		if wire.Is(err, kit.CodeNotFound) {
			return refuse() // revoked while this request ran
		}
		if err != nil {
			return "", Principal{}, err
		}
		if fromCookie {
			setSessionCookie(ctx, secret)
		}
	}
	return kit.UID(s.UserID), Principal{SessionID: s.ID}, nil
}

// openSession signs the user in: a new session, and its cookie on the
// response.
func openSession(ctx context.Context, userID string) error {
	secret, digest := newSecret()
	now := wire.Now(ctx)
	if err := Sessions.Insert(ctx, Session{
		ID: kit.NewID("session"), UserID: userID, TokenHash: digest,
		CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(SessionFor),
		UserAgent: kit.UserAgent(ctx), IP: kit.ClientIP(ctx),
	}); err != nil {
		return err
	}
	setSessionCookie(ctx, secret)
	return nil
}

// setSessionCookie gives the browser the session's secret. kit adds Secure
// outside dev; HttpOnly and SameSite=Lax keep it from scripts and from
// cross-site requests.
func setSessionCookie(ctx context.Context, secret string) {
	kit.SetCookie(ctx, &http.Cookie{
		Name: CookieName, Value: secret, Path: "/", MaxAge: int(SessionFor / time.Second),
		HttpOnly: true, SameSite: http.SameSiteLaxMode,
	})
}

// revokeSessions ends every session of the user but keep ("" for none).
func revokeSessions(ctx context.Context, userID, keep string) error {
	sessions, err := Sessions.Find(ctx, "user", userID)
	if err != nil {
		return err
	}
	for _, s := range sessions {
		if s.ID == keep {
			continue
		}
		if err := Sessions.Delete(ctx, s.ID); err != nil && !wire.Is(err, kit.CodeNotFound) {
			return err
		}
	}
	return nil
}

// caller returns the authenticated user of the request and their session.
// Every endpoint that calls it is declared with kit.Auth(), so a missing
// user is kit's bug or an in-process call without one: refused either way.
func caller(ctx context.Context) (string, Principal, error) {
	uid, ok := kit.UserID(ctx)
	if !ok {
		return "", Principal{}, kit.Unauthenticated("sign in first")
	}
	p, _ := kit.AuthData[Principal](ctx)
	return string(uid), p, nil
}
