// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package circle

import (
	"context"
	"testing"
)

// givenSequenceRequestContact builds contacts/usecase/RequestContact with the ports it goes through wrapped by rec
// (rec.Record…), and returns what starts it.
func givenSequenceRequestContact(t *testing.T, rec sequenceRecorder) func(context.Context) error {
	t.Helper()
	w := newWorld()
	uc := NewRequestContact(rec.RecordCirclePeopleV1(w), rec.RecordCircleLinksV1(w), rec.RecordCircleClockV1(w), rec.RecordCircleNoticesV1(w))
	return func(ctx context.Context) error {
		_, err := uc.Execute(ctx, RequestContactInput{From: "user_ada", Email: "bob@example.com"})
		return err
	}
}
