// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:b86d0cd87931d5d99bbe4360b3b902a3632eea398c69f6b97c0556cdb2a00c3c

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

// TestChangePasswordReplacesThePassword is ChangePassword's scenario "replaces the password" (success).
func TestChangePasswordReplacesThePassword(t *testing.T) {
	uc, in, then := givenChangePasswordReplacesThePassword(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ChangePassword: %v, want a success", err)
	}
	then(t, out)
}

// TestChangePasswordRefusesAWrongCurrentPassword is ChangePassword's scenario "refuses a wrong current password" (refusal).
func TestChangePasswordRefusesAWrongCurrentPassword(t *testing.T) {
	uc, in := givenChangePasswordRefusesAWrongCurrentPassword(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrInvalidCredentials) {
		t.Fatalf("ChangePassword: %v, want account.ErrInvalidCredentials", err)
	}
}

// TestForgotPasswordMailsTheLinkToAMember is ForgotPassword's scenario "mails the link to a member" (success).
func TestForgotPasswordMailsTheLinkToAMember(t *testing.T) {
	uc, in, then := givenForgotPasswordMailsTheLinkToAMember(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ForgotPassword: %v, want a success", err)
	}
	then(t, out)
}

// TestForgotPasswordSaysNothingOfAnUnknownAddress is ForgotPassword's scenario "says nothing of an unknown address" (success).
func TestForgotPasswordSaysNothingOfAnUnknownAddress(t *testing.T) {
	uc, in, then := givenForgotPasswordSaysNothingOfAnUnknownAddress(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ForgotPassword: %v, want a success", err)
	}
	then(t, out)
}

// TestResendVerificationMailsANewLinkToAnUnverifiedAccount is ResendVerification's scenario "mails a new link to an unverified account" (success).
func TestResendVerificationMailsANewLinkToAnUnverifiedAccount(t *testing.T) {
	uc, in, then := givenResendVerificationMailsANewLinkToAnUnverifiedAccount(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ResendVerification: %v, want a success", err)
	}
	then(t, out)
}

// TestResendVerificationSaysNothingOfAnUnknownAddress is ResendVerification's scenario "says nothing of an unknown address" (success).
func TestResendVerificationSaysNothingOfAnUnknownAddress(t *testing.T) {
	uc, in, then := givenResendVerificationSaysNothingOfAnUnknownAddress(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ResendVerification: %v, want a success", err)
	}
	then(t, out)
}

// TestResetPasswordChoosesANewPassword is ResetPassword's scenario "chooses a new password" (success).
func TestResetPasswordChoosesANewPassword(t *testing.T) {
	uc, in, then := givenResetPasswordChoosesANewPassword(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("ResetPassword: %v, want a success", err)
	}
	then(t, out)
}

// TestResetPasswordRefusesALinkUsedOnce is ResetPassword's scenario "refuses a link used once" (refusal).
func TestResetPasswordRefusesALinkUsedOnce(t *testing.T) {
	uc, in := givenResetPasswordRefusesALinkUsedOnce(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrBadLink) {
		t.Fatalf("ResetPassword: %v, want account.ErrBadLink", err)
	}
}

// TestResetPasswordRefusesAShortPassword is ResetPassword's scenario "refuses a short password" (refusal).
func TestResetPasswordRefusesAShortPassword(t *testing.T) {
	uc, in := givenResetPasswordRefusesAShortPassword(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrWeakPassword) {
		t.Fatalf("ResetPassword: %v, want account.ErrWeakPassword", err)
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

// TestSignInRefusesAnUnverifiedAccount is SignIn's scenario "refuses an unverified account" (refusal).
func TestSignInRefusesAnUnverifiedAccount(t *testing.T) {
	uc, in := givenSignInRefusesAnUnverifiedAccount(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrUnverified) {
		t.Fatalf("SignIn: %v, want account.ErrUnverified", err)
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

// TestSignUpCreatesAnAccountToVerify is SignUp's scenario "creates an account to verify" (success).
func TestSignUpCreatesAnAccountToVerify(t *testing.T) {
	uc, in, then := givenSignUpCreatesAnAccountToVerify(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("SignUp: %v, want a success", err)
	}
	then(t, out)
}

// TestSignUpCreatesAVerifiedAccountWhereAddressesAreTrusted is SignUp's scenario "creates a verified account where addresses are trusted" (success).
func TestSignUpCreatesAVerifiedAccountWhereAddressesAreTrusted(t *testing.T) {
	uc, in, then := givenSignUpCreatesAVerifiedAccountWhereAddressesAreTrusted(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("SignUp: %v, want a success", err)
	}
	then(t, out)
}

// TestSignUpAnswersATakenAddressAlike is SignUp's scenario "answers a taken address alike" (success).
func TestSignUpAnswersATakenAddressAlike(t *testing.T) {
	uc, in, then := givenSignUpAnswersATakenAddressAlike(t)
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

// TestVerifyEmailVerifiesTheAccountOfALink is VerifyEmail's scenario "verifies the account of a link" (success).
func TestVerifyEmailVerifiesTheAccountOfALink(t *testing.T) {
	uc, in, then := givenVerifyEmailVerifiesTheAccountOfALink(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("VerifyEmail: %v, want a success", err)
	}
	then(t, out)
}

// TestVerifyEmailRefusesAnExpiredLink is VerifyEmail's scenario "refuses an expired link" (refusal).
func TestVerifyEmailRefusesAnExpiredLink(t *testing.T) {
	uc, in := givenVerifyEmailRefusesAnExpiredLink(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrBadLink) {
		t.Fatalf("VerifyEmail: %v, want account.ErrBadLink", err)
	}
}

// TestVerifyEmailRefusesALinkOfAnotherPurpose is VerifyEmail's scenario "refuses a link of another purpose" (refusal).
func TestVerifyEmailRefusesALinkOfAnotherPurpose(t *testing.T) {
	uc, in := givenVerifyEmailRefusesALinkOfAnotherPurpose(t)
	_, err := uc.Execute(t.Context(), in)
	if !errors.Is(err, account.ErrBadLink) {
		t.Fatalf("VerifyEmail: %v, want account.ErrBadLink", err)
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

func (x recordedaccessSessionsV1) DeleteOf(ctx context.Context, userID string, keep string) error {
	*x.calls = append(*x.calls, "identity/port/Sessions@v1.DeleteOf")
	return x.next.DeleteOf(ctx, userID, keep)
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
