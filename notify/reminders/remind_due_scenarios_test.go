// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package reminders

import (
	"testing"
	"time"
)

// givenRemindDueMailsEachTaskDueSoonOnce builds RemindDue for "mails each task due soon once": the use case, its input, and
// what its output must be.
func givenRemindDueMailsEachTaskDueSoonOnce(t *testing.T) (RemindDue, RemindDueInput, func(*testing.T, RemindDueOutput)) {
	t.Helper()
	mailer, log := &memMailer{}, memLog{"task_told": true}
	uc := NewRemindDue(dueSoon{dueTask("task_a", "ada@example.com"), dueTask("task_told", "bob@example.com")}, log, mailer)
	in := RemindDueInput{Within: 24 * time.Hour}
	return uc, in, func(t *testing.T, out RemindDueOutput) {
		t.Helper()
		if out.Sent != 1 || len(mailer.sent) != 1 {
			t.Fatalf("sent %d, the mailer got %d mail(s), want 1", out.Sent, len(mailer.sent))
		}
		if m := mailer.sent[0]; m.UserID != "user_ada" || m.To != "ada@example.com" || m.Subject != "Due soon: ship the release" {
			t.Errorf("the mail: %+v", m)
		}
		again, err := uc.Execute(t.Context(), in)
		if err != nil || again.Sent != 0 || len(mailer.sent) != 1 {
			t.Errorf("a second run sent %d more (%v), want none", again.Sent, err)
		}
	}
}

// givenRemindDueSkipsATaskAssignedToNoAddress builds RemindDue for "skips a task assigned to no address": the use case, its input, and
// what its output must be.
func givenRemindDueSkipsATaskAssignedToNoAddress(t *testing.T) (RemindDue, RemindDueInput, func(*testing.T, RemindDueOutput)) {
	t.Helper()
	mailer, log := &memMailer{}, memLog{}
	uc := NewRemindDue(dueSoon{dueTask("task_a", "Ada"), dueTask("task_b", "")}, log, mailer)
	return uc, RemindDueInput{Within: 24 * time.Hour}, func(t *testing.T, out RemindDueOutput) {
		t.Helper()
		if out.Sent != 0 || len(mailer.sent) != 0 || len(log) != 0 {
			t.Errorf("sent %d, mailed %d, marked %d: want nothing", out.Sent, len(mailer.sent), len(log))
		}
	}
}
