// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:6d83e3d5800ec2321096deeccd3c573f533b4ecf81275cccebf1ad5e2b7143c2

package access

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/kitsunium/todo/identity/account"
)

// TestAuthenticateNamesTheMemberOfALiveSession is Authenticate's scenario "names the member of a live session" (success).
func TestAuthenticateNamesTheMemberOfALiveSession(t *testing.T) {
	uc, in, then := givenAuthenticateNamesTheMemberOfALiveSession(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("Authenticate: %v, want a success", err)
	}
	then(t, out)
}

// TestAuthenticateRefusesAnExpiredSession is Authenticate's scenario "refuses an expired session" (refusal).
func TestAuthenticateRefusesAnExpiredSession(t *testing.T) {
	uc, in := givenAuthenticateRefusesAnExpiredSession(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrNoSession) {
		t.Fatalf("Authenticate: %v, want account.ErrNoSession", err)
	}
}

// TestAuthenticateRefusesAnUnknownSecret is Authenticate's scenario "refuses an unknown secret" (refusal).
func TestAuthenticateRefusesAnUnknownSecret(t *testing.T) {
	uc, in := givenAuthenticateRefusesAnUnknownSecret(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrNoSession) {
		t.Fatalf("Authenticate: %v, want account.ErrNoSession", err)
	}
}

// TestSignInOpensASession is SignIn's scenario "opens a session" (success).
func TestSignInOpensASession(t *testing.T) {
	uc, in, then := givenSignInOpensASession(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("SignIn: %v, want a success", err)
	}
	then(t, out)
}

// TestSignInRefusesAWrongPassword is SignIn's scenario "refuses a wrong password" (refusal).
func TestSignInRefusesAWrongPassword(t *testing.T) {
	uc, in := givenSignInRefusesAWrongPassword(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrInvalidCredentials) {
		t.Fatalf("SignIn: %v, want account.ErrInvalidCredentials", err)
	}
}

// TestSignInRefusesAnUnknownAddress is SignIn's scenario "refuses an unknown address" (refusal).
func TestSignInRefusesAnUnknownAddress(t *testing.T) {
	uc, in := givenSignInRefusesAnUnknownAddress(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrInvalidCredentials) {
		t.Fatalf("SignIn: %v, want account.ErrInvalidCredentials", err)
	}
}

// TestSignInRefusesALockedAccount is SignIn's scenario "refuses a locked account" (refusal).
func TestSignInRefusesALockedAccount(t *testing.T) {
	uc, in := givenSignInRefusesALockedAccount(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrLocked) {
		t.Fatalf("SignIn: %v, want account.ErrLocked", err)
	}
}

// TestSignOutEndsTheSession is SignOut's scenario "ends the session" (success).
func TestSignOutEndsTheSession(t *testing.T) {
	uc, in, then := givenSignOutEndsTheSession(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("SignOut: %v, want a success", err)
	}
	then(t, out)
}

// TestSignUpCreatesAnAccount is SignUp's scenario "creates an account" (success).
func TestSignUpCreatesAnAccount(t *testing.T) {
	uc, in, then := givenSignUpCreatesAnAccount(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("SignUp: %v, want a success", err)
	}
	then(t, out)
}

// TestSignUpRefusesAnInvalidAddress is SignUp's scenario "refuses an invalid address" (refusal).
func TestSignUpRefusesAnInvalidAddress(t *testing.T) {
	uc, in := givenSignUpRefusesAnInvalidAddress(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrInvalidEmail) {
		t.Fatalf("SignUp: %v, want account.ErrInvalidEmail", err)
	}
}

// TestSignUpRefusesAShortPassword is SignUp's scenario "refuses a short password" (refusal).
func TestSignUpRefusesAShortPassword(t *testing.T) {
	uc, in := givenSignUpRefusesAShortPassword(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrWeakPassword) {
		t.Fatalf("SignUp: %v, want account.ErrWeakPassword", err)
	}
}

// TestSignUpRefusesATakenAddress is SignUp's scenario "refuses a taken address" (refusal).
func TestSignUpRefusesATakenAddress(t *testing.T) {
	uc, in := givenSignUpRefusesATakenAddress(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrEmailTaken) {
		t.Fatalf("SignUp: %v, want account.ErrEmailTaken", err)
	}
}

// sequenceRecorder wraps the ports a sequence goes through, and records each
// call in order.
type sequenceRecorder struct {
	calls *[]string
}

// RecordAccessAccountsV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordAccessAccountsV1(next AccountsV1) AccountsV1 {
	return recordedaccessAccountsV1{next: next, calls: r.calls}
}

type recordedaccessAccountsV1 struct {
	next  AccountsV1
	calls *[]string
}

func (x recordedaccessAccountsV1) ByEmail(ctx context.Context, email string) (account.Account, bool, error) {
	*x.calls = append(*x.calls, "identity/port/Accounts@v1.ByEmail")
	return x.next.ByEmail(ctx, email)
}

func (x recordedaccessAccountsV1) Get(ctx context.Context, id string) (account.Account, bool, error) {
	*x.calls = append(*x.calls, "identity/port/Accounts@v1.Get")
	return x.next.Get(ctx, id)
}

func (x recordedaccessAccountsV1) NewID(ctx context.Context) (string, error) {
	*x.calls = append(*x.calls, "identity/port/Accounts@v1.NewID")
	return x.next.NewID(ctx)
}

func (x recordedaccessAccountsV1) Save(ctx context.Context, a account.Account) error {
	*x.calls = append(*x.calls, "identity/port/Accounts@v1.Save")
	return x.next.Save(ctx, a)
}

// RecordAccessClockV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordAccessClockV1(next ClockV1) ClockV1 {
	return recordedaccessClockV1{next: next, calls: r.calls}
}

type recordedaccessClockV1 struct {
	next  ClockV1
	calls *[]string
}

func (x recordedaccessClockV1) Now(ctx context.Context) (time.Time, error) {
	*x.calls = append(*x.calls, "identity/port/Clock@v1.Now")
	return x.next.Now(ctx)
}

// RecordAccessPasswordsV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordAccessPasswordsV1(next PasswordsV1) PasswordsV1 {
	return recordedaccessPasswordsV1{next: next, calls: r.calls}
}

type recordedaccessPasswordsV1 struct {
	next  PasswordsV1
	calls *[]string
}

func (x recordedaccessPasswordsV1) Hash(ctx context.Context, password string) (string, error) {
	*x.calls = append(*x.calls, "identity/port/Passwords@v1.Hash")
	return x.next.Hash(ctx, password)
}

func (x recordedaccessPasswordsV1) Verify(ctx context.Context, password string, hash string) (bool, error) {
	*x.calls = append(*x.calls, "identity/port/Passwords@v1.Verify")
	return x.next.Verify(ctx, password, hash)
}

// RecordAccessSecretsV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordAccessSecretsV1(next SecretsV1) SecretsV1 {
	return recordedaccessSecretsV1{next: next, calls: r.calls}
}

type recordedaccessSecretsV1 struct {
	next  SecretsV1
	calls *[]string
}

func (x recordedaccessSecretsV1) Digest(ctx context.Context, secret string) (string, error) {
	*x.calls = append(*x.calls, "identity/port/Secrets@v1.Digest")
	return x.next.Digest(ctx, secret)
}

func (x recordedaccessSecretsV1) New(ctx context.Context) (string, string, error) {
	*x.calls = append(*x.calls, "identity/port/Secrets@v1.New")
	return x.next.New(ctx)
}

// RecordAccessSessionsV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordAccessSessionsV1(next SessionsV1) SessionsV1 {
	return recordedaccessSessionsV1{next: next, calls: r.calls}
}

type recordedaccessSessionsV1 struct {
	next  SessionsV1
	calls *[]string
}

func (x recordedaccessSessionsV1) ByTokenHash(ctx context.Context, hash string) (account.Session, bool, error) {
	*x.calls = append(*x.calls, "identity/port/Sessions@v1.ByTokenHash")
	return x.next.ByTokenHash(ctx, hash)
}

func (x recordedaccessSessionsV1) Delete(ctx context.Context, id string) error {
	*x.calls = append(*x.calls, "identity/port/Sessions@v1.Delete")
	return x.next.Delete(ctx, id)
}

func (x recordedaccessSessionsV1) NewID(ctx context.Context) (string, error) {
	*x.calls = append(*x.calls, "identity/port/Sessions@v1.NewID")
	return x.next.NewID(ctx)
}

func (x recordedaccessSessionsV1) Save(ctx context.Context, s account.Session) error {
	*x.calls = append(*x.calls, "identity/port/Sessions@v1.Save")
	return x.next.Save(ctx, s)
}

// TestSequenceSignIn is the order of sequence sign-in: the diagram's links, in the
// diagram's order (D21).
func TestSequenceSignIn(t *testing.T) {
	var got []string
	run := givenSequenceSignIn(t, sequenceRecorder{calls: &got})
	if err := run(t.Context()); err != nil {
		t.Fatalf("the sequence failed: %v", err)
	}
	want := []string{"identity/port/Accounts@v1.ByEmail", "identity/port/Clock@v1.Now", "identity/port/Passwords@v1.Verify", "identity/port/Secrets@v1.New", "identity/port/Sessions@v1.NewID", "identity/port/Sessions@v1.Save"}
	if !slices.Equal(got, want) {
		t.Errorf("the calls: %q, want %q", got, want)
	}
}
