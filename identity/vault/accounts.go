// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"
	"errors"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/identity/account"
)

// accounts is Accounts's own state: the store is the package's, it holds nothing.
type accounts struct {
}

// newAccounts is NewAccounts's body.
func newAccounts() *Accounts {
	return &Accounts{accounts{}}
}

// byEmail is Accounts.ByEmail's body: the account of an address, through
// the store's index, and whether there is one.
func (a *Accounts) byEmail(ctx context.Context, email string) (r0 account.Account, r1 bool, err error) {
	found, err := AccountsStore.Find(ctx, "by-email", email)
	if err != nil || len(found) == 0 {
		return account.Account{}, false, err
	}
	return accountOf(found[0]), true, nil
}

// get is Accounts.Get's body. It filters instead of reading the key: an
// account that is gone is an answer, never a failed read.
func (a *Accounts) get(ctx context.Context, id string) (r0 account.Account, r1 bool, err error) {
	found, err := AccountsStore.Filter(ctx, func(r AccountRecord) bool { return r.ID == id })
	if err != nil || len(found) == 0 {
		return account.Account{}, false, err
	}
	return accountOf(found[0]), true, nil
}

// newID is Accounts.NewID's body: a time-ordered id.
func (a *Accounts) newID(ctx context.Context) (r0 string, err error) {
	return app.NewID("user"), nil
}

// save is Accounts.Save's body: the account replaces what its id held.
func (a *Accounts) save(ctx context.Context, acc account.Account) (err error) {
	err = AccountsStore.Put(ctx, recordOf(acc))
	var refused *app.Error
	if errors.As(err, &refused) && refused.Code == app.WireConflict {
		return account.ErrEmailTaken // another account took the address since it was looked up
	}
	return err
}

// recordOf is an account as the store keeps it.
func recordOf(acc account.Account) AccountRecord {
	return AccountRecord{
		ID: acc.ID, Email: acc.Email, Name: acc.Name, Locale: acc.Locale, TimeZone: acc.TimeZone, PasswordHash: acc.PasswordHash, Verified: acc.Verified,
		FailedLogins: acc.FailedLogins, LockedUntil: acc.LockedUntil, CreatedAt: acc.CreatedAt, UpdatedAt: acc.UpdatedAt,
	}
}

// accountOf is the account a record keeps.
func accountOf(r AccountRecord) account.Account {
	return account.Account{
		ID: r.ID, Email: r.Email, Name: r.Name, Locale: r.Locale, TimeZone: r.TimeZone, PasswordHash: r.PasswordHash, Verified: r.Verified,
		FailedLogins: r.FailedLogins, LockedUntil: r.LockedUntil, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

// change is Accounts.Change's body: the store applies the change to the record it
// holds at that instant, alone among its writers, so two changes made at
// once both stay. An account that is gone is ErrInvalidCredentials: nobody signs in to it. The change's own error keeps nothing.
func (a *Accounts) change(ctx context.Context, id string, change func(*account.Account) error) (r0 account.Account, err error) {
	rec, err := AccountsStore.Update(ctx, id, func(r *AccountRecord) error {
		v := accountOf(*r)
		if err := change(&v); err != nil {
			return err
		}
		v.ID = r.ID // a change never renames a record
		*r = recordOf(v)
		return nil
	})
	var refused *app.Error
	switch {
	case errors.As(err, &refused) && refused.Code == app.WireNotFound:
		return account.Account{}, account.ErrInvalidCredentials
	case err != nil:
		return account.Account{}, err
	}
	return accountOf(rec), nil
}
