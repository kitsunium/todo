// Written once from the design; this file is yours. The shells in
// gen_design.go delegate to what it holds; a body that still answers
// errNotImplemented is a stub.

package reminders

import (
	"context"
	"strings"

	"github.com/kitsunium/todo/tasks/planner"
	"github.com/kitsunium/todo/tasks/task"
)

// execute is RemindDue's body: the owner of each open task due within
// Within is mailed about it once, at the address the directory gives, in the language
// and the time zone it gives. The mail leaves before the
// task is marked, so a mail that fails is tried again at the next run, and
// one that left is never sent twice unless the mark itself fails.
func (h *RemindDueHandler) execute(ctx context.Context, in *RemindDueInput) (RemindDueOutput, error) {
	due, err := h.dueSoon.Execute(ctx, planner.DueSoonInput{Within: in.Within})
	if err != nil {
		return RemindDueOutput{}, err
	}
	var out RemindDueOutput
	for _, t := range due.Tasks {
		seen, err := h.reminderLog.Seen(ctx, t.ID)
		if err != nil {
			return out, err
		}
		if seen {
			continue
		}
		to, found, err := h.directory.Recipient(ctx, t.OwnerID)
		if err != nil {
			return out, err
		}
		if !found {
			continue // the owner's account is gone: nobody to remind
		}
		if err := h.mailer.Send(ctx, reminder(t, to)); err != nil {
			return out, err
		}
		if err := h.reminderLog.Mark(ctx, t.ID); err != nil {
			return out, err
		}
		out.Sent++
	}
	return out, nil
}

// reminder is the mail a task due soon is reminded of with: the title as one
// line, a line break there would be a header; the mailer writes it in the
// recipient's language and the due date in their time zone.
func reminder(t task.Task, to Recipient) Mail {
	return Mail{To: to, Title: strings.Join(strings.Fields(t.Title), " "), Due: t.Due}
}
