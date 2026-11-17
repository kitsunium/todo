// Code generated from design/v1; DO NOT EDIT.
// design:digest sha256:390717aeedfce4c9881bfea93cf567e5dfb6e0f4b99777661b165bb23d396f17

package reminders

import (
	"context"
	"slices"
	"testing"
)

// TestMailActionMailsEachKindToItsMember is MailAction's scenario "mails each kind to its member" (success).
func TestMailActionMailsEachKindToItsMember(t *testing.T) {
	uc, in, then := givenMailActionMailsEachKindToItsMember(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("MailAction: %v, want a success", err)
	}
	then(t, out)
}

// TestMailActionMailsADeliveryOnce is MailAction's scenario "mails a delivery once" (success).
func TestMailActionMailsADeliveryOnce(t *testing.T) {
	uc, in, then := givenMailActionMailsADeliveryOnce(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("MailAction: %v, want a success", err)
	}
	then(t, out)
}

// TestMailActionMailsNobodyOfWhatAMemberDidToThemselves is MailAction's scenario "mails nobody of what a member did to themselves" (success).
func TestMailActionMailsNobodyOfWhatAMemberDidToThemselves(t *testing.T) {
	uc, in, then := givenMailActionMailsNobodyOfWhatAMemberDidToThemselves(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("MailAction: %v, want a success", err)
	}
	then(t, out)
}

// TestMailActionMailsNobodyTheDirectoryDoesNotKnow is MailAction's scenario "mails nobody the directory does not know" (success).
func TestMailActionMailsNobodyTheDirectoryDoesNotKnow(t *testing.T) {
	uc, in, then := givenMailActionMailsNobodyTheDirectoryDoesNotKnow(t)
	out, err := uc.Execute(t.Context(), in)
	if err != nil {
		t.Fatalf("MailAction: %v, want a success", err)
	}
	then(t, out)
}

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

// RecordRemindersActionLogV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRemindersActionLogV1(next ActionLogV1) ActionLogV1 {
	return recordedremindersActionLogV1{next: next, calls: r.calls}
}

type recordedremindersActionLogV1 struct {
	next  ActionLogV1
	calls *[]string
}

func (x recordedremindersActionLogV1) Mark(ctx context.Context, key string) error {
	*x.calls = append(*x.calls, "notify/port/ActionLog@v1.Mark")
	return x.next.Mark(ctx, key)
}

func (x recordedremindersActionLogV1) Seen(ctx context.Context, key string) (bool, error) {
	*x.calls = append(*x.calls, "notify/port/ActionLog@v1.Seen")
	return x.next.Seen(ctx, key)
}

// RecordRemindersDirectoryV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRemindersDirectoryV1(next DirectoryV1) DirectoryV1 {
	return recordedremindersDirectoryV1{next: next, calls: r.calls}
}

type recordedremindersDirectoryV1 struct {
	next  DirectoryV1
	calls *[]string
}

func (x recordedremindersDirectoryV1) Recipient(ctx context.Context, userID string) (Recipient, bool, error) {
	*x.calls = append(*x.calls, "notify/port/Directory@v1.Recipient")
	return x.next.Recipient(ctx, userID)
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

// RecordRemindersNotifierV1 records the calls to next, then makes them.
func (r sequenceRecorder) RecordRemindersNotifierV1(next NotifierV1) NotifierV1 {
	return recordedremindersNotifierV1{next: next, calls: r.calls}
}

type recordedremindersNotifierV1 struct {
	next  NotifierV1
	calls *[]string
}

func (x recordedremindersNotifierV1) Notify(ctx context.Context, n Notice) error {
	*x.calls = append(*x.calls, "notify/port/Notifier@v1.Notify")
	return x.next.Notify(ctx, n)
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

// TestSequenceMailAction is the order of sequence mail-action: the diagram's links, in the
// diagram's order (D21).
func TestSequenceMailAction(t *testing.T) {
	var got []string
	run := givenSequenceMailAction(t, sequenceRecorder{calls: &got})
	if err := run(t.Context()); err != nil {
		t.Fatalf("the sequence failed: %v", err)
	}
	want := []string{"notify/port/ActionLog@v1.Seen", "notify/port/Directory@v1.Recipient", "notify/port/Directory@v1.Recipient", "notify/port/Notifier@v1.Notify", "notify/port/ActionLog@v1.Mark"}
	if !slices.Equal(got, want) {
		t.Errorf("the calls: %q, want %q", got, want)
	}
}

// TestSequenceRemindDue is the order of sequence remind-due: the diagram's links, in the
// diagram's order (D21).
func TestSequenceRemindDue(t *testing.T) {
	var got []string
	run := givenSequenceRemindDue(t, sequenceRecorder{calls: &got})
	if err := run(t.Context()); err != nil {
		t.Fatalf("the sequence failed: %v", err)
	}
	want := []string{"notify/port/ReminderLog@v1.Seen", "notify/port/Directory@v1.Recipient", "notify/port/Mailer@v1.Send", "notify/port/ReminderLog@v1.Mark"}
	if !slices.Equal(got, want) {
		t.Errorf("the calls: %q, want %q", got, want)
	}
}
