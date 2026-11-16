// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:0c616a2ddc211ada253896609000dcaab6758455e2991084b86e140dd79690fe

package reminders

import (
	"context"
	"slices"
	"testing"
)

// TestRemindDueMailsEachTaskDueSoonOnce is RemindDue's scenario "mails each task due soon once" (success).
func TestRemindDueMailsEachTaskDueSoonOnce(t *testing.T) {
	uc, in, then := givenRemindDueMailsEachTaskDueSoonOnce(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("RemindDue: %v, want a success", err)
	}
	then(t, out)
}

// TestRemindDueSkipsATaskWhoseOwnerIsGone is RemindDue's scenario "skips a task whose owner is gone" (success).
func TestRemindDueSkipsATaskWhoseOwnerIsGone(t *testing.T) {
	uc, in, then := givenRemindDueSkipsATaskWhoseOwnerIsGone(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("RemindDue: %v, want a success", err)
	}
	then(t, out)
}

// sequenceRecorder wraps the ports a sequence goes through, and records each
// call in order.
type sequenceRecorder struct {
	calls *[]string
}

// RecordRemindersDirectoryV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRemindersDirectoryV1(next DirectoryV1) DirectoryV1 {
	return recordedremindersDirectoryV1{next: next, calls: r.calls}
}

type recordedremindersDirectoryV1 struct {
	next  DirectoryV1
	calls *[]string
}

func (x recordedremindersDirectoryV1) Address(ctx context.Context, userID string) (string, bool, error) {
	*x.calls = append(*x.calls, "notify/port/Directory@v1.Address")
	return x.next.Address(ctx, userID)
}

// RecordRemindersMailerV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRemindersMailerV1(next MailerV1) MailerV1 {
	return recordedremindersMailerV1{next: next, calls: r.calls}
}

type recordedremindersMailerV1 struct {
	next  MailerV1
	calls *[]string
}

func (x recordedremindersMailerV1) Send(ctx context.Context, m Mail) error {
	*x.calls = append(*x.calls, "notify/port/Mailer@v1.Send")
	return x.next.Send(ctx, m)
}

// RecordRemindersReminderLogV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRemindersReminderLogV1(next ReminderLogV1) ReminderLogV1 {
	return recordedremindersReminderLogV1{next: next, calls: r.calls}
}

type recordedremindersReminderLogV1 struct {
	next  ReminderLogV1
	calls *[]string
}

func (x recordedremindersReminderLogV1) Mark(ctx context.Context, taskID string) error {
	*x.calls = append(*x.calls, "notify/port/ReminderLog@v1.Mark")
	return x.next.Mark(ctx, taskID)
}

func (x recordedremindersReminderLogV1) Seen(ctx context.Context, taskID string) (bool, error) {
	*x.calls = append(*x.calls, "notify/port/ReminderLog@v1.Seen")
	return x.next.Seen(ctx, taskID)
}

// TestSequenceRemindDue is the order of sequence remind-due: the diagram's links, in the
// diagram's order (D21).
func TestSequenceRemindDue(t *testing.T) {
	var got []string
	run := givenSequenceRemindDue(t, sequenceRecorder{calls: &got})
	if err := run(t.Context()); err != nil {
		t.Fatalf("the sequence failed: %v", err)
	}
	want := []string{"notify/port/ReminderLog@v1.Seen", "notify/port/Directory@v1.Address", "notify/port/Mailer@v1.Send", "notify/port/ReminderLog@v1.Mark"}
	if !slices.Equal(got, want) {
		t.Errorf("the calls: %q, want %q", got, want)
	}
}
