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

	// UsersAPI turns user IDs into the names and addresses other users see.
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

// UsersOutput are the users found, in the order asked; an unknown ID is
// left out.
type UsersOutput struct {
	Users []UserRef `json:"users"`
}

// Users returns the users with the given IDs.
func Users(ctx context.Context, in IDs) (UsersOutput, error) {
	out := UsersOutput{Users: []UserRef{}}
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
		out.Users = append(out.Users, a.ref())
	}
	return out, nil
}

// Directory resolves user IDs through UsersAPI into a map, for the services
// that show users: an unknown ID maps to a placeholder rather than failing a
// whole list.
func Directory(ctx context.Context, ids ...string) (map[string]UserRef, error) {
	out := map[string]UserRef{}
	if len(ids) == 0 {
		return out, nil
	}
	found, err := UsersAPI.Call(ctx, IDs{IDs: ids})
	if err != nil {
		return nil, err
	}
	for _, u := range found.Users {
		out[u.ID] = u
	}
	for _, id := range ids {
		if _, ok := out[id]; !ok && id != "" {
			out[id] = UserRef{ID: id, Name: "Former user"}
		}
	}
	return out, nil
}
