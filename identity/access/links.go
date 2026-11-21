// What the use cases of the one-time links share: issuing one, redeeming
// one, and the rule a password follows.

package access

import (
	"context"
	"time"
	"unicode/utf8"

	"github.com/kitsunium/todo/identity/account"
)

// maxPassword is the longest password an account takes: stretching one
// costs time in proportion to nothing a caller should choose.
const maxPassword = 128

// acceptable reports whether a password has the length an account takes.
func acceptable(password string) bool {
	n := utf8.RuneCountInString(password)
	return n >= account.MinPassword && n <= maxPassword
}

// issue makes the member's one link of a purpose, good for valid from now,
// in place of any they had, and returns its secret, which only the mail
// will carry.
func issue(ctx context.Context, links LinksV1, secrets SecretsV1, userID, purpose string, now time.Time, valid time.Duration) (string, error) {
	if err := links.DeleteOf(ctx, userID, purpose); err != nil {
		return "", err
	}
	secret, digest, err := secrets.New(ctx)
	if err != nil {
		return "", err
	}
	id, err := links.NewID(ctx)
	if err != nil {
		return "", err
	}
	l := account.Link{ID: id, Hash: digest, Purpose: purpose, UserID: userID, CreatedAt: now, ExpiresAt: now.Add(valid)}
	return secret, links.Save(ctx, l)
}

// redeem is the link a secret opens for a purpose, while it lasts. A link
// that is unknown, expired or meant for something else is ErrBadLink, the
// same for all three, so a link reveals nothing.
func redeem(ctx context.Context, links LinksV1, secrets SecretsV1, clock ClockV1, secret, purpose string) (account.Link, time.Time, error) {
	digest, err := secrets.Digest(ctx, secret)
	if err != nil {
		return account.Link{}, time.Time{}, err
	}
	l, found, err := links.ByHash(ctx, digest)
	if err != nil {
		return account.Link{}, time.Time{}, err
	}
	now, err := clock.Now(ctx)
	if err != nil {
		return account.Link{}, time.Time{}, err
	}
	if !found || l.Purpose != purpose || !now.Before(l.ExpiresAt) {
		return account.Link{}, now, account.ErrBadLink
	}
	return l, now, nil
}
