// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package reminders

import (
	"context"
	"testing"
	"time"
)

// givenSequenceRemindDue builds notify/usecase/RemindDue with the ports it goes through wrapped by rec
// (rec.Record…), and returns what starts it.
func givenSequenceRemindDue(t *testing.T, rec sequenceRecorder) func(context.Context) error {
	t.Helper()
	uc := NewRemindDue(
		dueSoon{dueTask("task_a", "user_ada")},
		rec.RecordRemindersReminderLogV1(memLog{}),
		rec.RecordRemindersDirectoryV1(everyone),
		rec.RecordRemindersMailerV1(&memMailer{}),
	)
	return func(ctx context.Context) error {
		_, err := uc.Execute(ctx, RemindDueInput{Within: 24 * time.Hour})
		return err
	}
}
