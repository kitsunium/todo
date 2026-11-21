// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package reminders

import (
	"testing"
)

// action is what the five kinds of event are told with: Ada did it to Bob.
func action(kind string) MailActionInput {
	return MailActionInput{EventID: "event_" + kind, Kind: kind, ActorID: "user_ada", UserID: "user_bob", Subject: "the release", SubjectID: "task_a"}
}

// givenMailActionMailsEachKindToItsMember builds MailAction for "mails each kind to its member": the use case, its input, and
// what its output must be.
func givenMailActionMailsEachKindToItsMember(t *testing.T) (MailAction, MailActionInput, func(*testing.T, MailActionOutput)) {
	t.Helper()
	notifier := &memNotifier{}
	uc := NewMailAction(memLog{}, everyone, notifier)
	kinds := []string{OnContactRequested, OnContactAccepted, OnGroupInvited, OnShared, OnAssigned}
	return uc, action(OnContactRequested), func(t *testing.T, out MailActionOutput) {
		t.Helper()
		for _, kind := range kinds[1:] {
			if _, err := uc.Execute(t.Context(), action(kind)); err != nil {
				t.Fatalf("%s: %v", kind, err)
			}
		}
		if out.Sent != 1 || len(notifier.sent) != len(kinds) {
			t.Fatalf("sent %d, the notifier got %d notice(s), want %d", out.Sent, len(notifier.sent), len(kinds))
		}
		for i, n := range notifier.sent {
			want := Notice{To: everyone["user_bob"], Kind: kinds[i], Actor: "Ada", Subject: "the release", SubjectID: "task_a"}
			if n != want {
				t.Errorf("notice %d: %+v, want %+v", i, n, want)
			}
		}
		if other, err := uc.Execute(t.Context(), action("edited")); err != nil || other.Sent != 0 || len(notifier.sent) != len(kinds) {
			t.Errorf("a kind nobody is mailed of sent %d (%v), want none", other.Sent, err)
		}
	}
}

// givenMailActionMailsADeliveryOnce builds MailAction for "mails a delivery once": the use case, its input, and
// what its output must be.
func givenMailActionMailsADeliveryOnce(t *testing.T) (MailAction, MailActionInput, func(*testing.T, MailActionOutput)) {
	t.Helper()
	notifier := &memNotifier{}
	uc := NewMailAction(memLog{}, everyone, notifier)
	in := action(OnShared)
	return uc, in, func(t *testing.T, out MailActionOutput) {
		t.Helper()
		again, err := uc.Execute(t.Context(), in)
		if err != nil || out.Sent != 1 || again.Sent != 0 || len(notifier.sent) != 1 {
			t.Errorf("sent %d then %d (%v), the notifier got %d: want one mail", out.Sent, again.Sent, err, len(notifier.sent))
		}
		other := in
		other.EventID = "event_another"
		if next, err := uc.Execute(t.Context(), other); err != nil || next.Sent != 1 || len(notifier.sent) != 2 {
			t.Errorf("another event sent %d (%v), the notifier got %d: want a second mail", next.Sent, err, len(notifier.sent))
		}
	}
}

// givenMailActionMailsNobodyOfWhatAMemberDidToThemselves builds MailAction for "mails nobody of what a member did to themselves": the use case, its input, and
// what its output must be.
func givenMailActionMailsNobodyOfWhatAMemberDidToThemselves(t *testing.T) (MailAction, MailActionInput, func(*testing.T, MailActionOutput)) {
	t.Helper()
	notifier, log := &memNotifier{}, memLog{}
	in := action(OnAssigned)
	in.UserID = in.ActorID
	return NewMailAction(log, everyone, notifier), in, func(t *testing.T, out MailActionOutput) {
		t.Helper()
		if out.Sent != 0 || len(notifier.sent) != 0 || len(log) != 0 {
			t.Errorf("sent %d, mailed %d, marked %d: want nothing", out.Sent, len(notifier.sent), len(log))
		}
	}
}

// givenMailActionMailsNobodyTheDirectoryDoesNotKnow builds MailAction for "mails nobody the directory does not know": the use case, its input, and
// what its output must be.
func givenMailActionMailsNobodyTheDirectoryDoesNotKnow(t *testing.T) (MailAction, MailActionInput, func(*testing.T, MailActionOutput)) {
	t.Helper()
	notifier, log := &memNotifier{}, memLog{}
	uc := NewMailAction(log, everyone, notifier)
	in := action(OnShared)
	in.UserID = "user_gone"
	return uc, in, func(t *testing.T, out MailActionOutput) {
		t.Helper()
		stranger := action(OnShared)
		stranger.ActorID = "user_gone"
		again, err := uc.Execute(t.Context(), stranger)
		if out.Sent != 0 || again.Sent != 0 || err != nil || len(notifier.sent) != 0 || len(log) != 0 {
			t.Errorf("sent %d and %d (%v), mailed %d, marked %d: want nothing", out.Sent, again.Sent, err, len(notifier.sent), len(log))
		}
	}
}
