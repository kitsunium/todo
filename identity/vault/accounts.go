// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"

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
	return AccountsStore.Put(ctx, AccountRecord{
		ID: acc.ID, Email: acc.Email, Name: acc.Name, PasswordHash: acc.PasswordHash,
		FailedLogins: acc.FailedLogins, LockedUntil: acc.LockedUntil, CreatedAt: acc.CreatedAt, UpdatedAt: acc.UpdatedAt,
	})
}

// accountOf is the account a record keeps.
func accountOf(r AccountRecord) account.Account {
	return account.Account{
		ID: r.ID, Email: r.Email, Name: r.Name, PasswordHash: r.PasswordHash,
		FailedLogins: r.FailedLogins, LockedUntil: r.LockedUntil, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}
