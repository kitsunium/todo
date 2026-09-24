package identity

import (
	"context"

	"github.com/kitsunium/platform/kit"
	"github.com/kitsunium/todo/internal/wire"
)

// The directory: how the other services learn who a user is. They never read
// the accounts store; they call these, in-process.
var (
	// UserByEmailAPI finds the user who owns an address — contacts asks it
	// whether to send a contact request or an invitation.
	UserByEmailAPI = Service.Endpoint("GET /internal/users/by-email", UserByEmail, kit.Private())

	// UsersAPI turns user IDs into the names and addresses other users see,
	// and the language each reads — notify writes to them in it.
	UsersAPI = Service.Endpoint("POST /internal/users/batch", Users, kit.Private())
)

// EmailQuery is an address to look up.
type EmailQuery struct {
	Email string `query:"email"`
}

// UserByEmailOutput is the user who owns the address, if one does.
type UserByEmailOutput struct {
	User *UserRef `json:"user,omitempty"`
}

// UserByEmail returns the user who owns an address. Only a verified address
// counts: an account nobody confirmed is nobody yet, so an invitation to it
// waits for the verification — and then becomes a request.
func UserByEmail(ctx context.Context, in EmailQuery) (UserByEmailOutput, error) {
	a, found, err := accountByEmail(ctx, NormalizeEmail(in.Email))
	if err != nil || !found || a.Status == Unverified {
		return UserByEmailOutput{}, err
	}
	ref := a.ref()
	return UserByEmailOutput{User: &ref}, nil
}

// IDs are user IDs.
type IDs struct {
	IDs []string `json:"ids"`
}

// Person is a user as the product's services see them: what other users
// see, and the language they read. It stays behind the private endpoints —
// no user learns another's language.
type Person struct {
	UserRef
	// Locale is "fr" or "en"; an account that never chose reads French.
	Locale wire.Locale `json:"locale"`
}

// UsersOutput are the users found, in the order asked; an unknown ID is
// left out.
type UsersOutput struct {
	Users []Person `json:"users"`
}

// Users returns the users with the given IDs.
func Users(ctx context.Context, in IDs) (UsersOutput, error) {
	out := UsersOutput{Users: []Person{}}
	seen := map[string]bool{}
	for _, id := range in.IDs {
		if id == "" || seen[id] {
			continue
		}
		seen[id] = true
		a, err := Accounts.Get(ctx, id)
		if wire.Is(err, kit.CodeNotFound) {
			continue
		}
		if err != nil {
			return UsersOutput{}, err
		}
		out.Users = append(out.Users, Person{UserRef: a.ref(), Locale: a.locale()})
	}
	return out, nil
}

// People resolves user IDs through UsersAPI into a map, for the services
// that write to users: an unknown ID is left out — nobody to write to.
func People(ctx context.Context, ids ...string) (map[string]Person, error) {
	out := map[string]Person{}
	if len(ids) == 0 {
		return out, nil
	}
	found, err := UsersAPI.Call(ctx, IDs{IDs: ids})
	if err != nil {
		return nil, err
	}
	for _, p := range found.Users {
		out[p.ID] = p
	}
	return out, nil
}

// Directory resolves user IDs through UsersAPI into a map, for the services
// that show users: an unknown ID maps to a placeholder rather than failing a
// whole list.
func Directory(ctx context.Context, ids ...string) (map[string]UserRef, error) {
	people, err := People(ctx, ids...)
	if err != nil {
		return nil, err
	}
	out := make(map[string]UserRef, len(people))
	for id, p := range people {
		out[id] = p.UserRef
	}
	for _, id := range ids {
		if _, ok := out[id]; !ok && id != "" {
			out[id] = UserRef{ID: id, Name: "Former user"}
		}
	}
	return out, nil
}
