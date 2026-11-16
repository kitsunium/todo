// The ports the use cases' tests run on: accounts and sessions in memory,
// passwords "stretched" by a prefix, secrets counted, a clock a test sets.

package access

import (
	"context"
	"strconv"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

// epoch is the instant every test starts at.
var epoch = time.Date(2026, 11, 16, 9, 0, 0, 0, time.UTC)

// passphrase is the right one of every test account.
const passphrase = "correct horse battery"

// world is every port of the domain, in memory.
type world struct {
	accounts map[string]account.Account
	sessions map[string]account.Session
	links    map[string]account.Link
	mails    []string // what the courier was asked, "kind to secret"
	now      time.Time
	n        int
}

func newWorld() *world {
	return &world{accounts: map[string]account.Account{}, sessions: map[string]account.Session{}, links: map[string]account.Link{}, now: epoch}
}

// member adds a verified account with the test passphrase.
func (w *world) member(email string) account.Account {
	w.n++
	a := account.Account{ID: "user_" + strconv.Itoa(w.n), Email: email, Name: "Ada", PasswordHash: "hash:" + passphrase, Verified: true, CreatedAt: w.now, UpdatedAt: w.now}
	w.accounts[a.ID] = a
	return a
}

// session opens a session of a whose secret is secret, last seen at seen.
func (w *world) session(a account.Account, secret string, seen time.Time) account.Session {
	w.n++
	s := account.Session{ID: "session_" + strconv.Itoa(w.n), UserID: a.ID, TokenHash: "digest:" + secret, CreatedAt: seen, LastSeenAt: seen, ExpiresAt: seen.Add(account.SessionFor)}
	w.sessions[s.ID] = s
	return s
}

// link gives a member a one-time link of a purpose whose secret is secret.
func (w *world) link(a account.Account, purpose, secret string, expires time.Time) account.Link {
	w.n++
	l := account.Link{ID: "link_" + strconv.Itoa(w.n), Hash: "digest:" + secret, Purpose: purpose, UserID: a.ID, CreatedAt: epoch, ExpiresAt: expires}
	w.links[l.ID] = l
	return l
}

type (
	linksPort     struct{ *world }
	courierPort   struct{ *world }
	accountsPort  struct{ *world }
	sessionsPort  struct{ *world }
	passwordsPort struct{ *world }
	secretsPort   struct{ *world }
	clockPort     struct{ *world }
)

func (p accountsPort) NewID(context.Context) (string, error) {
	p.n++
	return "user_" + strconv.Itoa(p.n), nil
}

func (p accountsPort) Save(_ context.Context, a account.Account) error {
	p.accounts[a.ID] = a
	return nil
}

func (p accountsPort) Get(_ context.Context, id string) (account.Account, bool, error) {
	a, ok := p.accounts[id]
	return a, ok, nil
}

func (p accountsPort) Change(_ context.Context, id string, change func(*account.Account) error) (account.Account, error) {
	a, ok := p.accounts[id]
	if !ok {
		return account.Account{}, account.ErrInvalidCredentials
	}
	if err := change(&a); err != nil {
		return account.Account{}, err
	}
	p.accounts[id] = a
	return a, nil
}

func (p accountsPort) ByEmail(_ context.Context, email string) (account.Account, bool, error) {
	for _, a := range p.accounts {
		if a.Email == email {
			return a, true, nil
		}
	}
	return account.Account{}, false, nil
}

func (p sessionsPort) NewID(context.Context) (string, error) {
	p.n++
	return "session_" + strconv.Itoa(p.n), nil
}

func (p sessionsPort) Save(_ context.Context, s account.Session) error {
	p.sessions[s.ID] = s
	return nil
}

func (p sessionsPort) ByTokenHash(_ context.Context, hash string) (account.Session, bool, error) {
	for _, s := range p.sessions {
		if s.TokenHash == hash {
			return s, true, nil
		}
	}
	return account.Session{}, false, nil
}

func (p sessionsPort) Delete(_ context.Context, id string) error {
	delete(p.sessions, id)
	return nil
}

func (p sessionsPort) Of(_ context.Context, userID string) ([]account.Session, error) {
	var out []account.Session
	for _, s := range p.sessions {
		if s.UserID == userID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (p sessionsPort) DeleteOf(_ context.Context, userID, keep string) error {
	for id, s := range p.sessions {
		if s.UserID == userID && id != keep {
			delete(p.sessions, id)
		}
	}
	return nil
}

func (p linksPort) NewID(context.Context) (string, error) {
	p.n++
	return "link_" + strconv.Itoa(p.n), nil
}

func (p linksPort) Save(_ context.Context, l account.Link) error {
	p.links[l.ID] = l
	return nil
}

func (p linksPort) ByHash(_ context.Context, hash string) (account.Link, bool, error) {
	for _, l := range p.links {
		if l.Hash == hash {
			return l, true, nil
		}
	}
	return account.Link{}, false, nil
}

func (p linksPort) DeleteOf(_ context.Context, userID, purpose string) error {
	for id, l := range p.links {
		if l.UserID == userID && l.Purpose == purpose {
			delete(p.links, id)
		}
	}
	return nil
}

func (p courierPort) Verification(_ context.Context, to account.Account, secret string) error {
	p.mails = append(p.mails, "verify "+to.Email+" "+secret)
	return nil
}

func (p courierPort) Reset(_ context.Context, to account.Account, secret string) error {
	p.mails = append(p.mails, "reset "+to.Email+" "+secret)
	return nil
}

func (p courierPort) AlreadyMember(_ context.Context, to account.Account) error {
	p.mails = append(p.mails, "already "+to.Email)
	return nil
}

func (passwordsPort) Hash(_ context.Context, password string) (string, error) {
	return "hash:" + password, nil
}

func (passwordsPort) Verify(_ context.Context, password, hash string) (bool, error) {
	return hash != "" && hash == "hash:"+password, nil
}

func (p secretsPort) New(context.Context) (string, string, error) {
	p.n++
	secret := "secret_" + strconv.Itoa(p.n)
	return secret, "digest:" + secret, nil
}

func (secretsPort) Digest(_ context.Context, secret string) (string, error) {
	return "digest:" + secret, nil
}

func (p clockPort) Now(context.Context) (time.Time, error) { return p.now, nil }

func (w *world) signUp() SignUp {
	return NewSignUp(accountsPort{w}, passwordsPort{w}, clockPort{w}, linksPort{w}, secretsPort{w}, courierPort{w})
}

func (w *world) signIn() SignIn {
	return NewSignIn(accountsPort{w}, clockPort{w}, passwordsPort{w}, secretsPort{w}, sessionsPort{w})
}

func (w *world) authenticate() Authenticate {
	return NewAuthenticate(secretsPort{w}, sessionsPort{w}, clockPort{w}, accountsPort{w})
}
