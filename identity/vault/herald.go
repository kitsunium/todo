// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package vault

import (
	"context"
	"time"

	"github.com/kitsunium/sdk/framework/app"
	"github.com/kitsunium/todo/identity/account"
)

// herald is Herald's own state: the topic is the package's, it holds nothing.
type herald struct {
}

// newHerald is NewHerald's body.
func newHerald() *Herald {
	return &Herald{herald{}}
}

// verified is Herald.Verified's body: the account's address, proven, is
// published on the topic "accounts", under a message id of its own.
func (a *Herald) verified(ctx context.Context, acc account.Account) (err error) {
	return AccountsTopic.Publish(ctx, Announcement{
		ID: app.NewID("account"), Kind: KindVerified, UserID: acc.ID, Email: acc.Email, At: app.Now(ctx).UTC().Truncate(time.Millisecond),
	})
}
