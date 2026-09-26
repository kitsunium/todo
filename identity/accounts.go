package identity

import (
	"context"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/internal/wire"
)

// AccountStatus is where an account stands in its lifecycle.
type AccountStatus string

// The states of an account.
const (
	// Unverified accounts have not proved their address yet: they cannot
	// sign in.
	Unverified AccountStatus = "unverified"
	// Active accounts sign in.
	Active AccountStatus = "active"
	// Locked accounts refused too many wrong passwords in a row; they unlock
	// by themselves, or with a password reset.
	Locked AccountStatus = "locked"
)

// Account is one user's account. It never leaves this service: the API and
// the other services see a User or a UserRef.
type Account struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	// Locale is the language the user reads the product and its mails in.
	// An account opened before the product spoke two languages has none,
	// and reads French: see locale.
	Locale wire.Locale `json:"locale,omitempty"`
	// TimeZone is the zone of the IANA database the user's browser said it
	// was in, at sign-up and at each sign-in: the times in their mails are
	// written in it. Empty reads UTC.
	TimeZone string `json:"timeZone,omitempty"`
	// PasswordHash is the PHC string of the SDK's password hashing: salted,
	// slow, self-describing. The password itself is never stored.
	PasswordHash string        `json:"passwordHash" kit:"secret"`
	Status       AccountStatus `json:"status"`
	// FailedLogins counts the wrong passwords since the last success.
	FailedLogins int        `json:"failedLogins"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	VerifiedAt   *time.Time `json:"verifiedAt,omitempty"`
	LockedAt     *time.Time `json:"lockedAt,omitempty"`
}

// ref is how other users see the account.
func (a Account) ref() UserRef { return UserRef{ID: a.ID, Name: a.Name, Email: a.Email} }

// user is how its owner sees the account.
func (a Account) user() User {
	return User{UserRef: a.ref(), Locale: a.locale(), CreatedAt: a.CreatedAt}
}

// locale is the language the account reads: its own, or French.
func (a Account) locale() wire.Locale { return a.Locale.Resolve() }

// Accounts keeps every account, keyed by user ID. No two accounts share an
// address: the unique index refuses the second.
//
// fr: Accounts garde chaque compte, sous l’ID de son utilisateur. Deux comptes
// ne partagent jamais une adresse : l’index unique refuse le second.
var Accounts = Service.Store("accounts", func(a Account) string { return a.ID },
	kit.Unique("email", func(a Account) string { return a.Email }))

// MaxFailedLogins wrong passwords in a row lock an account.
const MaxFailedLogins = 5

// LockFor is how long a locked account stays locked.
const LockFor = 15 * time.Minute

// AccountLifecycle is the life of an account: it is verified by the link of
// its first mail, locked by its own guard after too many wrong passwords,
// unlocked by its own timer — or by a password reset, which also verifies an
// address that never was.
//
// fr: AccountLifecycle est la vie d’un compte : il est vérifié par le lien de
// son premier mail, verrouillé par sa propre garde après trop de mots de passe
// faux, déverrouillé par son propre timer — ou par une réinitialisation du mot
// de passe, qui vérifie aussi une adresse qui ne l’avait jamais été.
var AccountLifecycle = Service.Workflow("accounts", Accounts, func(a *Account) *AccountStatus { return &a.Status }).
	Initial(Unverified).
	On("verify", Unverified, Active).
	When("lock", Active, Locked, tooManyFailures).
	After("unlock", LockFor, Locked, Active).
	On("reset", Locked, Active).
	On("reset", Unverified, Active).
	OnEnter(Active, activate).
	OnEnter(Locked, stampLock).
	OnTransition(announceAccount)

// tooManyFailures holds once an account refused MaxFailedLogins passwords
// in a row. kit checks it each time the account is written.
//
// fr: tooManyFailures est vraie dès qu’un compte a refusé MaxFailedLogins mots
// de passe d’affilée. kit la vérifie à chaque écriture du compte.
func tooManyFailures(a Account) bool { return a.FailedLogins >= MaxFailedLogins }

// activate gives an account a clean slate: no failures, no lock, and the
// date its address was proved.
//
// fr: activate remet un compte à zéro : plus d’échecs, plus de verrou, et la
// date à laquelle son adresse a été prouvée.
func activate(ctx context.Context, a *Account) error {
	now := wire.Now(ctx)
	a.FailedLogins, a.LockedAt, a.UpdatedAt = 0, nil, now
	if a.VerifiedAt == nil {
		a.VerifiedAt = &now
	}
	return nil
}

// stampLock records when the account locked.
//
// fr: stampLock note quand le compte s’est verrouillé.
func stampLock(ctx context.Context, a *Account) error {
	now := wire.Now(ctx)
	a.LockedAt, a.UpdatedAt = &now, now
	return nil
}

// AccountEvent is what happened to an account. It never carries a password,
// a hash or a token.
type AccountEvent struct {
	ID     string    `json:"id"`
	Kind   string    `json:"kind"`
	UserID string    `json:"userId"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
	At     time.Time `json:"at"`
}

// The kinds of AccountEvent.
const (
	SignedUp        = "signed_up"
	Verified        = "verified"
	PasswordChanged = "password_changed"
	AccountLocked   = "locked"
)

// AccountEvents announces sign-ups, verified addresses, changed passwords and
// locked accounts to the services that care: contacts turns the invitations
// waiting for a newly verified address into requests.
//
// fr: AccountEvents annonce les inscriptions, les adresses vérifiées, les mots
// de passe changés et les comptes verrouillés aux services que cela intéresse :
// contacts change en demandes les invitations qui attendaient une adresse tout
// juste vérifiée.
var AccountEvents = Service.Topic[AccountEvent]("accounts")

// announceAccount publishes the transitions other services care about,
// whoever fired them: an endpoint, the lock guard.
//
// fr: announceAccount publie les transitions qui intéressent les autres
// services, quel que soit leur déclencheur : un endpoint, la garde de
// verrouillage.
func announceAccount(ctx context.Context, c kit.Change[Account, AccountStatus]) error {
	kind := ""
	switch {
	case c.Event == "create":
		kind = SignedUp
	case c.From == Unverified && c.To == Active:
		kind = Verified
	case c.To == Locked:
		kind = AccountLocked
	default:
		return nil
	}
	return publishAccount(ctx, c.Entity, kind)
}

// publishAccount announces kind for the account a.
func publishAccount(ctx context.Context, a Account, kind string) error {
	return AccountEvents.Publish(ctx, AccountEvent{
		ID: kit.NewID("event"), Kind: kind, UserID: a.ID, Email: a.Email, Name: a.Name, At: wire.Now(ctx),
	})
}

// accountByEmail finds the account of an address, if there is one. It reads
// the unique index with Find, so an unknown address is an answer rather than
// an error: sign-up and password recovery ask about strangers all the time.
func accountByEmail(ctx context.Context, email string) (Account, bool, error) {
	found, err := Accounts.Find(ctx, "email", email)
	if err != nil || len(found) == 0 {
		return Account{}, false, err
	}
	return found[0], true, nil
}
