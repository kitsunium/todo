// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package reminders

import (
	"context"
	"testing"
)

// givenSequenceMailAction builds notify/usecase/MailAction with the ports it goes through wrapped by rec
// (rec.Record…), and returns what starts it.
func givenSequenceMailAction(t *testing.T, rec sequenceRecorder) func(context.Context) error {
	t.Helper()
	uc := NewMailAction(
		rec.RecordRemindersActionLogV1(memLog{}),
		rec.RecordRemindersDirectoryV1(everyone),
		rec.RecordRemindersNotifierV1(&memNotifier{}),
	)
	return func(ctx context.Context) error {
		// Ada shared a task with Bob: the log is asked, the directory names
		// the actor then the member, the mail leaves, the event is marked.
		_, err := uc.Execute(ctx, action(OnShared))
		return err
	}
}
