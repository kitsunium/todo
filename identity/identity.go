// Package identity owns who the users are: their accounts and passwords, the
// sessions that sign them in, and the one-time links that verify an address
// or reset a password. Every other service knows a user by ID and asks this
// one, through its private endpoints, for the rest.
package identity

import (
	"strings"
	"time"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/sdk/pkg/v1/mail"
	"github.com/kitsunium/todo/internal/wire"
)

// Service owns accounts, passwords and sessions.
var Service = kit.NewService("identity", "Accounts, passwords, sessions, and the one-time links that verify an address or reset a password.")

// UserRef is how the product shows a user to another user.
type UserRef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// User is a user as they see their own account.
type User struct {
	UserRef
	// Locale is the language the user reads the product and its mails in:
	// "fr" or "en".
	Locale    wire.Locale `json:"locale"`
	CreatedAt time.Time   `json:"createdAt"`
}

// UserOutput is the answer of every endpoint that returns the caller: {user}.
type UserOutput struct {
	User User `json:"user"`
}

// NormalizeEmail returns the form an address is stored and compared in:
// trimmed and lower-cased.
func NormalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ValidEmail reports whether addr, normalized, is an address the product can
// write to: one '@', a domain with a dot, and whatever the SDK's mail rules
// accept as a recipient — so an account is never opened for an address the
// outbox would refuse.
func ValidEmail(addr string) bool {
	local, domain, ok := strings.Cut(addr, "@")
	if !ok || local == "" || len(addr) > 254 ||
		!strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}
	probe := mail.Address{Addr: addr}
	return mail.Validate(mail.Message{From: probe, To: []mail.Address{probe}, Text: "."}) == nil
}
