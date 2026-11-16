// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package access

import (
	"context"
	"testing"
)

// givenSequenceSignIn builds identity/usecase/SignIn with the ports it goes through wrapped by rec
// (rec.Record…), and returns what starts it.
func givenSequenceSignIn(t *testing.T, rec sequenceRecorder) func(context.Context) error {
	t.Helper()
	w := newWorld()
	w.member("ada@example.com")
	uc := NewSignIn(
		rec.RecordAccessAccountsV1(accountsPort{w}),
		rec.RecordAccessClockV1(clockPort{w}),
		rec.RecordAccessPasswordsV1(passwordsPort{w}),
		rec.RecordAccessSecretsV1(secretsPort{w}),
		rec.RecordAccessSessionsV1(sessionsPort{w}),
	)
	return func(ctx context.Context) error {
		_, err := uc.Execute(ctx, SignInInput{Email: "ada@example.com", Password: passphrase})
		return err
	}
}
